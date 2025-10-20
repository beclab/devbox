package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var startupMessage = "Welcome to the olares world!"

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	envVar := os.Getenv("OLARES_STUDIO_HELLO_MESSAGE")

	response := startupMessage

	if envVar != "" {
		response = envVar
	}

	fmt.Fprint(w, response)
}

func main() {
	if len(os.Args) > 1 {
		startupMessage = strings.Join(os.Args[1:], ",")
	}
	port := "80"

	http.HandleFunc("/", helloHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
