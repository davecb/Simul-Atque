package main

import (
	"github.com/davecb/trace"
	
	"net/http/httputil"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"os"
)


// t is a debugging tool shared by the server components
var t = trace.New(os.Stderr, true) // or (nil, false)
// logger goes to stdout, as do timing records for each access
var logger = log.New(os.Stdout, "", log.Ldate | log.Ltime | log.Lshortfile)

const (
	//host = "10.92.10.201:5280"
	host = ":5280"
)


// main starts the web server, and also a smoke test for it
func main() {
	defer t.Begin()()

	go runSmokeTest()
	startWebserver()
}

// startWebserver for all object requests
func startWebserver() {
	defer t.Begin()()

	// handle image vs content part of prefixes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer t.Begin(r)()

		t.Printf("started\n")
		time.Sleep(10 * time.Millisecond)
		_, err := w.Write([]byte("success"))
		if err != nil {
			// log and try to return 500 via the broken ResponseWriter
			t.Printf("ERROR, could not write to ResponseWriter, %v\n", err)
			http.Error(w, err.Error(), 500)
		}
	})

	err := http.ListenAndServe(host, nil) // nolint
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
	t.Printf("\n%s\n%s\n", responseToString(resp), bodyToString(body))
	resp.Body.Close()         // nolint
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
