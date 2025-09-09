package analyzer

import (
	"net/url"
)

// NewAnalysisWorkerPool creates a new worker pool for link analysis
func NewAnalysisWorkerPool(workers int, analyzer *Analyzer) *AnalysisWorkerPool {
	// Increase buffer sizes for high-throughput processing
	bufferSize := workers * BufferMultiplier // 4x buffer for better throughput

	return &AnalysisWorkerPool{
		workers:  workers,
		jobQueue: make(chan AnalysisJob, bufferSize),
		results:  make(chan LinkResult, bufferSize),
		stopChan: make(chan struct{}),
		analyzer: analyzer,
	}
}

// Starts the worker pool and begins processing jobs
func (wp *AnalysisWorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.workerWg.Add(1)
		go wp.worker()
	}
}

// Stops the worker pool gracefully and waits for all workers to finish
func (wp *AnalysisWorkerPool) Stop() {
	close(wp.stopChan)
	close(wp.jobQueue)
	wp.workerWg.Wait()
	close(wp.results)
}

// SubmitJob submits a job to the worker pool for processing
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
			Link:         job.Link,
			IsInternal:   false,
			IsAccessible: false,
			Error:        err,
		}
		return
	}

	result := wp.analyzer.analyzeSingleLink(job.Link, baseURL)
	wp.results <- result
}

// analyzeSingleLink analyzes a single link for accessibility and type
func (a *Analyzer) analyzeSingleLink(link string, baseURL *url.URL) LinkResult {
	linkProcessor := NewLinkProcessor()
	return linkProcessor.ProcessLink(link, baseURL, a.isLinkAccessible)
}
