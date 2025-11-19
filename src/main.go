package main

import (
	"fmt"
	"log"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

func main() {
	fmt.Println("Hardware Monitor - Press Ctrl+C to stop")
	fmt.Println("=========================================")

	// Run monitoring loop every second
	for {
		displaySystemStats()
		time.Sleep(1 * time.Second)
	}
}

// displaySystemStats shows CPU, Memory, and Disk usage
func displaySystemStats() {
	// Clear screen (Windows command)
	fmt.Print("\033[2J\033[H")

	fmt.Println("Hardware Monitor - Press Ctrl+C to stop")
	fmt.Println("=========================================")
	fmt.Printf("Time: %s\n\n", time.Now().Format("15:04:05"))

	// Get CPU usage
	getCPUUsage()

	// Get Memory usage
	getMemoryUsage()

	// Get Disk usage
	getDiskUsage()
}

// getCPUUsage displays CPU usage percentage
func getCPUUsage() {
	// Get CPU usage percentage (averaged over 1 second)
	percentages, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		log.Printf("Error getting CPU usage: %v", err)
		return
	}

	if len(percentages) > 0 {
		fmt.Printf("🖥️  CPU Usage: %.1f%%\n", percentages[0])
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

	fmt.Printf("🧠 Memory Usage: %.1f%% (Used: %.1f GB / Total: %.1f GB)\n",
		vmStat.UsedPercent,
		float64(vmStat.Used)/1024/1024/1024,
		float64(vmStat.Total)/1024/1024/1024)
}

// getDiskUsage displays disk usage for C: drive
func getDiskUsage() {
	// Get disk usage for C: drive (main Windows drive)
	diskStat, err := disk.Usage("C:")
	if err != nil {
		log.Printf("Error getting disk usage: %v", err)
		return
	}

	fmt.Printf("💾 Disk Usage (C:): %.1f%% (Used: %.1f GB / Total: %.1f GB)\n",
		diskStat.UsedPercent,
		float64(diskStat.Used)/1024/1024/1024,
		float64(diskStat.Total)/1024/1024/1024)
}
