// Package main provides data fetching and coordination logic for the hardware monitor.
// This file contains all system statistics gathering, concurrent processing, and result coordination.
package main

import (
	"log"
	"sync" // For WaitGroup concurrency coordination
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
		// Now we work with our clean interface types!
		switch result.Type {
		case "cpu":
			if cpuUsage, ok := result.Value.(float64); ok {
				stats.CPUUsage = cpuUsage
			}
		case "memory":
			// Now we get clean MemoryInfo instead of gopsutil's VirtualMemoryStat
			if memInfo, ok := result.Value.(*MemoryInfo); ok {
				stats.MemoryUsage = memInfo.UsedPercent
				// Convert bytes to gigabytes using config constant
				stats.MemoryUsed = float64(memInfo.Used) / float64(config.BytesToGB)
				stats.MemoryTotal = float64(memInfo.Total) / float64(config.BytesToGB)
			}
		case "disk":
			// Now we get clean DiskInfo instead of gopsutil's UsageStat
			if diskInfo, ok := result.Value.(*DiskInfo); ok {
				stats.DiskUsage = diskInfo.UsedPercent
				// Convert bytes to gigabytes using config constant
				stats.DiskUsed = float64(diskInfo.Used) / float64(config.BytesToGB)
				stats.DiskTotal = float64(diskInfo.Total) / float64(config.BytesToGB)
			}
		}
	}

	// SEND COMPLETE STATS - Send our filled struct to the waiting function
	statsCh <- stats
}

// fetchCPUMetric retrieves CPU usage using our SystemMonitor interface.
// This demonstrates interface usage - we don't know or care what implementation is used!
func fetchCPUMetric(wg *sync.WaitGroup, results chan<- MetricResult) {
	// ALWAYS call Done() when function exits - use defer for safety
	defer wg.Done()

	// USE THE INTERFACE! This is the key change.
	// We call monitor.GetCPUUsage instead of cpu.Percent directly
	// The function doesn't know if it's talking to GopsutilMonitor, MockMonitor, etc.
	cpuUsage, err := monitor.GetCPUUsage(config.CPUSampleDuration)
	if err != nil {
		// Interface already wrapped the error nicely
		results <- MetricResult{Type: "cpu", Value: nil, Error: err}
		return
	}

	// Success! Send the clean result
	results <- MetricResult{Type: "cpu", Value: cpuUsage, Error: nil}
}

// fetchMemoryMetric retrieves memory usage using our SystemMonitor interface.
// Clean and simple - just like the CPU version!
func fetchMemoryMetric(wg *sync.WaitGroup, results chan<- MetricResult) {
	// ALWAYS call Done() when function exits - use defer for safety
	defer wg.Done()

	// USE THE INTERFACE! Much simpler than the old version
	memoryInfo, err := monitor.GetMemoryUsage()
	if err != nil {
		// Interface already wrapped the error nicely
		results <- MetricResult{Type: "memory", Value: nil, Error: err}
		return
	}

	// Success! Send the clean result
	results <- MetricResult{Type: "memory", Value: memoryInfo, Error: nil}
}

// fetchDiskMetric retrieves disk usage using our SystemMonitor interface.
// Clean and consistent with other interface-based functions!
func fetchDiskMetric(wg *sync.WaitGroup, results chan<- MetricResult) {
	// ALWAYS call Done() when function exits - use defer for safety
	defer wg.Done()

	// USE THE INTERFACE! Consistent pattern across all metrics
	diskInfo, err := monitor.GetDiskUsage(config.DiskDrive)
	if err != nil {
		// Interface already wrapped the error nicely
		results <- MetricResult{Type: "disk", Value: nil, Error: err}
		return
	}

	// Success! Send the clean result
	results <- MetricResult{Type: "disk", Value: diskInfo, Error: nil}
}
