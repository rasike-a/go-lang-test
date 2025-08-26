# Quick Start Guide

## Prerequisites
1. **Install Go 1.21+**: Download from [golang.org](https://golang.org/doc/install)
2. **Verify installation**: `go version`

## Running the Application

### Option 1: Using the startup script
```bash
./run.sh
```

### Option 2: Manual steps
```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Start the application
go run main.go
```

### Option 3: Using Make
```bash
make run
```

### Option 4: Using Docker
```bash
docker build -t web-page-analyzer .
docker run -p 8080:8080 web-page-analyzer
```

## Expected Output
```
2024/08/25 10:30:45 Server starting on port 8080
2024/08/25 10:30:45 Visit http://localhost:8080 to use the application
```

## Using the Application

1. **Open browser** to `http://localhost:8080`
2. **Enter a URL** to analyze (e.g., `https://example.com`)
3. **Click "Analyze Page"** to get results
4. **View detailed analysis** including:
   - HTML version
   - Page title
   - Heading breakdown (H1-H6)
   - Internal/external link counts
   - Login form detection
   - Accessibility issues

## Sample Analysis Results

When you analyze `https://example.com`, you'll get results like:
```json
{
  "url": "https://example.com",
  "html_version": "HTML5",
  "page_title": "Example Domain",
  "heading_counts": {
    "h1": 1
  },
  "internal_links": 1,
  "external_links": 1,
  "inaccessible_links": 0,
  "has_login_form": false,
  "status_code": 200
}
```

## Troubleshooting

### Port Already in Use
```bash
# Use different port
PORT=3000 go run main.go
```

### Connection Issues
- Check firewall settings
- Ensure port 8080 is available
- Try `http://127.0.0.1:8080` instead of localhost

### Go Not Found
Install Go from [golang.org](https://golang.org/doc/install) or:
```bash
# macOS
brew install go

# Ubuntu/Debian
sudo apt install golang-go

# Windows
# Download installer from golang.org
```

## Testing the API Directly

```bash
# Test with curl
curl -X POST http://localhost:8080/analyze \
  -d "url=https://example.com" \
  -H "Content-Type: application/x-www-form-urlencoded"
```