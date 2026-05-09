package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))

	port := "8080"
	srv := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	if err := srv.ListenAndServe(); err != nil {
		fmt.Errorf("server startup error: %w", err)
	}
}
