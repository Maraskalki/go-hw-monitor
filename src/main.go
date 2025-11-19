// Package main provides a hardware monitoring application with terminal UI.
// It demonstrates Go concurrency, system programming, and real-time data visualization.
package main

import (
	"fmt"
	"log"
	"time"

	// Import with alias - 'ui' is shorter than 'termui'
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

// SystemStats holds real-time system monitoring data.
// It groups related hardware metrics for easy handling and display.
type SystemStats struct {
	CPUUsage    float64 // CPU percentage (0-100)
	MemoryUsage float64 // Memory percentage (0-100)
	MemoryUsed  float64 // Memory used in GB
	MemoryTotal float64 // Total memory in GB
	DiskUsage   float64 // Disk percentage (0-100)
	DiskUsed    float64 // Disk used in GB
	DiskTotal   float64 // Total disk space in GB
}

// Global configuration - accessible from anywhere in this package
var config = Config

// main - Entry point of our program
// This function demonstrates: error handling, defer, UI setup, event loops
func main() {
	// Initialize termui - sets up the terminal for drawing
	// This can fail if terminal doesn't support required features
	if err := ui.Init(); err != nil {
		// log.Fatalf prints error message and exits program with error code
		log.Fatalf("failed to initialize termui: %v", err)
	}
	// defer ensures ui.Close() runs when main() exits (even if it panics!)
	// This is Go's way of guaranteed cleanup - like "finally" in other languages
	defer ui.Close()

	// Create UI components (widgets) - these are like building blocks
	// widgets.NewGauge() returns a pointer to a new Gauge widget
	cpuGauge := widgets.NewGauge()    // Visual progress bar for CPU
	memoryGauge := widgets.NewGauge() // Visual progress bar for Memory
	diskGauge := widgets.NewGauge()   // Visual progress bar for Disk
	infoList := widgets.NewList()     // Text list for detailed information

	// Setup UI layout - position and style all widgets
	setupUI(cpuGauge, memoryGauge, diskGauge, infoList)

	// Create ticker for periodic updates - like a timer that "ticks" every interval
	// time.Ticker sends current time on its channel at regular intervals
	ticker := time.NewTicker(config.RefreshInterval)
	defer ticker.Stop() // Always cleanup resources when function exits

	// ui.PollEvents() returns a channel that receives user input events
	// Events include: keyboard presses, mouse clicks, window resize, etc.
	uiEvents := ui.PollEvents()

	// Do initial update to show data immediately (don't wait for first tick)
	updateDisplay(cpuGauge, memoryGauge, diskGauge, infoList)

	// MAIN EVENT LOOP - This is the heart of the application!
	// This infinite loop waits for events and handles them
	// Think of it like: "Wait for either user input OR timer tick, then react"
	for {
		// select statement - Go's powerful concurrency primitive
		// Like a traffic controller: waits for data on multiple channels
		// Executes the FIRST case that receives data
		// Blocks (waits) until something happens
		select {
		// Case 1: User input event received
		case e := <-uiEvents:
			// Check what type of event occurred
			switch e.ID {
			case "q", "<C-c>": // User pressed 'q' or Ctrl+C
				return // Exit the program cleanly
			case "<Resize>": // User resized the terminal window
				// Type assertion: convert generic payload to specific type
				// This says "I know this is a Resize event, give me the data"
				payload := e.Payload.(ui.Resize)
				// Recalculate layout for new window size
				setupUIWithSize(cpuGauge, memoryGauge, diskGauge, infoList, payload.Width, payload.Height)
				// Redraw everything with new layout
				ui.Render(cpuGauge, memoryGauge, diskGauge, infoList)
			}
		// Case 2: Timer tick received (time to update data)
		case <-ticker.C: // ticker.C is a channel that receives time values
			// Fetch new system data and update the display
			updateDisplay(cpuGauge, memoryGauge, diskGauge, infoList)
		}
	}
}

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
	cpuGauge.SetRect(0, 0, width/3, height/2) // Left third: x from 0 to width/3
	cpuGauge.BarColor = ui.ColorYellow        // Yellow bar (warning color)
	cpuGauge.BorderStyle.Fg = ui.ColorWhite   // White border
	cpuGauge.TitleStyle.Fg = ui.ColorCyan     // Cyan title

	// Memory Gauge - Middle third of screen, top half
	memoryGauge.Title = "Memory Usage"
	memoryGauge.SetRect(width/3, 0, 2*width/3, height/2) // Middle third
	memoryGauge.BarColor = ui.ColorGreen                 // Green bar (safe color)
	memoryGauge.BorderStyle.Fg = ui.ColorWhite
	memoryGauge.TitleStyle.Fg = ui.ColorCyan

	// Disk Gauge - Right third of screen, top half
	diskGauge.Title = "Disk Usage"
	diskGauge.SetRect(2*width/3, 0, width, height/2) // Right third
	diskGauge.BarColor = ui.ColorRed                 // Red bar (danger color)
	diskGauge.BorderStyle.Fg = ui.ColorWhite
	diskGauge.TitleStyle.Fg = ui.ColorCyan

	// Info List - Full width, bottom half
	infoList.Title = "System Information"
	infoList.SetRect(0, height/2, width, height) // Full width, bottom half
	infoList.TextStyle = ui.NewStyle(ui.ColorWhite)
	infoList.WrapText = false // Don't wrap long lines
	infoList.BorderStyle.Fg = ui.ColorWhite
	infoList.TitleStyle.Fg = ui.ColorCyan
}

