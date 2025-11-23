// Package main provides a hardware monitoring application with terminal UI.
// It demonstrates Go concurrency, system programming, and real-time data visualization.
package main

import (
	"log"
	"time"

	// Import with alias - 'ui' is shorter than 'termui'
	ui "github.com/gizak/termui/v3"
)

// Global configuration - accessible from anywhere in this package
var config = Config

// Global monitor instance - this is our interface in action!
// We use the constructor function to create a clean instance
var monitor SystemMonitor = NewGopsutilMonitor()

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

	// Create UI components using the factory function from ui.go
	cpuGauge, memoryGauge, diskGauge, infoList := createWidgets()

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
