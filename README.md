# Contents

- [Contents](#contents)
- [Introduction](#introduction)
- [Features](#features)
- [Setup](#setup)
  - [Installation](#installation)
  - [Usage](#usage)
- [Common Commands](#common-commands)

# Introduction

A real-time hardware monitoring application built in Go that displays CPU, memory, and disk usage. Demonstrates system programming with Go using the `gopsutil` library.

# Features

- Real-time monitoring of CPU usage percentage
- Memory usage display (percentage and GB format)
- Disk usage monitoring for C: drive
- Clean terminal interface with emojis
- Updates every second with live system stats
- Uses `gopsutil` library for cross-platform system information

# Setup

## Installation

### Install

- Windows Package Manager (Recommended):

  ```ps
  winget install GoLang.Go
  ```

- Manual Installation:

  - Visit [https://golang.org/dl/](https://golang.org/dl/)
  - Download the Windows installer (.msi)
  - Run the installer and follow the setup wizard

- Verify Installation:

  ```ps
  go version
  ```

### Update

To update Go to the latest version:

```ps
# Using Windows Package Manager
winget upgrade GoLang.Go

# Verify the update
go version
```

## Usage

### Clone and Run

1. **Clone the repository:**

   ```ps
   git clone https://github.com/Maraskalki/go-hw-monitor.git
   cd go-hw-monitor
   ```

2. **Download dependencies:**

   ```ps
   go mod download
   ```

3. **Run the application:**

   ```ps
   go run src\main.go
   ```

### Building

1. **Build executable:**

   ```ps
   # Create executable in build directory
   go build -o build\hw-monitor.exe .\src
   ```

2. **Run the executable:**

   ```ps
   .\build\hw-monitor.exe
   ```

### Testing

1. **Run all tests:**

   ```ps
   # Run all tests in the project
   go test ./src

   # Run tests with verbose output
   go test -v ./src
   ```

2. **Run specific tests:**

   ```ps
   # Run tests matching a pattern
   go test -v ./src -run TestGopsutil

   # Run a specific test function
   go test -v ./src -run TestGopsutilMonitorCPUUsage
   ```

3. **Test coverage:**

   ```ps
   # Run tests with coverage report
   go test -cover ./src

   # Generate detailed coverage report (use quotes in PowerShell)
   go test "-coverprofile=coverage.out" ./src

   # View coverage report in terminal
   go tool cover "-func=coverage.out"

   # View coverage report in browser
   go tool cover "-html=coverage.out"
   ```

4. **Test with build tags:**

   ```ps
   # Skip hardware tests in CI environments
   go test -short ./src

   # Run only quick tests
   go test -v ./src -short
   ```

### Dependency Management

#### Add a New Dependency

1. **Import in your Go code:**

   ```go
   import "github.com/example/package"
   ```

2. **Add dependency using go get:**

   ```ps
   go get github.com/example/package
   ```

3. **Or add specific version:**

   ```ps
   go get github.com/example/package@v1.2.3
   ```

#### Update Dependencies

```ps
# Update all dependencies to latest versions
go get -u ./...

# Update to patch versions only (safer)
go get -u=patch ./...

# Update specific package
go get -u github.com/example/package

# Clean up unused dependencies
go mod tidy
```

#### View Dependencies

```ps
# List all dependencies
go list -m all

# View dependency graph
go mod graph

# Show why a dependency is needed
go mod why github.com/example/package
```

# Common Commands

| Command                                        | Description              |
| ---------------------------------------------- | ------------------------ |
| `go run src\main.go`                           | Run the hardware monitor |
| `go build -o build\hw-monitor.exe src\main.go` | Compile the program      |
| `go mod init <name>`                           | Initialize a new module  |
| `go mod tidy`                                  | Clean up dependencies    |
| `go get <package>`                             | Add a dependency         |
| `go list -m all`                               | List all dependencies    |
| `go version`                                   | Check Go version         |
