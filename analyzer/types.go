package analyzer

import (
	"sync"
	"time"
)

// AnalysisResult represents the result of analyzing a web page
type AnalysisResult struct {
	URL               string         `json:"url"`
	HTMLVersion       string         `json:"html_version"`
	PageTitle         string         `json:"page_title"`
	HeadingCounts     map[string]int `json:"heading_counts"`
	InternalLinks     int            `json:"internal_links"`
	ExternalLinks     int            `json:"external_links"`
	InaccessibleLinks int            `json:"inaccessible_links"`
	HasLoginForm      bool           `json:"has_login_form"`
	Error             *AnalysisError `json:"error,omitempty"`
	StatusCode        int            `json:"status_code,omitempty"`
}

// CacheEntry represents a cached analysis result
type CacheEntry struct {
	Result    *AnalysisResult
	Timestamp time.Time
	TTL       time.Duration
}

// LinkResult represents the result of analyzing a single link
type LinkResult struct {
	Link         string
	IsInternal   bool
	IsAccessible bool
	Error        error
}

// AnalysisJob represents a job for the worker pool
type AnalysisJob struct {
	Link    string
	BaseURL string
}

// AnalysisWorkerPool manages concurrent link analysis
type AnalysisWorkerPool struct {
	workers  int
	jobQueue chan AnalysisJob
	results  chan LinkResult
	stopChan chan struct{}
	analyzer *Analyzer
	workerWg sync.WaitGroup
}
