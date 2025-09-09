#!/bin/bash

# Installation script for Go and gRPC dependencies

echo "=== Installing Go and gRPC Dependencies ==="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go first:"
    echo "1. Visit https://golang.org/dl/"
    echo "2. Download Go for macOS"
    echo "3. Install the package"
    echo "4. Add Go to your PATH"
    echo ""
    echo "Or use Homebrew:"
    echo "  brew install go"
    exit 1
fi

echo "Go version: $(go version)"

# Install protobuf compiler
if ! command -v protoc &> /dev/null; then
    echo "Installing protobuf compiler..."
    if command -v brew &> /dev/null; then
        brew install protobuf
    else
        echo "Please install protobuf manually:"
        echo "1. Visit https://github.com/protocolbuffers/protobuf/releases"
        echo "2. Download protoc for macOS"
        echo "3. Extract and add to PATH"
    fi
fi

# Install Go protobuf plugins
echo "Installing Go protobuf plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Add Go bin to PATH if not already there
export PATH=$PATH:$(go env GOPATH)/bin

echo "=== Installation Complete ==="
echo "You can now run:"
echo "  make -f Makefile.grpc all"
echo "  make -f Makefile.grpc test"
