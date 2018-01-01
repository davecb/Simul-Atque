package main

import (
	"net/http/httputil"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"flag"
	"strconv"
)

var start = make(chan bool)
var done = make(chan bool)
var port = 5280
var serviceTime = 100.0  // milliseconds
var serviceDuration time.Duration
var queuingCenters = 1
var bytes = 0
var host = ":5280"
var verbose = false


// main starts the web server, and also a smoke test for it
func main() {
	// parse opts
	flag.IntVar(&queuingCenters, "servers", 1, "number of servers")
	flag.Float64Var(&serviceTime, "service-time", 100, "service time in milliseconds")
	flag.IntVar(&bytes, "bytes", 0, "bytes to return, currently locked to 0")
	flag.IntVar(&port, "port", 5280, "port to use")
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.Parse()

	host = ":" + strconv.FormatUint(uint64(port), 10)
	serviceDuration = time.Duration(serviceTime) * time.Millisecond
	for i := 0; i < queuingCenters; i++ {
		go queuingCenter(start, done)
	}

	go runSmokeTest()
	fmt.Print("#date      time         latency  xferTime thinkTime bytes key rc op\n")
	startWebserver()
}

// queuingCenter gets work, does it and reports completion
func queuingCenter(start, done chan bool) {
	for {
		<- start
		time.Sleep(serviceDuration) // "work"
		done <- true
	}
}

// startWebserver for all object requests
func startWebserver() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// finite server: bottleneck is done channel

		initial := time.Now()
		start <- true
		// work happens here, in the queuing centre(s)
		<- done
		w.Write([]byte("success!\n"))
		go func() {
			// print in a goroutine
			end := time.Since(initial)
			fmt.Printf("%s %f 0.0 0.0 0 %s 200 GET\n",
				initial.Format("2006-01-02 15:04:05.000"),
				end.Seconds(), r.RequestURI)
		}()
		// return from the HandleFunc sends/closes the response

		// infinite number of servers...
		//initial := time.Now()
		//time.Sleep(serviceDuration)
		//w.Write(nil)
		//end := time.Since(initial)
		//fmt.Printf("%s %f 0.0 0.0 0 %s 200 GET\n",
		//	initial.Format("2006-01-02 15:04:05.000"),
		//	end.Seconds(), r.RequestURI)
	})
	err := http.ListenAndServe(host, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}


// runSmokeTest checks that the server is up, panics if not
func runSmokeTest() {
	time.Sleep(time.Second * 2)
	key := "albert/the/alligator.html"
	resp, err := http.Get("http://" + host + "/" + key)
	if err != nil {
		panic(fmt.Sprintf("Got an error in the get: %v", err))
	}
	body, err :=  ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Got an error in the body read: %v", err))
	}
	if verbose {
		fmt.Printf("\n%s\n", responseToString(resp))
		fmt.Printf("\n%s\n", bodyToString(body))
	}
	err = resp.Body.Close()
	if err != nil {
		panic(fmt.Sprintf("Got an error in the body close: %v", err))
	}
}

// requestToString provides extra information about an http request if it can
func requestToString(req *http.Request) string {
	var dump []byte
	var err error

	if req == nil {
		return "Request: <nil>\n"
	}
	dump, err = httputil.DumpRequestOut(req, true)
	if err != nil {
		return fmt.Sprintf("fatal error dumping http request, %v\n", err)
	}
	return fmt.Sprintf("Request: \n%s", dump)
}

// responseToString provides extra information about an http response
func responseToString(resp *http.Response) string {
	if resp == nil {
		return "Response: <nil>\n"
	}
	s := requestToString(resp.Request)
	contents, err := httputil.DumpResponse(resp, false)
	if err != nil {
		return fmt.Sprintf("error dumping http response, %v\n", err)
	}
	s += "Response information:\n"
	s += fmt.Sprintf("    Length: %d\n", resp.ContentLength)
	s += fmt.Sprintf("    Status code: %d\n", resp.StatusCode)
	s += fmt.Sprintf("Response contents: \n%s", string(contents))
	return s
}

// bodyToString provides the body
func bodyToString(body []byte) string {
	if body == nil {
		return "Body: <nil>\n"
	}
	return fmt.Sprintf("Body:\n %s\n", body)
}
