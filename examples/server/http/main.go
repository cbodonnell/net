package main

import (
	"flag"
	"fmt"
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
		fmt.Fprintf(w, "%s %s %s\n", r.Method, r.URL, r.Proto)
		for k, v := range r.Header {
			fmt.Fprintf(w, "Header field %q, Value %q\n", k, v)
		}
		fmt.Fprintf(w, "Host = %q\n", r.Host)
		fmt.Fprintf(w, "RemoteAddr= %q\n", r.RemoteAddr)
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
		}
		for k, v := range r.Form {
			fmt.Fprintf(w, "Form field %q, Value %q\n", k, v)
		}
	}

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(portString, nil))
}
