package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

// Global configuration
var config = Config

func main() {
	// Run monitoring loop using configured refresh interval
	for {
		displaySystemStats()
		time.Sleep(config.RefreshInterval)
	}
}

// clearScreen clears the console screen in a cross-platform way
func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// fetchCPUUsage fetches CPU usage as a goroutine
func fetchCPUUsage(cpuCh chan float64) {
	if percentages, err := cpu.Percent(config.CPUSampleDuration, false); err == nil && len(percentages) > 0 {
		cpuCh <- percentages[0]
	} else {
		cpuCh <- -1 // Error indicator
	}
}

// fetchMemoryUsage fetches memory usage as a goroutine
func fetchMemoryUsage(memCh chan *mem.VirtualMemoryStat) {
	if vmStat, err := mem.VirtualMemory(); err == nil {
		memCh <- vmStat
	} else {
		memCh <- nil
	}
}

// fetchDiskUsage fetches disk usage as a goroutine
func fetchDiskUsage(diskCh chan *disk.UsageStat) {
	if diskStat, err := disk.Usage(config.DiskDrive); err == nil {
		diskCh <- diskStat
	} else {
		diskCh <- nil
	}
}

// displaySystemStats shows CPU, Memory, and Disk usage
func displaySystemStats() {
	// Clear screen (Windows compatible)
	clearScreen()

	fmt.Println(config.Title)
	fmt.Println(config.Separator)
	fmt.Printf("Time: %s\n\n", time.Now().Format(config.TimeFormat))

	// Create channels for concurrent data fetching
	cpuCh := make(chan float64, 1)
	memCh := make(chan *mem.VirtualMemoryStat, 1)
	diskCh := make(chan *disk.UsageStat, 1)

	// Start all data fetching goroutines
	go fetchCPUUsage(cpuCh)
	go fetchMemoryUsage(memCh)
	go fetchDiskUsage(diskCh)

	// Display results as they come in
	displayResults(cpuCh, memCh, diskCh)
}

// displayResults displays stats from concurrent fetches
func displayResults(cpuCh chan float64, memCh chan *mem.VirtualMemoryStat, diskCh chan *disk.UsageStat) {
	// Get CPU usage
	if cpuUsage := <-cpuCh; cpuUsage >= 0 {
		fmt.Printf("CPU Usage: %.*f%%\n", config.DecimalPlaces, cpuUsage)
	} else {
		fmt.Println("CPU Usage: Error")
	}

	// Get Memory usage
	if vmStat := <-memCh; vmStat != nil {
		fmt.Printf("Memory Usage: %.*f%% (Used: %.*f GB / Total: %.*f GB)\n",
			config.DecimalPlaces, vmStat.UsedPercent,
			config.DecimalPlaces, float64(vmStat.Used)/1024/1024/1024,
			config.DecimalPlaces, float64(vmStat.Total)/1024/1024/1024)
	} else {
		fmt.Println("Memory Usage: Error")
	}

	// Get Disk usage
	if diskStat := <-diskCh; diskStat != nil {
		fmt.Printf("Disk Usage (%s): %.*f%% (Used: %.*f GB / Total: %.*f GB)\n",
			config.DiskDrive,
			config.DecimalPlaces, diskStat.UsedPercent,
			config.DecimalPlaces, float64(diskStat.Used)/1024/1024/1024,
			config.DecimalPlaces, float64(diskStat.Total)/1024/1024/1024)
	} else {
		fmt.Printf("Disk Usage (%s): Error\n", config.DiskDrive)
	}
}
