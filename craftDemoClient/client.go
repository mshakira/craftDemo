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
	TIMEOUT = 2 // no of secs for url timeout
	RETRY = 5 // no of retries for url failures
	NUM_GO_ROUTINES = 3 // maximum number of go routines to fan out
)

type Incidents struct {
	Name   string     `json:"Name"`
	Report []Incident `json:"Report"`
}

type Incident struct {
	Number      string `json:"number"`
	AssignedTo  string `json:"assigned_to"`
	Description string `json:"description"`
	State       string `json:"state"`
	Priority    string `json:"priority"`
	Severity    string `json:"severity"`
}

type PrioritySum struct {
	Priority string
	Sum      int
}

func countGoRoutines() int {
	return runtime.NumGoroutine()
}

func mapFunc(ctx context.Context, priority string) <-chan map[string]int {
	out := make(chan map[string]int)
	go func() {
		defer close(out)
		m := make(map[string]int)
		m[priority] = 1
		select {
		case out <- m:
		case <-ctx.Done():
			return // returning not to leak the goroutine
		}
	}()
	return out
}

func ReduceFunc(ctx context.Context, ch []<-chan map[string]int) <-chan map[string]int {
	var wg sync.WaitGroup
	out := make(chan map[string]int)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan map[string]int) {
		for n := range c {
			select {
			case out <- n:
			case <-ctx.Done():
				return // returning not to leak the goroutine
			}
		}
		wg.Done()
	}
	wg.Add(len(ch))
	for _, c := range ch {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		fmt.Printf("All gorountines stopped\n")
		wg.Wait()
		close(out)
	}()
	return out
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

func ValidateResponse(res *http.Response) (err error) {
	var resLength int
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
func mergeIncs(incs chan map[string]int, out chan map[string]int) {
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
	out <- sev
}

func generateAggReportPriority(report []Incident) (sum *[]PrioritySum,err error ){
	// Send all inc details into one channel
	// Fan out that channel to bounded go routines. This will merge the values and
	// send to single output channel
	// final merge will happen in single output channel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send all inc details into one channel incs

	incs, errc := walkIncs(ctx, report)
	if errc != nil {
		return nil, errc
	}

	// Fan out that channel to bounded go routines. This will merge the values and
	// send to single output channel
	// This provides first level of merge on the data
	c := make(chan map[string]int)
	var wg sync.WaitGroup
	wg.Add(NUM_GO_ROUTINES)

	for i := 0; i < NUM_GO_ROUTINES; i++ {
		go func() {
			mergeIncs(incs, c)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	// final merge
	final := make(chan map[string]int)
	go func() {
		defer close(final)
		mergeIncs(c, final)
	}()

	var sumObj []PrioritySum
	for obj := range final {
		for k, v := range obj {
			sumObj = append(sumObj, PrioritySum{k, v})
		}
	}
	return &sumObj, nil
}

func main() {
	url := os.Args[1]

	// get the response using http client
	res := GetResponse(url)

	// validate response based on headers
	respErr := ValidateResponse(res)
	if respErr != nil {
		log.Fatal(respErr)
	}

	fmt.Printf("Goroutine count %d\n", countGoRoutines())

	// Read the body
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
	tableFmt := tableFormat.Format(incidents.Report)
	fmt.Printf("%s\n",tableFmt)

	// generate aggregated report based on priority
	aggReport, err := generateAggReportPriority(incidents.Report)
	if err != nil {
		log.Fatal(err)
	}

	aggTableFmt := tableFormat.Format(*aggReport)
	fmt.Printf("%s\n",aggTableFmt)
	//fmt.Printf("Goroutine count %d\n",countGoRoutines())
	//pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)

}
