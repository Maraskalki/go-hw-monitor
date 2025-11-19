package main

import "time"

// Config holds all configuration constants for the hardware monitor
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
}{
	// Refresh the display every 5 seconds
	RefreshInterval: 5 * time.Second,

	// Time format (24-hour format HH:MM:SS)
	TimeFormat: "15:04:05",

	// Display text
	Title:     "Hardware Monitor - Press Ctrl+C to stop",
	Separator: "=========================================",

	// Number formatting
	DecimalPlaces: 1,

	// System monitoring settings
	DiskDrive:         "C:",
	CPUSampleDuration: 1 * time.Second,
}
