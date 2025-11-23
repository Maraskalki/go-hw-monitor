package main

import "time"

// Config holds all configuration constants for the hardware monitor.
// This includes both user-configurable settings and application constants.
var Config = struct {
	// Display settings
	RefreshInterval time.Duration
	TimeFormat      string
	Title           string
	Separator       string

	// Precision
	DecimalPlaces int

	// System settings
	DiskDrive         string
	CPUSampleDuration time.Duration

	// Universal constants - these don't change across configurations
	BytesToGB     int64 // Convert bytes to gigabytes (1024³)
	ScreenThirds  int   // Divide screen into thirds for layout
	ScreenHalves  int   // Divide screen into halves for layout
	MetricCount   int   // Number of metrics we collect (CPU, Memory, Disk)
	ChannelBuffer int   // Buffer size for stats channel
	ResultsBuffer int   // Buffer size for results channel
}{
	// Refresh the display every second
	RefreshInterval: 1 * time.Second,

	// Time format (24-hour format HH:MM:SS)
	TimeFormat: "15:04:05",

	// Display text
	Title:     "Hardware Monitor - Press Ctrl+C to stop",
	Separator: "=========================================",

	// Number formatting
	DecimalPlaces: 1,

	// System monitoring settings
	DiskDrive:         "C:",
	CPUSampleDuration: 100 * time.Millisecond,

	// Universal constants - initialized once
	BytesToGB:     1024 * 1024 * 1024, // 1024³
	ScreenThirds:  3,
	ScreenHalves:  2,
	MetricCount:   3,
	ChannelBuffer: 1,
	ResultsBuffer: 3,
}
