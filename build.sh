#!/bin/bash

# Build for Linux
echo "Building for Linux..."
go build -o chia-log-analyzer.go-linux-amd64 chia-log-analyzer.go

# Build for Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o chia-log-analyzer.go-windows-amd64.exe chia-log-analyzer.go

echo "Build done."
