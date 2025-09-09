package analyzer

import (
	"net/url"
	"strings"
)

// LinkProcessor provides common link processing functionality
type LinkProcessor struct{}

// NewLinkProcessor creates a new link processor
func NewLinkProcessor() *LinkProcessor {
	return &LinkProcessor{}
}

// ProcessLink processes a single link and returns the result
func (lp *LinkProcessor) ProcessLink(link string, baseURL *url.URL, isAccessibleChecker func(string) bool) LinkResult {
	// Skip empty links and fragments
	if link == "" || strings.HasPrefix(link, "#") {
		return LinkResult{
			Link:         link,
			IsInternal:   false,
			IsAccessible: false,
			Error:        nil,
		}
	}

	// Parse the link URL
	linkURL, err := url.Parse(link)
	if err != nil {
		return LinkResult{
			Link:         link,
			IsInternal:   false,
			IsAccessible: false,
			Error:        err,
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
		isAccessible = isAccessibleChecker(linkURL.String())
	} else {
		isAccessible = true // Assume internal links are accessible
	}

	return LinkResult{
		Link:         link,
		IsInternal:   isInternal,
		IsAccessible: isAccessible,
		Error:        nil,
	}
}

// IsSpecialProtocol checks if a link uses a special protocol that should be skipped
func (lp *LinkProcessor) IsSpecialProtocol(link string) bool {
	specialProtocols := []string{
		"javascript:",
		"mailto:",
		"tel:",
		"ftp:",
		"file:",
		"data:",
		"blob:",
		"chrome:",
		"moz-extension:",
	}

	for _, protocol := range specialProtocols {
		if strings.HasPrefix(link, protocol) {
			return true
		}
	}
	return false
}

// CreateErrorLinkResult creates a LinkResult with an error
func (lp *LinkProcessor) CreateErrorLinkResult(link string, err error) LinkResult {
	return LinkResult{
		Link:         link,
		IsInternal:   false,
		IsAccessible: false,
		Error:        err,
	}
}
