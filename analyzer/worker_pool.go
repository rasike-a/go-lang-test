package analyzer

import (
	"net/url"
	"strings"
)

// NewAnalysisWorkerPool creates a new worker pool for link analysis
func NewAnalysisWorkerPool(workers int, analyzer *Analyzer) *AnalysisWorkerPool {
	// Increase buffer sizes for high-throughput processing
	bufferSize := workers * 4 // 4x buffer for better throughput
	
	return &AnalysisWorkerPool{
		workers:  workers,
		jobQueue: make(chan AnalysisJob, bufferSize),
		results:  make(chan LinkResult, bufferSize),
		stopChan: make(chan struct{}),
		analyzer: analyzer,
	}
}

// Start starts the worker pool
func (wp *AnalysisWorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.workerWg.Add(1)
		go wp.worker()
	}
}

// Stop stops the worker pool and waits for all workers to finish
func (wp *AnalysisWorkerPool) Stop() {
	close(wp.stopChan)
	close(wp.jobQueue)
	wp.workerWg.Wait()
	close(wp.results)
}

// SubmitJob submits a job to the worker pool
func (wp *AnalysisWorkerPool) SubmitJob(job AnalysisJob) {
	select {
	case wp.jobQueue <- job:
		// Job submitted successfully
	default:
		// Job queue is full, process synchronously for high-throughput
		// This prevents blocking when processing many links
		wp.processJob(job)
	}
}

// GetResults returns the results channel
func (wp *AnalysisWorkerPool) GetResults() <-chan LinkResult {
	return wp.results
}

// worker is the main worker goroutine
func (wp *AnalysisWorkerPool) worker() {
	defer wp.workerWg.Done()
	
	for {
		select {
		case job, ok := <-wp.jobQueue:
			if !ok {
				return
			}
			wp.processJob(job)
		case <-wp.stopChan:
			return
		}
	}
}

// processJob processes a single analysis job
func (wp *AnalysisWorkerPool) processJob(job AnalysisJob) {
	baseURL, err := url.Parse(job.BaseURL)
	if err != nil {
		wp.results <- LinkResult{
			Link:        job.Link,
			IsInternal:  false,
			IsAccessible: false,
			Error:       err,
		}
		return
	}
	
	result := wp.analyzer.analyzeSingleLink(job.Link, baseURL)
	wp.results <- result
}

// analyzeSingleLink analyzes a single link for accessibility and type
func (a *Analyzer) analyzeSingleLink(link string, baseURL *url.URL) LinkResult {
	// Skip empty links and fragments
	if link == "" || strings.HasPrefix(link, "#") {
		return LinkResult{
			Link:        link,
			IsInternal:  false,
			IsAccessible: false,
			Error:       nil,
		}
	}
	
	// Parse the link URL
	linkURL, err := url.Parse(link)
	if err != nil {
		return LinkResult{
			Link:        link,
			IsInternal:  false,
			IsAccessible: false,
			Error:       err,
		}
	}
	
	// Resolve relative URLs
	if !linkURL.IsAbs() {
		linkURL = baseURL.ResolveReference(linkURL)
	}
	
	// Determine if link is internal or external
	isInternal := linkURL.Hostname() == baseURL.Hostname()
	
	// Check if link is accessible (only for external links to avoid infinite loops)
	var isAccessible bool
	if !isInternal {
		isAccessible = a.isLinkAccessible(linkURL.String())
	} else {
		isAccessible = true // Assume internal links are accessible
	}
	
	return LinkResult{
		Link:        link,
		IsInternal:  isInternal,
		IsAccessible: isAccessible,
		Error:       nil,
	}
}