// updateDisplay fetches current system stats and updates all UI components.
// It uses concurrent data fetching for optimal performance and responsiveness.
func updateDisplay(cpuGauge, memoryGauge, diskGauge *widgets.Gauge, infoList *widgets.List) {
	// CONCURRENT DATA FETCHING - Don't block the UI!
	// Create a channel to receive the complete system stats
	statsCh := make(chan SystemStats, 1) // Buffered channel (can hold 1 value)
	// Start a goroutine to fetch all data concurrently
	go fetchSystemStats(statsCh) // This runs in the background

	// BLOCKING RECEIVE - Wait for the goroutine to send us data
	stats := <-statsCh // This blocks until data arrives

	// UPDATE GAUGES - Convert our data to visual elements
	// Gauges expect integer percentages (0-100)
	cpuGauge.Percent = int(stats.CPUUsage)                 // Convert float to int
	cpuGauge.Label = fmt.Sprintf("%.1f%%", stats.CPUUsage) // Format with 1 decimal

	memoryGauge.Percent = int(stats.MemoryUsage)
	memoryGauge.Label = fmt.Sprintf("%.1f%%", stats.MemoryUsage)

	diskGauge.Percent = int(stats.DiskUsage)
	diskGauge.Label = fmt.Sprintf("%.1f%%", stats.DiskUsage)

	// UPDATE INFO LIST - Create detailed text information
	// infoList.Rows is a slice of strings (like an array but dynamic)
	infoList.Rows = []string{
		fmt.Sprintf("Time: %s", time.Now().Format(config.TimeFormat)),
		"", // Empty line for spacing
		fmt.Sprintf("CPU: %.1f%%", stats.CPUUsage),
		"",
		fmt.Sprintf("Memory: %.1f%% (%.1f GB / %.1f GB)",
			stats.MemoryUsage, stats.MemoryUsed, stats.MemoryTotal),
		"",
		fmt.Sprintf("Disk (%s): %.1f%% (%.1f GB / %.1f GB)",
			config.DiskDrive, stats.DiskUsage, stats.DiskUsed, stats.DiskTotal),
		"",
		"Press 'q' or Ctrl+C to quit", // User instruction
	}

	// RENDER - Actually draw everything to the screen
	// This is when the user sees the updated information
	ui.Render(cpuGauge, memoryGauge, diskGauge, infoList)
}

