package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	var port uint

	flag.UintVar(&port, "port", 5555, "The port to listen on")
	flag.Usage = usage
	flag.Parse()

	portString := fmt.Sprintf(":%d", port)

	log.Printf("Listening on %s\n", portString)

	handler := func(w http.ResponseWriter, r *http.Request) {
		var requestDump bytes.Buffer
		// Write the request method, URL, and protocol
		requestDump.WriteString(fmt.Sprintf("%s %s %s\r\n", r.Method, r.URL, r.Proto))
		// Write the headers
		for k, v := range r.Header {
			requestDump.WriteString(fmt.Sprintf("%s: %s\r\n", k, v[0]))
		}
		// Write the Host header
		requestDump.WriteString(fmt.Sprintf("Host: %s\r\n", r.Host))
		// Write the body
		body, _ := ioutil.ReadAll(r.Body)
		requestDump.Write(body)
		// Write the request to the response
		w.Write(requestDump.Bytes())
	}

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(portString, nil))
}
