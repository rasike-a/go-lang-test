# Quick Start Guide

## Prerequisites
1. **Install Go 1.21+**: Download from [golang.org](https://golang.org/doc/install)
2. **Verify installation**: `go version`
3. **Optional**: Install Docker for containerized deployment

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

### Option 3: Using Make (Recommended)
```bash
# Development mode (go run)
make dev

# Build and run
make run

# Build Docker image
make docker-build

# Build and run with Docker
make docker-run
```

### Option 4: Using Docker directly
```bash
docker build -t web-page-analyzer .
docker run -p 8080:8080 web-page-analyzer
```

## Expected Output
```
{"level":"info","timestamp":"2025-09-01T09:51:56.087+0530","caller":"assignment/main.go:90","msg":"Server starting on port 8080"}
{"level":"info","timestamp":"2025-09-01T09:51:56.087+0530","caller":"assignment/main.go:91","msg":"Visit http://localhost:8080 to use the application"}
{"level":"info","timestamp":"2025-09-01T09:51:56.087+0530","caller":"assignment/main.go:92","msg":"Metrics available at http://localhost:8080/metrics"}
{"level":"info","timestamp":"2025-09-01T09:51:56.087+0530","caller":"assignment/main.go:94","msg":"Profiling available at http://localhost:8080/debug/pprof/"}
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
   - Inaccessible links count
   - Login form detection
   - Accessibility issues

## Available Endpoints

### Main Application
- `GET /` - Web interface
- `POST /analyze` - Analyze a web page

### Health & Monitoring
- `GET /health` - Health check endpoint
- `GET /metrics` - Application metrics (JSON format)
- `GET /debug/pprof/` - Go profiling endpoints

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
  "internal_links": 0,
  "external_links": 1,
  "inaccessible_links": 0,
  "has_login_form": false
}
```

## Testing the API Directly

```bash
# Test with curl
curl -X POST http://localhost:8080/analyze \
  -d "url=https://example.com" \
  -H "Content-Type: application/x-www-form-urlencoded"

# Check health status
curl http://localhost:8080/health

# Get metrics
curl http://localhost:8080/metrics
```

## Troubleshooting

### Port Already in Use
```bash
# Use different port
PORT=3000 go run main.go

# Or stop existing process
lsof -i :8080
kill <PID>
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

### Docker Issues
```bash
# Check if Docker is running
docker ps

# Build with no cache if needed
docker build --no-cache -t web-page-analyzer .
```

### Static Files Not Loading
If you see 404 errors for CSS/JS files:
- Ensure you're using the latest Docker image (includes static files)
- Check that the application is running locally (not in old Docker container)
- Verify static files exist in the `static/` directory

## Development

### Available Make Commands
```bash
make help          # Show all available commands
make build         # Build the binary
make dev           # Run in development mode
make test          # Run tests
make test-coverage # Run tests with coverage
make fmt           # Format code
make clean         # Clean build artifacts
make deps          # Download dependencies
```

### Project Structure
```
├── main.go              # Application entry point
├── handlers/            # HTTP handlers
├── analyzer/            # Web page analysis logic
├── middleware/          # HTTP middleware
├── logger/              # Logging configuration
├── static/              # Static files (CSS, JS)
│   ├── css/
│   └── js/
├── Dockerfile           # Docker configuration
├── Makefile            # Build automation
└── run.sh              # Startup script
```