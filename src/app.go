// Package main provides the application structure and lifecycle management for the hardware monitor.
// This file contains the App struct and all its methods for clean separation of concerns.
package main

import (
	"fmt"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// App encapsulates the application state and provides a clean interface for the monitor.
// This struct groups related components and makes the code more organized and testable.
type App struct {
	cpuGauge    *widgets.Gauge
	memoryGauge *widgets.Gauge
	diskGauge   *widgets.Gauge
	infoList    *widgets.List
	ticker      *time.Ticker
	uiEvents    <-chan ui.Event
	monitor     SystemMonitor // App manages its own monitor instance
}

// newApp creates a new App instance with all components initialized and configured.
// It now handles its own UI initialization and creates its own monitor for complete encapsulation.
func newApp() (*App, error) {
	// Initialize the terminal UI system first
	if err := ui.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize termui: %w", err)
	}

	// Create the monitor instance - App handles its own dependencies
	monitor := NewGopsutilMonitor()

	// Create UI components using the factory function from ui.go
	cpuGauge, memoryGauge, diskGauge, infoList := createWidgets()

	// Setup UI layout - position and style all widgets
	setupUI(cpuGauge, memoryGauge, diskGauge, infoList)

	// Create ticker for periodic updates
	ticker := time.NewTicker(config.RefreshInterval)

	// Get UI event channel
	uiEvents := ui.PollEvents()

	return &App{
		cpuGauge:    cpuGauge,
		memoryGauge: memoryGauge,
		diskGauge:   diskGauge,
		infoList:    infoList,
		ticker:      ticker,
		uiEvents:    uiEvents,
		monitor:     monitor, // App owns its monitor
	}, nil
}

// cleanup properly releases resources when the application exits.
// Now handles both ticker and UI cleanup for complete resource management.
func (app *App) cleanup() {
	if app.ticker != nil {
		app.ticker.Stop()
	}
	// Close the UI system
	ui.Close()
}

// run executes the main application loop with event handling.
func (app *App) run() {
	// Do initial update to show data immediately
	app.updateDisplay()

	// Main event loop - clean and focused
	for {
		select {
		case e := <-app.uiEvents:
			if app.handleUIEvent(e) {
				return // Exit requested
			}
		case <-app.ticker.C:
			app.updateDisplay()
		}
	}
}

// handleUIEvent processes user input events and returns true if the app should exit.
func (app *App) handleUIEvent(e ui.Event) bool {
	switch e.ID {
	case "q", "<C-c>":
		return true // Signal to exit
	case "<Resize>":
		app.handleResize(e)
	}
	return false // Continue running
}

// handleResize recalculates layout when the terminal window is resized.
func (app *App) handleResize(e ui.Event) {
	payload := e.Payload.(ui.Resize)
	setupUIWithSize(app.cpuGauge, app.memoryGauge, app.diskGauge, app.infoList, payload.Width, payload.Height)
	ui.Render(app.cpuGauge, app.memoryGauge, app.diskGauge, app.infoList)
}

// updateDisplay refreshes the UI with current system data.
func (app *App) updateDisplay() {
	updateDisplay(app.cpuGauge, app.memoryGauge, app.diskGauge, app.infoList, app.monitor)
}
