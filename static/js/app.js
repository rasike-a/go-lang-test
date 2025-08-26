// Web Page Analyzer - Main JavaScript Application

document.addEventListener('DOMContentLoaded', function() {
    const analyzeForm = document.getElementById('analyzeForm');
    const urlInput = document.getElementById('url');
    const submitBtn = document.getElementById('submitBtn');
    const resultsDiv = document.getElementById('results');

    // Form submission handler
    analyzeForm.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const url = urlInput.value.trim();
        if (!url) {
            showError('Please enter a valid URL');
            return;
        }
        
        // Add loading state to button
        setButtonLoading(true);
        
        // Show results container with loading state
        showResults();
        showLoading();
        
        try {
            const formData = new FormData();
            formData.append('url', url);
            
            const response = await fetch('/analyze', {
                method: 'POST',
                body: formData
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const result = await response.json();
            displayResults(result);
        } catch (error) {
            console.error('Analysis error:', error);
            showError(`Error: Failed to analyze the page. ${error.message}`);
        } finally {
            // Reset button state
            setButtonLoading(false);
        }
    });

    // Set button loading state
    function setButtonLoading(loading) {
        if (loading) {
            submitBtn.disabled = true;
            submitBtn.classList.add('btn-loading');
            submitBtn.textContent = 'Analyzing...';
        } else {
            submitBtn.disabled = false;
            submitBtn.classList.remove('btn-loading');
            submitBtn.textContent = 'Analyze Page';
        }
    }

    // Show results container
    function showResults() {
        resultsDiv.style.display = 'block';
    }

    // Show loading state
    function showLoading() {
        resultsDiv.innerHTML = '<div class="loading">Analyzing web page, please wait...</div>';
    }

    // Show error message
    function showError(message) {
        resultsDiv.innerHTML = `<div class="error">${message}</div>`;
    }

    // Display analysis results
    function displayResults(result) {
        if (result.error) {
            let errorMsg = result.error;
            if (result.status_code) {
                errorMsg = `HTTP ${result.status_code}: ${result.error}`;
            }
            showError(errorMsg);
            return;
        }
        
        const headingsList = generateHeadingsList(result.heading_counts);
        
        resultsDiv.innerHTML = `
            <h2 class="results-header">Analysis Results</h2>
            <div class="result-item">
                <div class="result-label">URL</div>
                <div class="result-value">${result.url}</div>
            </div>
            <div class="result-item">
                <div class="result-label">HTML Version</div>
                <div class="result-value">${result.html_version}</div>
            </div>
            <div class="result-item">
                <div class="result-label">Page Title</div>
                <div class="result-value">${result.page_title || '<em>No title found</em>'}</div>
            </div>
            <div class="result-item">
                <div class="result-label">Headings</div>
                <div class="result-value">${headingsList}</div>
            </div>
            <div class="result-item">
                <div class="result-label">Links</div>
                <div class="result-value">
                    <strong>Internal:</strong> ${result.internal_links}<br>
                    <strong>External:</strong> ${result.external_links}<br>
                    <strong>Inaccessible:</strong> ${result.inaccessible_links}
                </div>
            </div>
            <div class="result-item">
                <div class="result-label">Login Form</div>
                <div class="result-value">${result.has_login_form ? 'Yes' : 'No'}</div>
            </div>
        `;
    }

    // Generate headings list HTML
    function generateHeadingsList(headingCounts) {
        if (!headingCounts || Object.keys(headingCounts).length === 0) {
            return '<em>No headings found</em>';
        }
        
        let headingsList = '<ul class="headings-list">';
        for (const [level, count] of Object.entries(headingCounts)) {
            headingsList += `<li><strong>${level.toUpperCase()}:</strong> ${count}</li>`;
        }
        headingsList += '</ul>';
        
        return headingsList;
    }

    // Add some interactive enhancements
    urlInput.addEventListener('focus', function() {
        this.parentElement.classList.add('focused');
    });

    urlInput.addEventListener('blur', function() {
        this.parentElement.classList.remove('focused');
    });

    // Add keyboard shortcuts
    document.addEventListener('keydown', function(e) {
        // Ctrl/Cmd + Enter to submit form
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            e.preventDefault();
            analyzeForm.dispatchEvent(new Event('submit'));
        }
        
        // Escape to clear form
        if (e.key === 'Escape') {
            urlInput.value = '';
            resultsDiv.style.display = 'none';
            urlInput.focus();
        }
    });
});
