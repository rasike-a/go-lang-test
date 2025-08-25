package main

import (
	"log"
	"net/http"
	"os"

	"web-page-analyzer/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := handlers.NewServer()

	http.HandleFunc("/", server.IndexHandler)
	http.HandleFunc("/analyze", server.AnalyzeHandler)

	log.Printf("Server starting on port %s", port)
	log.Printf("Visit http://localhost:%s to use the application", port)
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}