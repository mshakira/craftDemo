package main

import (
	"context"
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

func GetResponse(url string) (res *http.Response) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{
		Transport: tr,
		Timeout:   time.Second * 2, // Maximum of 2 secs
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Fatal(err)
	}

	retry := 5

	for i := 1; i <= retry; i++ {

		var getErr error

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

func walkIncs(ctx context.Context, report []Incident) (chan map[string]int, error) {
	out := make(chan map[string]int)
	defer close(out)
	for _, obj := range report {
		//ch := mapFunc(obj.Priority)
		//fmt.Println(<-ch)
		m := make(map[string]int)
		m[obj.Priority] = 1
		select {
		case out <- m:
		case <-ctx.Done():
			return out, nil
		}
	}
	return out, nil
}

func mergeIncs(ctx context.Context, incs chan map[string]int, out chan map[string]int) {
	sev := make(map[string]int)
	for {
		select {
		case obj := <-incs:
			for k, v := range obj {
				if _, ok := sev[k]; ok {
					sev[k] += v
				} else {
					sev[k] = v
				}
			}
			case <-ctx.Done():
				return
		}
	}
}

func main() {
	url := os.Args[1]

	res := GetResponse(url)

	respErr := ValidateResponse(res)
	if respErr != nil {
		log.Fatal(respErr)
	}

	fmt.Printf("Goroutine count %d\n", countGoRoutines())

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	defer res.Body.Close()

	var incidents Incidents

	jsonErr := json.Unmarshal(body, &incidents)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	//tableFormat.Format(incidents.Report)

	// Send all inc details into one channel
	// Fan out that channel to bounded go routines. This will merge the values and
	// send to single output channel
	// final merge will happen in single output channel

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send all inc details into one channel incs
	incs, errc := walkIncs(ctx, incidents.Report)
	if errc != nil {
		fmt.Println(errc)
		return
	}

	// Fan out that channel to bounded go routines. This will merge the values and
	//	// send to single output channel
	// This provides first level of merge on the data
	c := make(chan map[string]int)
	var wg sync.WaitGroup
	const numMergeIncs = 3
	wg.Add(numMergeIncs)

	for i := 0; i < numMergeIncs; i++ {
		go func() {
			mergeIncs(ctx, incs, c)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	final := make(chan map[string]int)
	go func() {
		defer close(final)
		mergeIncs(ctx, c, final)
	}()

	sev := make(map[string]int)
	for obj := range final {
		for k, v := range obj {
			fmt.Printf("New K %v v %v\n ", k, v)
		}
	}

	/*var ch []<-chan map[string]int

	for _, obj := range incidents.Report {
		//ch := mapFunc(obj.Priority)
		//fmt.Println(<-ch)
		ch = append(ch, mapFunc(ctx, obj.Priority))
	}
	// TODO bounded go routines
	/*sev := make(map[string]int)
	for obj := range ReduceFunc(ctx, ch) {
		for k, v := range obj {
			if _, ok := sev[k]; ok {
				sev[k] += v
			} else {
				sev[k] = v
			}
		}
	}*/

	/*var sum []PrioritySum

	for k, v := range sev {
		sum = append(sum,PrioritySum{k,v})
		//fmt.Printf("%v %v\n", k, v)
	}
	tableFormat.Format(sum)*/
	//fmt.Printf("Goroutine count %d\n",countGoRoutines())
	//pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)

}
