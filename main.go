package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"web-page-analyzer/handlers"
	"web-page-analyzer/middleware"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := handlers.NewServer()

	// Create middleware chain
	middlewareChain := middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Route handling
			switch r.URL.Path {
			case "/":
				server.IndexHandler(w, r)
			case "/analyze":
				server.AnalyzeHandler(w, r)
			default:
				http.NotFound(w, r)
			}
		}),
		middleware.PanicRecovery,
		middleware.Logging,
		middleware.CORS,
		middleware.SecurityHeaders,
		middleware.Timeout(30*time.Second),
	)

	// Serve static files with middleware
	staticHandler := middleware.Chain(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
		middleware.PanicRecovery,
		middleware.Logging,
		middleware.SecurityHeaders,
	)

	// Set up routes
	http.Handle("/static/", staticHandler)
	http.Handle("/", middlewareChain)

	log.Printf("Server starting on port %s", port)
	log.Printf("Visit http://localhost:%s to use the application", port)
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}