// Package main provides UI components and display logic for the hardware monitor.
// This file contains all terminal UI setup, layout, and rendering functions.
package main

import (
	"fmt"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// setupUI configures the initial layout of all UI components.
// It automatically detects terminal dimensions and delegates to setupUIWithSize.
func setupUI(cpuGauge, memoryGauge, diskGauge *widgets.Gauge, infoList *widgets.List) {
	// Get current terminal dimensions
	termWidth, termHeight := ui.TerminalDimensions()
	// Delegate to the more specific function with size parameters
	setupUIWithSize(cpuGauge, memoryGauge, diskGauge, infoList, termWidth, termHeight)
}

// setupUIWithSize configures the layout of UI components for specific dimensions.
// It creates a responsive 2x2 grid: 3 gauges on top, info panel on bottom.
// Coordinates use SetRect(x1, y1, x2, y2) where (0,0) is top-left.
func setupUIWithSize(cpuGauge, memoryGauge, diskGauge *widgets.Gauge, infoList *widgets.List, width, height int) {
	// COORDINATE SYSTEM: SetRect(x1, y1, x2, y2)
	// (0,0) is top-left corner, coordinates increase right and down
	// We're creating a 2x2 grid: 3 gauges on top, info panel on bottom

	// CPU Gauge - Left third of screen, top half
	cpuGauge.Title = "CPU Usage"
	cpuGauge.SetRect(0, 0, width/config.ScreenThirds, height/config.ScreenHalves) // Left third
	cpuGauge.BarColor = ui.ColorYellow                                            // Yellow bar (warning color)
	cpuGauge.BorderStyle.Fg = ui.ColorWhite                                       // White border
	cpuGauge.TitleStyle.Fg = ui.ColorCyan                                         // Cyan title

	// Memory Gauge - Middle third of screen, top half
	memoryGauge.Title = "Memory Usage"
	memoryGauge.SetRect(width/config.ScreenThirds, 0, 2*width/config.ScreenThirds, height/config.ScreenHalves) // Middle third
	memoryGauge.BarColor = ui.ColorGreen                                                                       // Green bar (safe color)
	memoryGauge.BorderStyle.Fg = ui.ColorWhite
	memoryGauge.TitleStyle.Fg = ui.ColorCyan

	// Disk Gauge - Right third of screen, top half
	diskGauge.Title = "Disk Usage"
	diskGauge.SetRect(2*width/config.ScreenThirds, 0, width, height/config.ScreenHalves) // Right third
	diskGauge.BarColor = ui.ColorRed                                                     // Red bar (danger color)
	diskGauge.BorderStyle.Fg = ui.ColorWhite
	diskGauge.TitleStyle.Fg = ui.ColorCyan

	// Info List - Full width, bottom half
	infoList.Title = "System Information"
	infoList.SetRect(0, height/config.ScreenHalves, width, height) // Full width, bottom half
	infoList.TextStyle = ui.NewStyle(ui.ColorWhite)
	infoList.WrapText = false // Don't wrap long lines
	infoList.BorderStyle.Fg = ui.ColorWhite
	infoList.TitleStyle.Fg = ui.ColorCyan
}

// updateDisplay fetches current system stats and updates all UI components.
// It uses concurrent data fetching for optimal performance and responsiveness.
func updateDisplay(cpuGauge, memoryGauge, diskGauge *widgets.Gauge, infoList *widgets.List, monitor SystemMonitor) {
	// CONCURRENT DATA FETCHING - Don't block the UI!
	// Create a channel to receive the complete system stats
	statsCh := make(chan SystemStats, config.ChannelBuffer) // Buffered channel
	// Start a goroutine to fetch all data concurrently
	go fetchSystemStats(monitor, statsCh) // This runs in the background

	// BLOCKING RECEIVE - Wait for the goroutine to send us data
	stats := <-statsCh // This blocks until data arrives

	// UPDATE GAUGES - Convert our data to visual elements
	// Gauges expect integer percentages (0-100)
	cpuGauge.Percent = int(stats.CPUUsage)                                       // Convert float to int
	cpuGauge.Label = fmt.Sprintf("%.*f%%", config.DecimalPlaces, stats.CPUUsage) // Format with configured precision

	memoryGauge.Percent = int(stats.MemoryUsage)
	memoryGauge.Label = fmt.Sprintf("%.*f%%", config.DecimalPlaces, stats.MemoryUsage)

	diskGauge.Percent = int(stats.DiskUsage)
	diskGauge.Label = fmt.Sprintf("%.*f%%", config.DecimalPlaces, stats.DiskUsage)

	// UPDATE INFO LIST - Create detailed text information
	// infoList.Rows is a slice of strings (like an array but dynamic)
	infoList.Rows = []string{
		fmt.Sprintf("Time: %s", time.Now().Format(config.TimeFormat)),
		"", // Empty line for spacing
		fmt.Sprintf("CPU: %.*f%%", config.DecimalPlaces, stats.CPUUsage),
		"",
		fmt.Sprintf("Memory: %.*f%% (%.*f GB / %.*f GB)",
			config.DecimalPlaces, stats.MemoryUsage, config.DecimalPlaces, stats.MemoryUsed, config.DecimalPlaces, stats.MemoryTotal),
		"",
		fmt.Sprintf("Disk (%s): %.*f%% (%.*f GB / %.*f GB)",
			config.DiskDrive, config.DecimalPlaces, stats.DiskUsage, config.DecimalPlaces, stats.DiskUsed, config.DecimalPlaces, stats.DiskTotal),
		"",
		"Press 'q' or Ctrl+C to quit", // User instruction
	}

	// RENDER - Actually draw everything to the screen
	// This is when the user sees the updated information
	ui.Render(cpuGauge, memoryGauge, diskGauge, infoList)
}

// createWidgets creates and returns all the UI widgets needed for the application.
// This is a factory function that centralizes widget creation.
func createWidgets() (*widgets.Gauge, *widgets.Gauge, *widgets.Gauge, *widgets.List) {
	// Create UI components (widgets) - these are like building blocks
	// widgets.NewGauge() returns a pointer to a new Gauge widget
	cpuGauge := widgets.NewGauge()    // Visual progress bar for CPU
	memoryGauge := widgets.NewGauge() // Visual progress bar for Memory
	diskGauge := widgets.NewGauge()   // Visual progress bar for Disk
	infoList := widgets.NewList()     // Text list for detailed information

	return cpuGauge, memoryGauge, diskGauge, infoList
}
