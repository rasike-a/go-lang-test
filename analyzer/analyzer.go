package analyzer

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type AnalysisResult struct {
	URL               string            `json:"url"`
	HTMLVersion       string            `json:"html_version"`
	PageTitle         string            `json:"page_title"`
	HeadingCounts     map[string]int    `json:"heading_counts"`
	InternalLinks     int               `json:"internal_links"`
	ExternalLinks     int               `json:"external_links"`
	InaccessibleLinks int               `json:"inaccessible_links"`
	HasLoginForm      bool              `json:"has_login_form"`
	Error             string            `json:"error,omitempty"`
	StatusCode        int               `json:"status_code,omitempty"`
}

type Analyzer struct {
	httpClient *http.Client
	timeout    time.Duration
}

func NewAnalyzer(timeout time.Duration) *Analyzer {
	return &Analyzer{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

func (a *Analyzer) AnalyzeURL(targetURL string) *AnalysisResult {
	result := &AnalysisResult{
		URL:           targetURL,
		HeadingCounts: make(map[string]int),
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		result.Error = fmt.Sprintf("Invalid URL: %v", err)
		return result
	}

	if parsedURL.Scheme == "" {
		targetURL = "https://" + targetURL
		parsedURL, err = url.Parse(targetURL)
		if err != nil {
			result.Error = fmt.Sprintf("Invalid URL: %v", err)
			return result
		}
	}

	resp, err := a.httpClient.Get(targetURL)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to fetch URL: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	if resp.StatusCode >= 400 {
		result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read response body: %v", err)
		return result
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		result.Error = fmt.Sprintf("Failed to parse HTML: %v", err)
		return result
	}

	a.analyzeDocument(doc, result, parsedURL, string(body))
	return result
}

func (a *Analyzer) analyzeDocument(doc *html.Node, result *AnalysisResult, baseURL *url.URL, htmlContent string) {
	result.HTMLVersion = a.detectHTMLVersion(htmlContent)

	var traverse func(*html.Node)
	var links []string

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					result.PageTitle = strings.TrimSpace(n.FirstChild.Data)
				}
			case "h1", "h2", "h3", "h4", "h5", "h6":
				result.HeadingCounts[n.Data]++
			case "a":
				for _, attr := range n.Attr {
					if attr.Key == "href" && attr.Val != "" {
						links = append(links, attr.Val)
					}
				}
			case "form":
				if a.isLoginForm(n) {
					result.HasLoginForm = true
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	a.analyzeLinks(links, baseURL, result)
}

func (a *Analyzer) detectHTMLVersion(htmlContent string) string {
	htmlContent = strings.TrimSpace(strings.ToLower(htmlContent))
	
	if strings.Contains(htmlContent, "<!doctype html>") {
		return "HTML5"
	}
	
	if strings.Contains(htmlContent, `"-//w3c//dtd xhtml 1.1//en"`) {
		return "XHTML 1.1"
	}
	if strings.Contains(htmlContent, `"-//w3c//dtd xhtml 1.0 strict//en"`) {
		return "XHTML 1.0 Strict"
	}
	if strings.Contains(htmlContent, `"-//w3c//dtd xhtml 1.0 transitional//en"`) {
		return "XHTML 1.0 Transitional"
	}
	if strings.Contains(htmlContent, `"-//w3c//dtd xhtml 1.0 frameset//en"`) {
		return "XHTML 1.0 Frameset"
	}
	
	if strings.Contains(htmlContent, `"-//w3c//dtd html 4.01//en"`) {
		return "HTML 4.01 Strict"
	}
	if strings.Contains(htmlContent, `"-//w3c//dtd html 4.01 transitional//en"`) {
		return "HTML 4.01 Transitional"
	}
	if strings.Contains(htmlContent, `"-//w3c//dtd html 4.01 frameset//en"`) {
		return "HTML 4.01 Frameset"
	}
	
	if strings.Contains(htmlContent, `"-//w3c//dtd html 3.2 final//en"`) {
		return "HTML 3.2"
	}
	
	if strings.Contains(htmlContent, `"-//ietf//dtd html 2.0//en"`) {
		return "HTML 2.0"
	}
	
	return "Unknown"
}

func (a *Analyzer) analyzeLinks(links []string, baseURL *url.URL, result *AnalysisResult) {
	for _, link := range links {
		link = strings.TrimSpace(link)
		if link == "" || strings.HasPrefix(link, "#") || strings.HasPrefix(link, "javascript:") || strings.HasPrefix(link, "mailto:") || strings.HasPrefix(link, "tel:") {
			continue
		}

		parsedLink, err := url.Parse(link)
		if err != nil {
			continue
		}

		resolvedLink := baseURL.ResolveReference(parsedLink)
		
		if resolvedLink.Host == baseURL.Host {
			result.InternalLinks++
		} else {
			result.ExternalLinks++
		}

		if !a.isLinkAccessible(resolvedLink.String()) {
			result.InaccessibleLinks++
		}
	}
}

func (a *Analyzer) isLinkAccessible(link string) bool {
	req, err := http.NewRequest("HEAD", link, nil)
	if err != nil {
		return false
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode < 400
}

func (a *Analyzer) isLoginForm(formNode *html.Node) bool {
	hasPasswordField := false
	hasUsernameField := false
	
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			var inputType, inputName string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "type":
					inputType = strings.ToLower(attr.Val)
				case "name":
					inputName = strings.ToLower(attr.Val)
				}
			}
			
			if inputType == "password" {
				hasPasswordField = true
			}
			
			if inputType == "email" || inputType == "text" || inputType == "" {
				if strings.Contains(inputName, "user") || 
				   strings.Contains(inputName, "login") || 
				   strings.Contains(inputName, "email") {
					hasUsernameField = true
				}
			}
		}
		
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	
	traverse(formNode)
	return hasPasswordField && hasUsernameField
}