#!/bin/bash

# Web Page Analyzer Startup Script
# Make sure Go 1.21+ is installed before running

# Default configuration
DEFAULT_PORT=8080
DEFAULT_ENV=development

# Parse command line arguments
PORT=${PORT:-$DEFAULT_PORT}
ENV=${ENV:-$DEFAULT_ENV}

echo "🚀 Starting Web Page Analyzer..."
echo "📋 Checking Go installation..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first:"
    echo "   - Visit: https://golang.org/doc/install"
    echo "   - Or use package manager: brew install go (macOS) / apt install golang (Ubuntu)"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go version $GO_VERSION is installed, but Go $REQUIRED_VERSION+ is required"
    exit 1
fi

echo "✅ Go version: $(go version)"
echo "🌐 Port: $PORT"
echo "🏷️  Environment: $ENV"

# Check if port is already in use
if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo "⚠️  Port $PORT is already in use. Attempting to stop existing process..."
    PID=$(lsof -ti:$PORT)
    if [ ! -z "$PID" ]; then
        kill $PID 2>/dev/null
        sleep 2
        if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
            echo "❌ Could not free port $PORT. Please stop the process manually or use a different port:"
            echo "   PORT=3000 ./run.sh"
            exit 1
        fi
    fi
fi

echo "📦 Installing dependencies..."
if ! go mod download; then
    echo "❌ Failed to download dependencies"
    exit 1
fi

echo "🧪 Running tests..."
if ! go test ./...; then
    echo "❌ Tests failed. Please fix issues before running."
    exit 1
fi

echo "✅ All tests passed!"
echo "🌐 Starting server on http://localhost:$PORT"
echo "   Environment: $ENV"
echo "   Press Ctrl+C to stop"
echo ""

# Export environment variables for the application
export PORT=$PORT
export ENV=$ENV

# Start the application
go run main.go