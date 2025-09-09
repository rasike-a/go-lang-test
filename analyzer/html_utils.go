package analyzer

import (
	"strings"

	"golang.org/x/net/html"
)

// HTMLTraverser provides common HTML traversal functionality
type HTMLTraverser struct{}

// NewHTMLTraverser creates a new HTML traverser
func NewHTMLTraverser() *HTMLTraverser {
	return &HTMLTraverser{}
}

// TraverseElements traverses HTML nodes and calls the provided function for each element
func (ht *HTMLTraverser) TraverseElements(doc *html.Node, elementName string, fn func(*html.Node)) {
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == elementName {
			fn(n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
}

// TraverseAllElements traverses HTML nodes and calls the provided function for each element
func (ht *HTMLTraverser) TraverseAllElements(doc *html.Node, fn func(*html.Node)) {
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			fn(n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
}

// GetAttributeValue extracts the value of a specific attribute from an HTML node
func (ht *HTMLTraverser) GetAttributeValue(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}

// GetMultipleAttributeValues extracts multiple attribute values from an HTML node
func (ht *HTMLTraverser) GetMultipleAttributeValues(node *html.Node, attrNames []string) map[string]string {
	values := make(map[string]string)
	for _, attr := range node.Attr {
		for _, name := range attrNames {
			if attr.Key == name {
				values[name] = strings.TrimSpace(attr.Val)
			}
		}
	}
	return values
}

// IsElement checks if a node is an element with the specified name
func (ht *HTMLTraverser) IsElement(node *html.Node, elementName string) bool {
	return node.Type == html.ElementNode && node.Data == elementName
}

// HasAttribute checks if a node has a specific attribute
func (ht *HTMLTraverser) HasAttribute(node *html.Node, attrName string) bool {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return true
		}
	}
	return false
}
