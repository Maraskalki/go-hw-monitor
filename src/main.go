// Package main provides a hardware monitoring application with terminal UI.
// It demonstrates Go concurrency, system programming, and real-time data visualization.
package main

import (
	"fmt"
	"log"
	"sync" // For WaitGroup concurrency coordination
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

// MetricResult represents the result of a single metric collection operation.
// It provides proper error handling instead of using sentinel values.
type MetricResult struct {
	Type  string      // Metric type: "cpu", "memory", or "disk"
	Value interface{} // The actual metric data
	Error error       // Any error that occurred during collection
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
func updateDisplay(cpuGauge, memoryGauge, diskGauge *widgets.Gauge, infoList *widgets.List) {
	// CONCURRENT DATA FETCHING - Don't block the UI!
	// Create a channel to receive the complete system stats
	statsCh := make(chan SystemStats, config.ChannelBuffer) // Buffered channel
	// Start a goroutine to fetch all data concurrently
	go fetchSystemStats(statsCh) // This runs in the background

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

// fetchSystemStats gathers all system statistics using WaitGroup coordination.
// It demonstrates proper Go concurrency patterns with error handling.
func fetchSystemStats(statsCh chan SystemStats) {
	// Create empty stats struct to fill with data
	var stats SystemStats

	// WAITGROUP COORDINATION - Better than manual channel management
	var wg sync.WaitGroup
	results := make(chan MetricResult, config.ResultsBuffer) // Buffered channel for all results

	// START ALL GOROUTINES WITH WAITGROUP COORDINATION
	// Each goroutine will signal completion via wg.Done()
	wg.Add(config.MetricCount) // We're starting configured number of goroutines

	go fetchCPUMetric(&wg, results)    // Goroutine 1: Get CPU data
	go fetchMemoryMetric(&wg, results) // Goroutine 2: Get memory data
	go fetchDiskMetric(&wg, results)   // Goroutine 3: Get disk data

	// WAIT FOR ALL GOROUTINES TO COMPLETE
	// This is safer than waiting for channels individually
	go func() {
		wg.Wait()      // Block until all goroutines call Done()
		close(results) // Signal that no more data will be sent
	}()

	// COLLECT AND PROCESS ALL RESULTS
	// Range over channel until it's closed
	for result := range results {
		if result.Error != nil {
			// Log error but continue with other metrics
			log.Printf("Error fetching %s metric: %v", result.Type, result.Error)
			continue
		}

		// Process successful results based on type
		switch result.Type {
		case "cpu":
			if cpuUsage, ok := result.Value.(float64); ok {
				stats.CPUUsage = cpuUsage
			}
		case "memory":
			if vmStat, ok := result.Value.(*mem.VirtualMemoryStat); ok {
				stats.MemoryUsage = vmStat.UsedPercent
				// Convert bytes to gigabytes using config constant
				stats.MemoryUsed = float64(vmStat.Used) / float64(config.BytesToGB)
				stats.MemoryTotal = float64(vmStat.Total) / float64(config.BytesToGB)
			}
		case "disk":
			if diskStat, ok := result.Value.(*disk.UsageStat); ok {
				stats.DiskUsage = diskStat.UsedPercent
				// Convert bytes to gigabytes using config constant
				stats.DiskUsed = float64(diskStat.Used) / float64(config.BytesToGB)
				stats.DiskTotal = float64(diskStat.Total) / float64(config.BytesToGB)
			}
		}
	}

	// SEND COMPLETE STATS - Send our filled struct to the waiting function
	statsCh <- stats
}

// fetchCPUMetric retrieves CPU usage with proper error handling and WaitGroup coordination.
// It demonstrates how to integrate WaitGroup with error handling.
func fetchCPUMetric(wg *sync.WaitGroup, results chan<- MetricResult) {
	// ALWAYS call Done() when function exits - use defer for safety
	defer wg.Done()

	// cpu.Percent() measures CPU usage over a time period
	// config.CPUSampleDuration is how long to measure (100ms for responsiveness)
	// false means "don't get per-CPU stats, just overall average"
	percentages, err := cpu.Percent(config.CPUSampleDuration, false)
	if err != nil {
		// Send proper error instead of sentinel value
		results <- MetricResult{Type: "cpu", Value: nil, Error: fmt.Errorf("failed to get CPU usage: %w", err)}
		return
	}

	if len(percentages) == 0 {
		// Handle edge case where no data is returned
		results <- MetricResult{Type: "cpu", Value: nil, Error: fmt.Errorf("no CPU usage data returned")}
		return
	}

	// Success! Send the actual value with no error
	results <- MetricResult{Type: "cpu", Value: percentages[0], Error: nil}
}

// fetchMemoryMetric retrieves memory usage with proper error handling and WaitGroup coordination.
func fetchMemoryMetric(wg *sync.WaitGroup, results chan<- MetricResult) {
	// ALWAYS call Done() when function exits - use defer for safety
	defer wg.Done()

	// mem.VirtualMemory() returns a pointer to a struct with memory info
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		// Send proper error instead of nil pointer
		results <- MetricResult{Type: "memory", Value: nil, Error: fmt.Errorf("failed to get memory usage: %w", err)}
		return
	}

	// Success! Send the actual data with no error
	results <- MetricResult{Type: "memory", Value: vmStat, Error: nil}
}

// fetchDiskMetric retrieves disk usage with proper error handling and WaitGroup coordination.
func fetchDiskMetric(wg *sync.WaitGroup, results chan<- MetricResult) {
	// ALWAYS call Done() when function exits - use defer for safety
	defer wg.Done()

	// disk.Usage() gets information about the specified drive
	// config.DiskDrive is set in our configuration (usually "C:" on Windows)
	diskStat, err := disk.Usage(config.DiskDrive)
	if err != nil {
		// Send proper error instead of nil pointer
		results <- MetricResult{Type: "disk", Value: nil, Error: fmt.Errorf("failed to get disk usage for %s: %w", config.DiskDrive, err)}
		return
	}

	// Success! Send the actual data with no error
	results <- MetricResult{Type: "disk", Value: diskStat, Error: nil}
}
