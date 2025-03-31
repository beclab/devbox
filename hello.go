package main

import (
	"fmt"
	"log"
	"net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World")
}

func main() {
	http.HandleFunc("/", helloHandler)

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatalf("server startup failed: %v\n", err)
	}
}
