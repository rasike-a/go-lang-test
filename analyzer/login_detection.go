package analyzer

import (
	"strings"

	"golang.org/x/net/html"
)

// isLoginForm checks if a form element represents a login form
func (a *Analyzer) isLoginForm(formNode *html.Node) bool {
	var hasPasswordField bool
	var hasUsernameField bool
	
	// Traverse all form inputs
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			var inputType, inputName, inputId, inputPlaceholder string
			
			// Extract input attributes
			for _, attr := range n.Attr {
				switch attr.Key {
				case "type":
					inputType = strings.ToLower(attr.Val)
				case "name":
					inputName = strings.ToLower(attr.Val)
				case "id":
					inputId = strings.ToLower(attr.Val)
				case "placeholder":
					inputPlaceholder = strings.ToLower(attr.Val)
				}
			}
			
			// Check for password field
			if inputType == "password" {
				hasPasswordField = true
			}
			
			// Check for username/email field
			if inputType == "text" || inputType == "email" || inputType == "tel" {
				// Check name attribute for common patterns
				if strings.Contains(inputName, "user") ||
				   strings.Contains(inputName, "login") ||
				   strings.Contains(inputName, "email") ||
				   strings.Contains(inputName, "account") ||
				   strings.Contains(inputName, "phone") {
					hasUsernameField = true
				}
				
				// Check ID attribute for common patterns
				if strings.Contains(inputId, "user") ||
				   strings.Contains(inputId, "login") ||
				   strings.Contains(inputId, "email") ||
				   strings.Contains(inputId, "account") ||
				   strings.Contains(inputId, "phone") {
					hasUsernameField = true
				}
				
				// Check placeholder for common patterns
				if strings.Contains(inputPlaceholder, "user") ||
				   strings.Contains(inputPlaceholder, "login") ||
				   strings.Contains(inputPlaceholder, "email") ||
				   strings.Contains(inputPlaceholder, "account") ||
				   strings.Contains(inputPlaceholder, "phone") {
					hasUsernameField = true
				}
			}
			
			// If we already found a password field, check if this input suggests it's a login form
			if hasPasswordField && inputType == "text" {
				// Additional check for common login-related patterns
				if strings.Contains(inputName, "user") ||
				   strings.Contains(inputName, "login") ||
				   strings.Contains(inputName, "email") ||
				   strings.Contains(inputName, "account") ||
				   strings.Contains(inputName, "phone") ||
				   strings.Contains(inputId, "user") ||
				   strings.Contains(inputId, "login") ||
				   strings.Contains(inputId, "email") ||
				   strings.Contains(inputId, "account") ||
				   strings.Contains(inputId, "phone") ||
				   strings.Contains(inputPlaceholder, "user") ||
				   strings.Contains(inputPlaceholder, "login") ||
				   strings.Contains(inputPlaceholder, "email") ||
				   strings.Contains(inputPlaceholder, "account") ||
				   strings.Contains(inputPlaceholder, "phone") {
					hasUsernameField = true
				}
			}
		}
		
		// Check for login-related button text
		if n.Type == html.ElementNode && (n.Data == "button" || n.Data == "input") {
			var buttonType, buttonValue string
			
			for _, attr := range n.Attr {
				switch attr.Key {
				case "type":
					buttonType = attr.Val
				case "value":
					buttonValue = strings.ToLower(attr.Val)
				}
			}
			
			// Check if button suggests login functionality
			if (buttonType == "submit" || buttonType == "button") &&
			   (strings.Contains(buttonValue, "login") ||
				strings.Contains(buttonValue, "sign in") ||
				strings.Contains(buttonValue, "log in") ||
				strings.Contains(buttonValue, "submit")) {
				// If we have a password field, this button text suggests it's a login form
				if hasPasswordField {
					hasUsernameField = true
				}
			}
		}
		
		// Continue traversing
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	
	traverse(formNode)
	
	// A form is considered a login form if it has both password and username fields
	return hasPasswordField && hasUsernameField
}
