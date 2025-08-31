package analyzer

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"web-page-analyzer/logger"
)

// analyzeLinksConcurrent analyzes links concurrently using a worker pool
func (a *Analyzer) analyzeLinksConcurrent(links []string, baseURL *url.URL, result *AnalysisResult) {
	if len(links) == 0 {
		return
	}
	
	// Determine optimal number of workers based on link count
	workers := a.calculateOptimalWorkers(len(links))
	
	// Create and start worker pool
	pool := NewAnalysisWorkerPool(workers, a)
	pool.Start()
	defer pool.Stop()
	
	// Submit all jobs
	for _, link := range links {
		pool.SubmitJob(AnalysisJob{
			Link:    link,
			BaseURL: baseURL.String(),
		})
	}
	
	// Collect results
	startTime := time.Now()
	internalCount := 0
	externalCount := 0
	inaccessibleCount := 0
	
	// Process results with timeout
	timeout := time.After(30 * time.Second)
	resultsReceived := 0
	
	for resultsReceived < len(links) {
		select {
		case linkResult := <-pool.GetResults():
			resultsReceived++
			
			if linkResult.Error != nil {
				logger.WithAnalysis(baseURL.String()).Errorw("Link analysis error",
					"link", linkResult.Link,
					"error", linkResult.Error,
				)
				continue
			}
			
			if linkResult.IsInternal {
				internalCount++
			} else {
				externalCount++
				if !linkResult.IsAccessible {
					inaccessibleCount++
					logger.WithAnalysis(baseURL.String()).Infow("Link inaccessible",
						"link", linkResult.Link,
					)
				}
			}
			
		case <-timeout:
			logger.WithAnalysis(baseURL.String()).Warnw("Link analysis timeout",
				"links_processed", resultsReceived,
				"total_links", len(links),
			)
			goto done
		}
	}
	
done:
	duration := time.Since(startTime)
	
	// Update result
	result.InternalLinks = internalCount
	result.ExternalLinks = externalCount
	result.InaccessibleLinks = inaccessibleCount
	
	logger.WithAnalysis(baseURL.String()).Infow("Links analysis completed",
		"total", len(links),
		"skipped", len(links)-resultsReceived,
		"internal", internalCount,
		"external", externalCount,
		"inaccessible", inaccessibleCount,
		"duration_ms", duration.Milliseconds(),
		"workers", workers,
	)
}

// calculateOptimalWorkers calculates the optimal number of workers based on link count
func (a *Analyzer) calculateOptimalWorkers(linkCount int) int {
	// Scale workers based on link count
	switch {
	case linkCount <= 5:
		return 1
	case linkCount <= 20:
		return 5
	case linkCount <= 50:
		return 10
	case linkCount <= 100:
		return 15
	default:
		return 20 // Cap at 20 workers to prevent resource exhaustion
	}
}

// isLinkAccessible checks if a link is accessible by making a HEAD request
func (a *Analyzer) isLinkAccessible(link string) bool {
	// Skip special protocols
	if strings.HasPrefix(link, "javascript:") || 
	   strings.HasPrefix(link, "mailto:") || 
	   strings.HasPrefix(link, "tel:") {
		return false
	}
	
	// Get HTTP client from pool
	client := a.getHTTPClient()
	defer a.putHTTPClient(client)
	
	// Create request with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "HEAD", link, nil)
	if err != nil {
		return false
	}
	
	// Set user agent to avoid being blocked
	req.Header.Set("User-Agent", "WebPageAnalyzer/1.0")
	
	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	// Check if response indicates accessibility
	// 2xx and 3xx status codes are generally considered accessible
	// 4xx and 5xx indicate the resource is not accessible
	return resp.StatusCode < 400
}

// getHTTPClient gets an HTTP client from the pool
func (a *Analyzer) getHTTPClient() *http.Client {
	if client, ok := a.httpClientPool.Get().(*http.Client); ok {
		return client
	}
	return a.httpClient
}

// putHTTPClient returns an HTTP client to the pool
func (a *Analyzer) putHTTPClient(client *http.Client) {
	a.httpClientPool.Put(client)
}
