package main

import (
	"flag"
	"fmt"
	"io"
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

	handler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
	}

	http.HandleFunc("/hello", handler)
	log.Fatal(http.ListenAndServe(portString, nil))
}
