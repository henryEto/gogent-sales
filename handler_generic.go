package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func handlerGeneric(w http.ResponseWriter, r *http.Request) {
	printRequest(r)
	w.WriteHeader(http.StatusOK)
}

func printRequest(r *http.Request) {
	log.Println("--- Incoming Request ---")
	log.Printf("Method: %s", r.Method)
	log.Printf("URL: %s", r.URL.String())
	log.Printf("Host: %s", r.Host)
	log.Printf("Protocol: %s", r.Proto)
	log.Println("Headers:")
	for name, values := range r.Header {
		for _, value := range values {
			log.Printf("  %s: %s", name, value)
		}
	}

	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
		} else {
			log.Printf("Body:\n%s", string(bodyBytes))
		}

	} else {
		log.Println("No request body")
	}
	log.Println("------------------------")
	fmt.Printf("\n\n")
}
