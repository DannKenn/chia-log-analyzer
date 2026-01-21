@echo off
echo Building for Windows (AMD64)...
set GOOS=windows
set GOARCH=amd64
go build -o chia-log-analyzer.go-windows-amd64.exe chia-log-analyzer.go

echo Building for Linux (AMD64)...
set GOOS=linux
set GOARCH=amd64
go build -o chia-log-analyzer.go-linux-amd64 chia-log-analyzer.go

echo Building for Linux (ARM)...
set GOOS=linux
set GOARCH=arm
go build -o chia-log-analyzer.go-linux-arm chia-log-analyzer.go

echo Build complete!
pause
