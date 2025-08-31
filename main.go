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
	"web-page-analyzer/logger"
	"web-page-analyzer/middleware"
)

var startTime = time.Now()

func main() {
	// Initialize structured logger
	logger.Init()
	defer logger.Sync()

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
			case "/cache-logging":
				handleCacheLogging(w, r, server)
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
			logger.Sugar.Infof("Server starting on port %s", port)
			logger.Sugar.Infof("Visit http://localhost:%s to use the application", port)
			logger.Sugar.Infof("Metrics available at http://localhost:%s/metrics", port)
			if os.Getenv("ENV") != "production" {
				logger.Sugar.Infof("Profiling available at http://localhost:%s/debug/pprof/", port)
			}

			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Sugar.Fatal("Server failed to start:", err)
			}
		}()

			// Wait for interrupt signal to gracefully shutdown the server
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		logger.Sugar.Info("Server shutting down...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

			// Attempt graceful shutdown
		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Sugar.Fatal("Server forced to shutdown:", err)
		}

		logger.Sugar.Info("Server exited gracefully")
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

// handleHealth returns system health status
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	uptime := time.Since(startTime)
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    uptime.String(),
	}
	
			if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Sugar.Errorw("Health response encoding error", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
}

// handleCacheLogging controls cache logging verbosity
func handleCacheLogging(w http.ResponseWriter, r *http.Request, server *handlers.Server) {
	if r.Method == http.MethodGet {
		// Get current cache logging status
		analyzer := server.GetAnalyzer()
		if analyzer == nil {
			http.Error(w, "Analyzer not available", http.StatusServiceUnavailable)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"cache_logging": "Use POST to control cache logging verbosity",
			"usage": map[string]string{
				"POST /cache-logging?verbose=true":  "Enable verbose cache logging",
				"POST /cache-logging?verbose=false": "Disable verbose cache logging",
			},
		}
		
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Sugar.Errorw("Cache logging response encoding error", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	
	if r.Method == http.MethodPost {
		// Control cache logging verbosity
		verbose := r.URL.Query().Get("verbose")
		analyzer := server.GetAnalyzer()
		if analyzer == nil {
			http.Error(w, "Analyzer not available", http.StatusServiceUnavailable)
			return
		}
		
		switch verbose {
		case "true":
			analyzer.SetCacheVerbose(true)
			logger.Sugar.Info("Cache verbose logging enabled")
		case "false":
			analyzer.SetCacheVerbose(false)
			logger.Sugar.Info("Cache verbose logging disabled")
		default:
			http.Error(w, "Invalid verbose parameter. Use 'true' or 'false'", http.StatusBadRequest)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"cache_logging": verbose == "true",
			"message":       "Cache logging verbosity updated",
			"timestamp":     time.Now().UTC().Format(time.RFC3339),
		}
		
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Sugar.Errorw("Cache logging response encoding error", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}