/**
 * Web Page Analyzer - Main JavaScript Application
 * Improved version with better separation of concerns
 */

document.addEventListener('DOMContentLoaded', function() {
    // Initialize components
    const analyzeForm = document.getElementById('analyzeForm');
    const urlInput = document.getElementById('url');
    const submitBtn = document.getElementById('submitBtn');
    const resultsDiv = document.getElementById('results');
    
    // Initialize the results renderer
    const resultsRenderer = new ResultsRenderer(resultsDiv);

    // Form submission handler
    analyzeForm.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const url = urlInput.value.trim();
        if (!url) {
            resultsRenderer.renderError('Please enter a valid URL');
            return;
        }
        
        // Add loading state to button
        setButtonLoading(true);
        
        // Show results container with loading state
        resultsRenderer.show();
        resultsRenderer.renderLoading();
        
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
            resultsRenderer.renderResults(result);
        } catch (error) {
            console.error('Analysis error:', error);
            resultsRenderer.renderError(`Error: Failed to analyze the page. ${error.message}`);
        } finally {
            // Reset button state
            setButtonLoading(false);
        }
    });

    /**
     * Set button loading state using data attributes
     */
    function setButtonLoading(loading) {
        const loadingText = submitBtn.dataset.loadingText || 'Analyzing...';
        const defaultText = submitBtn.dataset.defaultText || 'Analyze Page';
        
        if (loading) {
            submitBtn.disabled = true;
            submitBtn.classList.add('btn-loading');
            submitBtn.textContent = loadingText;
        } else {
            submitBtn.disabled = false;
            submitBtn.classList.remove('btn-loading');
            submitBtn.textContent = defaultText;
        }
    }

    /**
     * Add interactive enhancements
     */
    urlInput.addEventListener('focus', function() {
        this.parentElement.classList.add('focused');
    });

    urlInput.addEventListener('blur', function() {
        this.parentElement.classList.remove('focused');
    });

    /**
     * Add keyboard shortcuts
     */
    document.addEventListener('keydown', function(e) {
        // Ctrl/Cmd + Enter to submit form
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            e.preventDefault();
            analyzeForm.dispatchEvent(new Event('submit'));
        }
        
        // Escape to clear form
        if (e.key === 'Escape') {
            urlInput.value = '';
            resultsRenderer.hide();
            urlInput.focus();
        }
    });
});
