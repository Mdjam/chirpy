package main

import (
	"log"
	"net/http"
)

const port = "8080"

func main() {
	mux := http.NewServeMux()

	server := http.Server {
		Addr: ":" + port,
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Could not start server: %s", err)
	}
}
