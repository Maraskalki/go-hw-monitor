package main

import (
	"fmt"
	"log"
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

// displaySystemStats shows CPU, Memory, and Disk usage
func displaySystemStats() {
	// Clear screen (Windows compatible)
	clearScreen()

	fmt.Println(config.Title)
	fmt.Println(config.Separator)
	fmt.Printf("Time: %s\n\n", time.Now().Format(config.TimeFormat))

	// Get CPU usage
	getCPUUsage()

	// Get Memory usage
	getMemoryUsage()

	// Get Disk usage
	getDiskUsage()
}

// getCPUUsage displays CPU usage percentage
func getCPUUsage() {
	// Get CPU usage percentage using configured sample duration
	percentages, err := cpu.Percent(config.CPUSampleDuration, false)
	if err != nil {
		log.Printf("Error getting CPU usage: %v", err)
		return
	}

	if len(percentages) > 0 {
		fmt.Printf("CPU Usage: %.*f%%\n", config.DecimalPlaces, percentages[0])
	}
}

// getMemoryUsage displays memory usage
func getMemoryUsage() {
	// Get virtual memory statistics
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error getting memory usage: %v", err)
		return
	}

	fmt.Printf("Memory Usage: %.*f%% (Used: %.*f GB / Total: %.*f GB)\n",
		config.DecimalPlaces, vmStat.UsedPercent,
		config.DecimalPlaces, float64(vmStat.Used)/1024/1024/1024,
		config.DecimalPlaces, float64(vmStat.Total)/1024/1024/1024)
}

// getDiskUsage displays disk usage for configured drive
func getDiskUsage() {
	// Get disk usage for configured drive
	diskStat, err := disk.Usage(config.DiskDrive)
	if err != nil {
		log.Printf("Error getting disk usage: %v", err)
		return
	}

	fmt.Printf("Disk Usage (%s): %.*f%% (Used: %.*f GB / Total: %.*f GB)\n",
		config.DiskDrive,
		config.DecimalPlaces, diskStat.UsedPercent,
		config.DecimalPlaces, float64(diskStat.Used)/1024/1024/1024,
		config.DecimalPlaces, float64(diskStat.Total)/1024/1024/1024)
}
