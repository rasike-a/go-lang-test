#!/bin/bash

# Web Page Analyzer Startup Script
# Make sure Go 1.21+ is installed before running

echo "🚀 Starting Web Page Analyzer..."
echo "📋 Checking Go installation..."

if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first:"
    echo "   - Visit: https://golang.org/doc/install"
    echo "   - Or use package manager: brew install go (macOS) / apt install golang (Ubuntu)"
    exit 1
fi

echo "✅ Go version: $(go version)"
echo "📦 Installing dependencies..."
go mod download

echo "🧪 Running tests..."
go test ./...

if [ $? -eq 0 ]; then
    echo "✅ All tests passed!"
    echo "🌐 Starting server on http://localhost:8080"
    echo "   Press Ctrl+C to stop"
    echo ""
    go run main.go
else
    echo "❌ Tests failed. Please fix issues before running."
    exit 1
fi