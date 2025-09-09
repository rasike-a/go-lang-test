package analyzer

import (
	"context"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// analyzeDocument analyzes the HTML document and populates the result
func (a *Analyzer) analyzeDocument(doc *html.Node, result *AnalysisResult, baseURL *url.URL, htmlContent string) {
	// Detect HTML version
	result.HTMLVersion = a.detectHTMLVersion(htmlContent)

	// Extract page title
	result.PageTitle = a.extractPageTitle(doc)

	// Count headings
	result.HeadingCounts = a.countHeadings(doc)

	// Extract and analyze links
	links := a.extractLinks(doc)
	a.analyzeLinksConcurrent(links, baseURL, result)

	// Check for login forms
	result.HasLoginForm = a.hasLoginForm(doc)
}

// analyzeDocumentWithContext analyzes the HTML document with context support
func (a *Analyzer) analyzeDocumentWithContext(ctx context.Context, doc *html.Node, result *AnalysisResult, baseURL *url.URL, htmlContent string) {
	// Create a child context with a shorter timeout for HTML analysis
	analysisCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Check if context is cancelled before starting analysis
	select {
	case <-analysisCtx.Done():
		result.Error = NewTimeoutError("HTML analysis timeout", 10*time.Second)
		return
	default:
	}

	// Perform the analysis
	a.analyzeDocument(doc, result, baseURL, htmlContent)
}

// detectHTMLVersion detects the HTML version from the document content
func (a *Analyzer) detectHTMLVersion(htmlContent string) string {
	// Convert to lowercase for case-insensitive matching
	content := strings.ToLower(htmlContent)

	// Check for DOCTYPE declarations
	if strings.Contains(content, "<!doctype html>") {
		return "HTML5"
	}

	if strings.Contains(content, "<!doctype html public") {
		// Check for specific versions
		if strings.Contains(content, "xhtml 1.0 strict") {
			return "XHTML 1.0 Strict"
		}
		if strings.Contains(content, "xhtml 1.0") {
			return "XHTML 1.0"
		}
		if strings.Contains(content, "xhtml 1.1") {
			return "XHTML 1.1"
		}
		if strings.Contains(content, "html 4.01 transitional") {
			return "HTML 4.01 Transitional"
		}
		if strings.Contains(content, "html 4.01") {
			return "HTML 4.01"
		}
		if strings.Contains(content, "html 4.0") {
			return "HTML 4.0"
		}
		if strings.Contains(content, "html 3.2") {
			return "HTML 3.2"
		}
		if strings.Contains(content, "html 2.0") {
			return "HTML 2.0"
		}
		return "HTML 4.01" // Default for generic HTML public DOCTYPE
	}

	// If no DOCTYPE found, try to infer from HTML structure
	if strings.Contains(content, "<html") {
		// Only assume HTML5 if there's a DOCTYPE
		// Documents without DOCTYPE are "Unknown"
		return "Unknown"
	}

	return "Unknown"
}

// extractPageTitle extracts the page title from the HTML document
func (a *Analyzer) extractPageTitle(doc *html.Node) string {
	var title string
	traverser := NewHTMLTraverser()

	traverser.TraverseElements(doc, "title", func(n *html.Node) {
		if n.FirstChild != nil {
			title = strings.TrimSpace(n.FirstChild.Data)
		}
	})

	return title
}

// countHeadings counts the occurrences of each heading level
func (a *Analyzer) countHeadings(doc *html.Node) map[string]int {
	headings := make(map[string]int)
	traverser := NewHTMLTraverser()

	traverser.TraverseAllElements(doc, func(n *html.Node) {
		if strings.HasPrefix(n.Data, "h") && len(n.Data) == 2 {
			level := n.Data[1:]
			if level >= "1" && level <= "6" {
				headings["h"+level]++
			}
		}
	})

	return headings
}

// extractLinks extracts all links from the HTML document
func (a *Analyzer) extractLinks(doc *html.Node) []string {
	var links []string
	traverser := NewHTMLTraverser()

	traverser.TraverseElements(doc, "a", func(n *html.Node) {
		href := traverser.GetAttributeValue(n, "href")
		if href != "" && !strings.HasPrefix(href, "#") {
			links = append(links, href)
		}
	})

	return links
}

// hasLoginForm checks if the document contains a login form
func (a *Analyzer) hasLoginForm(doc *html.Node) bool {
	var hasLoginForm bool
	traverser := NewHTMLTraverser()

	traverser.TraverseElements(doc, "form", func(n *html.Node) {
		if a.isLoginForm(n) {
			hasLoginForm = true
		}
	})

	return hasLoginForm
}
