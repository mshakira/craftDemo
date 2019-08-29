package main

import (
	"context"
	"craftDemoClient/format/tableFormat"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const (
	TIMEOUT       = 2 // no of secs for url timeout
	RETRY         = 5 // no of retries for url failures
	NumGoRoutines = 10 // maximum number of go routines to fan out
)

// Incidents Json structure
type Incidents struct {
	Name   string     `json:"Name"`
	Report []Incident `json:"Report"`
}

// Individual inc structure
type Incident struct {
	Number      string `json:"number"`
	AssignedTo  string `json:"assigned_to"`
	Description string `json:"description"`
	State       string `json:"state"`
	Priority    string `json:"priority"`
	Severity    string `json:"severity"`
}

// Aggregated report structure based on priority
type PrioritySum struct {
	Priority string
	Sum      int
}

// To verify if go routine leaks are there
func countGoRoutines() int {
	return runtime.NumGoroutine()
}

// Initialize httpclient and request the given url
// Retry 5 times, while connecting to the server incase of error
func GetResponse(url string) (res *http.Response) {
	// InsecureSkipVerify to false for production
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Transport: tr,
		Timeout:   time.Second * TIMEOUT, // Maximum of 2 secs
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	retry := RETRY
	for i := 1; i <= retry; i++ {
		var getErr error
		// make the request
		res, getErr = client.Do(req)
		if getErr != nil {
			fmt.Printf("Attempt %d %v\n", i, getErr)
			if i > retry {
				log.Fatal(getErr)
			}
		} else {
			retry = 0
		}
	}
	return res
}

// Validate the response based on response header
func ValidateResponse(res *http.Response) (err error) {
	var resLength int
	// non 200 errors
	if res.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("Received %d status code\n", res.StatusCode))
	} else if res.Header["Content-Type"][0] != "application/json" {
		err = errors.New(fmt.Sprintf("Content type not spplication/json. Received => %s\n", res.Header["Content-Type"][0]))
	} else {
		if len(res.Header["Content-Length"]) > 0 {
			resLength, err = strconv.Atoi(res.Header["Content-Length"][0])
			if err == nil && resLength != 905 {
				err = errors.New(fmt.Sprintf("content-Length mismatch 905 vs %d\n", resLength))
			}
		}
	}
	return err
}

// walkIncs will walk through slice of incidents and sends required priority details
// to outbound channel. Once slice values are exhausted, close the output channel
// If done signal received, return early
func walkIncs(ctx context.Context, report []Incident) (chan map[string]int, error) {
	out := make(chan map[string]int)
	go func() {
		defer close(out)
		for _, obj := range report {
			m := make(map[string]int)
			m[obj.Priority] = 1
			select {
			case out <- m:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

// Merges the inputs based on priority
// Since input and outbound channels are same, we can compose this any number of times
// No need to close out channel because it is used by multiple go routines. Main has to close it
// For the same reason, we do not need context or done channels
func mergeIncs(ctx context.Context, incs chan map[string]int, out chan map[string]int) {
	sev := make(map[string]int)
	for obj := range incs {
		for k, v := range obj {
			// aggregate same key objects
			if _, ok := sev[k]; ok {
				sev[k] += v
			} else {
				sev[k] = v
			}
		}
	}
	// send aggregated value
	select {
		case out <- sev:
		case <-ctx.Done():
	}
	return
}

// Generate aggregated report based on priority
// Send all inc details into one channel
// Fan out that channel to bounded go routines. This will merge the values and
// send to single output channel
// We can have any levels of merging depending on load
func generateAggReportPriority(report []Incident) (sum *[]PrioritySum,err error ){

	// create context with cancel to inform goroutines to exit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send all inc details into one channel incs
	incs, errc := walkIncs(ctx, report)
	if errc != nil {
		return nil, errc
	}

	// Fan out the channel `incs` to bounded go routines. This will merge the values and
	// send to single output channel
	// This provides first level of merge on the data
	c := make(chan map[string]int)
	// use sync.WaitGroup for synchronization
	var wg sync.WaitGroup
	// NumGoRoutines - controls number of goroutines that can be spawned
	wg.Add(NumGoRoutines)

	// spawn NumGoRoutines times mergeIncs()
	for i := 0; i < NumGoRoutines; i++ {
		go func() {
			mergeIncs(ctx, incs, c)
			wg.Done()
		}()
	}

	// wait for all goroutines to end before closing channel c
	go func() {
		wg.Wait()
		close(c)
	}()

	// final merge - call mergeIncs one more time for final merge
	// make sure to close the channel
	final := make(chan map[string]int,1)
	go func() {
		defer close(final)
		mergeIncs(ctx, c, final)
	}()

	// read the final channel and create []PrioritySum struct
	var sumObj []PrioritySum
	for obj := range final {
		for k, v := range obj {
			sumObj = append(sumObj, PrioritySum{k, v})
		}
	}
	return &sumObj, nil
}

func main() {
	// get url as first arg
	url := os.Args[1]

	// get the response using http client
	res := GetResponse(url)

	// validate response based on headers
	respErr := ValidateResponse(res)
	if respErr != nil {
		log.Fatal(respErr)
	}

	// test initial no of goroutines
	fmt.Printf("Goroutine count %d\n", countGoRoutines())

	// Read the body
	// TODO - split readAll
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	defer res.Body.Close()

	// encode the body into json with given incidents struct
	var incidents Incidents
	jsonErr := json.Unmarshal(body, &incidents)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	// print the tableFormat of the incidents report
	tableFmt, err := tableFormat.Format(incidents.Report)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n",*tableFmt)

	// generate aggregated report based on priority
	aggReport, err := generateAggReportPriority(incidents.Report)
	if err != nil {
		log.Fatal(err)
	}

	aggTableFmt, err := tableFormat.Format(*aggReport)
	fmt.Printf("%s\n",*aggTableFmt)
	if err != nil {
		log.Fatal(err)
	}

	// print the final count of go routines
	fmt.Printf("Goroutine count %d\n",countGoRoutines())
	//pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)

}
