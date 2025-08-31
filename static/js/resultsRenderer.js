/**
 * ResultsRenderer - Handles rendering of analysis results using HTML templates
 */
class ResultsRenderer {
    constructor(container) {
        this.container = container;
        this.templates = this.loadTemplates();
    }
    
    /**
     * Load all HTML templates from the DOM
     */
    loadTemplates() {
        return {
            results: document.getElementById('resultsTemplate'),
            headings: document.getElementById('headingsTemplate'),
            loading: document.getElementById('loadingTemplate'),
            error: document.getElementById('errorTemplate')
        };
    }
    
    /**
     * Render the main analysis results
     */
    renderResults(result) {
        if (result.error) {
            this.renderError(result.error);
            return;
        }
        
        this.container.innerHTML = '';
        const clone = this.templates.results.content.cloneNode(true);
        
        // Populate all fields
        this.populateFields(clone, result);
        this.renderHeadings(clone, result.headings);
        this.renderLinks(clone, result);
        
        this.container.appendChild(clone);
        this.container.dataset.state = 'success';
    }
    
    /**
     * Populate data fields in the template
     */
    populateFields(element, data) {
        element.querySelectorAll('[data-field]').forEach(field => {
            const key = field.dataset.field;
            
            // Map template field names to API response field names
            let value;
            if (key === 'login_form') {
                value = data.has_login_form;
            } else if (key === 'html_version') {
                value = data.html_version;
            } else if (key === 'page_title') {
                value = data.page_title;
            } else if (key === 'headings') {
                value = data.headings;
            } else if (key === 'url') {
                value = data.url;
            } else {
                value = data[key];
            }
            
            // Handle special cases
            if (key === 'page_title') {
                field.textContent = value || 'No title found';
            } else if (key === 'login_form') {
                field.textContent = value ? 'Yes' : 'No';
            } else if (key === 'headings') {
                // Headings are handled separately by renderHeadings
                return;
            } else if (key === 'links') {
                // Links are handled separately by renderLinks
                return;
            } else {
                field.textContent = value || 'N/A';
            }
        });
    }
    
    /**
     * Render headings section with proper formatting
     */
    renderHeadings(element, headingCounts) {
        const headingsField = element.querySelector('[data-field="headings"]');
        if (!headingsField) return;
        
        if (!headingCounts || Object.keys(headingCounts).length === 0) {
            headingsField.innerHTML = '<em>No headings found</em>';
            return;
        }
        
        const headingsList = document.createElement('ul');
        headingsList.className = 'headings-list';
        
        for (const [level, count] of Object.entries(headingCounts)) {
            const li = document.createElement('li');
            li.innerHTML = `<strong>${level.toUpperCase()}:</strong> ${count}`;
            headingsList.appendChild(li);
        }
        
        headingsField.innerHTML = '';
        headingsField.appendChild(headingsList);
    }
    
    /**
     * Render links section with proper formatting
     */
    renderLinks(element, data) {
        const linksField = element.querySelector('[data-field="links"]');
        if (!linksField) return;
        
        linksField.innerHTML = `
            <strong>Internal:</strong> ${data.internal_links}<br>
            <strong>External:</strong> ${data.external_links}<br>
            <strong>Inaccessible:</strong> ${data.inaccessible_links}
        `;
    }
    
    /**
     * Render loading state
     */
    renderLoading() {
        this.container.innerHTML = '';
        const clone = this.templates.loading.content.cloneNode(true);
        this.container.appendChild(clone);
        this.container.dataset.state = 'loading';
    }
    
    /**
     * Render error state
     */
    renderError(error) {
        this.container.innerHTML = '';
        const clone = this.templates.error.content.cloneNode(true);
        
        // Set error message
        const messageField = clone.querySelector('[data-field="message"]');
        if (messageField) {
            messageField.textContent = error;
        }
        
        this.container.appendChild(clone);
        this.container.dataset.state = 'error';
    }
    
    /**
     * Clear results and reset state
     */
    clear() {
        this.container.innerHTML = '';
        this.container.removeAttribute('data-state');
    }
    
    /**
     * Show results container
     */
    show() {
        this.container.style.display = 'block';
    }
    
    /**
     * Hide results container
     */
    hide() {
        this.container.style.display = 'none';
    }
}
