#!/bin/bash
# Build script for Linux/macOS/Git Bash (alternative to Makefile)

echo "🔨 Building Ollama Code..."

# Create build directory
mkdir -p build

# Build the binary
go build -ldflags="-s -w" -trimpath -o build/ollama-code ./cmd/ollama-code

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Build complete: build/ollama-code"
    echo ""
    ls -lh build/ollama-code
else
    echo ""
    echo "❌ Build failed!"
    exit 1
fi
