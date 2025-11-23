// Package main provides a hardware monitoring application with terminal UI.
// This file contains only the entry point and global configuration.
package main

import (
	"log"
)

// Global configuration - accessible from anywhere in this package
var config = Config

// main - Entry point of our program, now completely focused on coordination
func main() {
	// Create and setup the application (handles its own UI initialization)
	app, err := newApp()
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}
	defer app.cleanup()

	// Run the main application loop
	app.run()
}