// fetchSystemStats gathers all system statistics using concurrent goroutines.
// It coordinates multiple system calls and returns consolidated data via channel.
func fetchSystemStats(statsCh chan SystemStats) {
	// Create empty stats struct to fill with data
	var stats SystemStats

	// CREATE CHANNELS for each type of data we need to fetch
	// These act like "mailboxes" for goroutines to send results
	cpuCh := make(chan float64, 1)                // For CPU percentage
	memCh := make(chan *mem.VirtualMemoryStat, 1) // For memory info (pointer to struct)
	diskCh := make(chan *disk.UsageStat, 1)       // For disk info (pointer to struct)

	// START ALL GOROUTINES AT ONCE - They run simultaneously!
	// This is much faster than fetching one after another
	go fetchCPUUsage(cpuCh)    // Goroutine 1: Get CPU data
	go fetchMemoryUsage(memCh) // Goroutine 2: Get memory data
	go fetchDiskUsage(diskCh)  // Goroutine 3: Get disk data

	// COLLECT RESULTS - Wait for each goroutine to send data
	// The order doesn't matter - we process them as they arrive

	// Collect CPU result
	if cpuUsage := <-cpuCh; cpuUsage >= 0 { // Negative values indicate error
		stats.CPUUsage = cpuUsage
	}

	// Collect Memory result
	if vmStat := <-memCh; vmStat != nil { // nil indicates error
		stats.MemoryUsage = vmStat.UsedPercent
		// Convert bytes to gigabytes: divide by 1024³
		stats.MemoryUsed = float64(vmStat.Used) / 1024 / 1024 / 1024
		stats.MemoryTotal = float64(vmStat.Total) / 1024 / 1024 / 1024
	}

	// Collect Disk result
	if diskStat := <-diskCh; diskStat != nil { // nil indicates error
		stats.DiskUsage = diskStat.UsedPercent
		// Convert bytes to gigabytes: divide by 1024³
		stats.DiskUsed = float64(diskStat.Used) / 1024 / 1024 / 1024
		stats.DiskTotal = float64(diskStat.Total) / 1024 / 1024 / 1024
	}

	// SEND COMPLETE STATS - Send our filled struct to the waiting function
	statsCh <- stats
}

// fetchCPUUsage retrieves current CPU utilization percentage.
// It measures usage over config.CPUSampleDuration and sends result to cpuCh.
// Sends -1 on error as an error indicator.
func fetchCPUUsage(cpuCh chan float64) {
	// cpu.Percent() measures CPU usage over a time period
	// config.CPUSampleDuration is how long to measure (100ms for responsiveness)
	// false means "don't get per-CPU stats, just overall average"
	if percentages, err := cpu.Percent(config.CPUSampleDuration, false); err == nil && len(percentages) > 0 {
		// Success! Send the first (and only) percentage value
		cpuCh <- percentages[0]
	} else {
		// Error occurred - send -1 as error indicator
		cpuCh <- -1
	}
}

// fetchMemoryUsage retrieves current memory usage statistics.
// It queries virtual memory info and sends a pointer to the result via memCh.
// Sends nil on error as an error indicator.
func fetchMemoryUsage(memCh chan *mem.VirtualMemoryStat) {
	// mem.VirtualMemory() returns a pointer to a struct with memory info
	if vmStat, err := mem.VirtualMemory(); err == nil {
		// Success! Send the pointer to the struct
		memCh <- vmStat
	} else {
		// Error occurred - send nil pointer as error indicator
		memCh <- nil
	}
}

// fetchDiskUsage retrieves disk usage statistics for the configured drive.
// It queries the drive specified in config.DiskDrive and sends result via diskCh.
// Sends nil on error as an error indicator.
func fetchDiskUsage(diskCh chan *disk.UsageStat) {
	// disk.Usage() gets information about the specified drive
	// config.DiskDrive is set in our configuration (usually "C:" on Windows)
	if diskStat, err := disk.Usage(config.DiskDrive); err == nil {
		// Success! Send the pointer to the disk stats struct
		diskCh <- diskStat
	} else {
		// Error occurred - send nil pointer as error indicator
		diskCh <- nil
	}
}
