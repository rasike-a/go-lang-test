package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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

	// Create middleware chain for main routes
	middlewareChain := middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Route handling
			switch r.URL.Path {
			case "/":
				server.IndexHandler(w, r)
			case "/analyze":
				server.AnalyzeHandler(w, r)
			case "/metrics":
				handleMetrics(w, r, server)
			case "/health":
				handleHealth(w, r)
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

	// Enable profiling endpoints in development
	if os.Getenv("ENV") != "production" {
		// Add pprof endpoints for profiling
		http.Handle("/debug/pprof/", http.DefaultServeMux)
		log.Println("Profiling enabled at /debug/pprof/")
	}

	// Create HTTP server with optimized settings
	httpServer := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		// Performance optimizations
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		log.Printf("Visit http://localhost:%s to use the application", port)
		log.Printf("Metrics available at http://localhost:%s/metrics", port)
		if os.Getenv("ENV") != "production" {
			log.Printf("Profiling available at http://localhost:%s/debug/pprof/", port)
		}
		
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}

// handleMetrics returns analyzer performance metrics
func handleMetrics(w http.ResponseWriter, r *http.Request, server *handlers.Server) {
	w.Header().Set("Content-Type", "application/json")
	
	// Get analyzer metrics
	analyzer := server.GetAnalyzer()
	if analyzer == nil {
		http.Error(w, "Analyzer not available", http.StatusServiceUnavailable)
		return
	}

	metrics := analyzer.GetMetrics()
	
	// Add runtime metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	response := map[string]interface{}{
		"analyzer": map[string]interface{}{
			"total_requests":   metrics.TotalRequests,
			"active_requests":  metrics.ActiveRequests,
			"total_duration":   metrics.TotalDuration.String(),
			"avg_duration":     metrics.AvgDuration.String(),
			"cache_hits":       metrics.CacheHits,
			"cache_misses":     metrics.CacheMisses,
		},
		"runtime": map[string]interface{}{
			"goroutines":       runtime.NumGoroutine(),
			"memory_alloc":     m.Alloc,
			"memory_sys":       m.Sys,
			"memory_heap_alloc": m.HeapAlloc,
			"memory_heap_sys":   m.HeapSys,
			"gc_cycles":        m.NumGC,
			"gc_pause_total":   m.PauseTotalNs,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// handleHealth returns server health status
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
	}
	
	json.NewEncoder(w).Encode(response)
}

var startTime = time.Now()