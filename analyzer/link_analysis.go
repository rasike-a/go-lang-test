package analyzer

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"sync"
	"web-page-analyzer/logger"
)

// analyzeLinksConcurrent analyzes links concurrently using a worker pool
func (a *Analyzer) analyzeLinksConcurrent(links []string, baseURL *url.URL, result *AnalysisResult) {
	if len(links) == 0 {
		return
	}

	// For high-link sites like GitHub, use ultra-aggressive parallel processing
	workers := a.calculateOptimalWorkers(len(links))

	logger.WithAnalysis(baseURL.String()).Infow("Starting parallel link analysis",
		"total_links", len(links),
		"workers", workers,
	)

	// Create channels for parallel processing
	jobs := make(chan string, len(links))
	results := make(chan LinkResult, len(links))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for link := range jobs {
				// Process link in parallel
				result := a.processLinkParallel(link, baseURL)
				results <- result
			}
		}(i)
	}

	// Submit all jobs
	go func() {
		for _, link := range links {
			jobs <- link
		}
		close(jobs)
	}()

	// Collect results with optimized timeout
	startTime := time.Now()
	internalCount := 0
	externalCount := 0
	inaccessibleCount := 0

	// Dynamic timeout based on link count - capped at 45 seconds for high-link sites
	timeoutDuration := time.Duration(len(links)/3) * time.Second
	if timeoutDuration < 30*time.Second {
		timeoutDuration = 30 * time.Second
	}
	if timeoutDuration > 45*time.Second {
		timeoutDuration = 45 * time.Second
	}

	timeout := time.After(timeoutDuration)
	resultsReceived := 0

	logger.WithAnalysis(baseURL.String()).Infow("Link analysis timeout configured",
		"timeout_duration", timeoutDuration,
		"total_links", len(links),
	)

	// Process results with early completion for high-link sites
	for resultsReceived < len(links) {
		select {
		case linkResult := <-results:
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
				}
			}

			// For high-link sites, log progress every 20 links
			if len(links) > 50 && resultsReceived%20 == 0 {
				logger.WithAnalysis(baseURL.String()).Infow("Link analysis progress",
					"processed", resultsReceived,
					"total", len(links),
					"internal", internalCount,
					"external", externalCount,
					"duration_ms", time.Since(startTime).Milliseconds(),
				)
			}

		case <-timeout:
			logger.WithAnalysis(baseURL.String()).Warnw("Link analysis timeout",
				"links_processed", resultsReceived,
				"total_links", len(links),
				"timeout_duration", timeoutDuration,
			)
			goto done
		}
	}

done:
	// Wait for all workers to finish
	wg.Wait()
	close(results)

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
		"timeout_duration", timeoutDuration,
	)
}

// processLinkParallel processes a single link in parallel
func (a *Analyzer) processLinkParallel(link string, baseURL *url.URL) LinkResult {
	linkProcessor := NewLinkProcessor()
	return linkProcessor.ProcessLink(link, baseURL, a.isLinkAccessible)
}

// calculateOptimalWorkers calculates the optimal number of workers based on link count
func (a *Analyzer) calculateOptimalWorkers(linkCount int) int {
	// Ultra-aggressive scaling for high-link sites like GitHub
	// This ensures maximum parallelization for complex sites
	switch {
	case linkCount <= 10:
		return MinWorkers
	case linkCount <= 25:
		return 12
	case linkCount <= 50:
		return 24
	case linkCount <= 100:
		return 48
	case linkCount <= 150:
		return 64
	case linkCount <= 200:
		return 80
	default:
		return MaxWorkers // Maximum workers for ultra-high-link sites
	}
}

// isLinkAccessible checks if a link is accessible by making a HEAD request
func (a *Analyzer) isLinkAccessible(link string) bool {
	linkProcessor := NewLinkProcessor()

	// Skip special protocols
	if linkProcessor.IsSpecialProtocol(link) {
		return false
	}

	// Create HTTP request with timeout
	client := a.getHTTPClient()
	defer a.putHTTPClient(client)

	req, err := http.NewRequest("HEAD", link, nil)
	if err != nil {
		return false
	}

	// Set realistic headers to avoid bot detection
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")

	// Make request with optimized timeout (3 seconds for faster response)
	ctx, cancel := context.WithTimeout(context.Background(), LinkCheckTimeout)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		// Log timeout or connection errors for debugging
		if ctx.Err() == context.DeadlineExceeded {
			logger.WithAnalysis(link).Debugw("Link check timeout", "timeout", "3s")
		}
		return false
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.WithAnalysis(link).Debugw("Failed to close response body", "error", closeErr)
		}
	}()

	// Consider 2xx and 3xx status codes as accessible
	// Early success detection - no need to wait longer once we get a response
	return resp.StatusCode >= 200 && resp.StatusCode < 400
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
