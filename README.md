# Chia log analyzer (Updated Fork)
This is a maintained fork of the Chia Log Analyzer, updated to work with the latest Chia versions and compiled as a standalone application.

Simply realtime chia log analyzer

![Screenshot](https://github.com/DannKenn/chia-log-analyzer/blob/main/v2.PNG)

## What's New
- **Standalone Executable**: No external dependencies required. Just download and run.
- **Updated Dependencies**: Built with modern Go libraries.
- **Bug Fixes**: improved stability and log parsing.

## Howto run on Linux
Download binary from the releases assets (`chia-log-analyzer.go-linux-amd64`)

You must set log level in your chia `.chia/mainnet/config/config.yaml` to `level: INFO`

Set binary as executable:
```bash
chmod +x chia-log-analyzer.go-linux-amd64
```

Run executable with path to `debug.log`:
```bash
./chia-log-analyzer.go-linux-amd64 --log=/path/to/debug.log
```

Or simply copy binary file to the directory with logs and run without parameters:
```bash
./chia-log-analyzer.go-linux-amd64
```

## Howto run on Windows
Download binary from the releases assets (`chia-log-analyzer.go-windows-amd64-signed.exe`)

You must set log level in your chia `C:\Users\<CurrentUserName>\.chia\mainnet\config\config.yaml` to `level: INFO`

Simply copy exe file to the directory with logs (`C:\Users\<CurrentUserName>\.chia\mainnet\log`) and run `chia-log-analyzer.go-windows-amd64-signed.exe`

Or run executable with path to `debug.log`:
```powershell
chia-log-analyzer.go-windows-amd64-signed.exe --log=C:\Users\<CurrentUserName>\.chia\mainnet\log\debug.log
```

## debug.log locations
Automatically trying to load `debug.log` from these locations:
- `./debug.log` (actual directory)
- get log path from the parameter `--log`
- `~/.chia/mainnet/log/debug.log` (default directory in home dir)

## Features
- monitoring of chia debug.log file
- simply show basic info about farming
- automatic refresh every 5s

## Supported platforms
- Linux (tested on Ubuntu) - download binary: `chia-log-analyzer.go-linux-amd64`
- RPI4 (use linux-arm builds) - download binary: `chia-log-analyzer.go-linux-arm`
- Windows10 - download binary: `chia-log-analyzer.go-windows-amd64-signed.exe`

## Keys
- `q` - exit

