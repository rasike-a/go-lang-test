package analyzer

import (
	"strings"

	"golang.org/x/net/html"
)

// isLoginForm checks if a form element represents a login form
func (a *Analyzer) isLoginForm(formNode *html.Node) bool {
	var hasPasswordField bool
	var hasUsernameField bool
	traverser := NewHTMLTraverser()

	// Traverse all form inputs
	traverser.TraverseElements(formNode, "input", func(n *html.Node) {
		// Extract input attributes
		attrs := traverser.GetMultipleAttributeValues(n, []string{"type", "name", "id", "placeholder"})
		inputType := strings.ToLower(attrs["type"])
		inputName := strings.ToLower(attrs["name"])
		inputId := strings.ToLower(attrs["id"])
		inputPlaceholder := strings.ToLower(attrs["placeholder"])

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
	})

	// Check for login-related button text
	traverser.TraverseAllElements(formNode, func(n *html.Node) {
		if traverser.IsElement(n, "button") || traverser.IsElement(n, "input") {
			buttonAttrs := traverser.GetMultipleAttributeValues(n, []string{"type", "value"})
			buttonType := buttonAttrs["type"]
			buttonValue := strings.ToLower(buttonAttrs["value"])

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
	})

	// A form is considered a login form if it has both password and username fields
	return hasPasswordField && hasUsernameField
}
