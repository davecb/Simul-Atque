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
	"flag"
)

const (
	host = ":5280"
)

// a req is a request to be run by a handler
type req struct {
	initial time.Time
	w http.ResponseWriter
	r *http.Request
}

// t is a debugging tool shared by the server components
var t = trace.New(os.Stderr, true) // or (nil, false)
var pipe chan req
var delay = 10.0
var count = 1


// main starts the web server, and also a smoke test for it
func main() {
	// parse opts
	flag.IntVar(&count, "servers", 1, "number of servers, default 1")
	flag.Float64Var(&delay, "service-time", 10, "service time in milliseconds, default 10")
	flag.Parse()

	pipe = make(chan req, count)
	for i := 0; i < count; i++ {
		go worker(pipe)
	}
	go runSmokeTest()
	startWebserver()
}

// worker reads r and w from a pipe and does some work
func worker(pipe chan req) {
	for x := range pipe {
		var w = x.w

		time.Sleep(time.Duration(delay) * time.Millisecond)
		_, err := w.Write(nil)  // FIXME, fails if non-nil
		end := time.Since(x.initial)
		if err != nil {
			// log and try to return 500 via the broken ResponseWriter
			t.Printf(`ERROR, worker could not write to ResponseWriter, "%v"\n`, err)
			http.Error(w, err.Error(), 500)
		}
		end -= time.Duration(10.0 * time.Millisecond)
		fmt.Printf("%s %f 0.010 0 %s 200 GET\n",
			x.initial.Format("2006-01-02 15:04:05.000"),
			end.Seconds(), x.r.RequestURI)
	}

}

// startWebserver for all object requests
func startWebserver() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pipe <- req{ time.Now(),w, r}
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
	_, err =  ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Got an error in the body read: %v", err))
	}
	t.Printf("\n%s\n", responseToString(resp))
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
