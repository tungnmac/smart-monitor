package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

func main() {
	// Thresholds for alerts
	cpuThreshold := 80.0
	ramThreshold := 80.0
	processCount := 5 // Top N processes to show

	for {
		fmt.Println("\n╔════════════════════════════════════════════════════════════════╗")
		fmt.Println("║              System Monitor with Process Alerts                ║")
		fmt.Println("╚════════════════════════════════════════════════════════════════╝")

		// 1. CPU Usage
		percent, _ := cpu.Percent(time.Second, false)
		cpuUsage := percent[0]
		cpuStatus := getStatus(cpuUsage, cpuThreshold)
		fmt.Printf("\n📊 CPU Usage:    %.2f%% %s\n", cpuUsage, cpuStatus)

		// 2. Memory (RAM) Usage
		v, _ := mem.VirtualMemory()
		ramStatus := getStatus(v.UsedPercent, ramThreshold)
		fmt.Printf("💾 RAM Usage:    %.2f%% (Used: %vMB / Total: %vMB) %s\n",
			v.UsedPercent, v.Used/1024/1024, v.Total/1024/1024, ramStatus)

		// 3. Disk Usage (Lấy ổ đĩa gốc /)
		d, _ := disk.Usage("/")
		diskStatus := getStatus(d.UsedPercent, 90.0)
		fmt.Printf("💿 Disk Usage:   %.2f%% (Free: %vGB) %s\n",
			d.UsedPercent, d.Free/1024/1024/1024, diskStatus)

		// 4. Process Monitoring - Top CPU consumers
		if cpuUsage > cpuThreshold {
			fmt.Printf("\n⚠️  WARNING: High CPU usage detected (%.2f%%)!\n", cpuUsage)
			fmt.Println("🔥 Top processes by CPU:")
			showTopProcessesByCPU(processCount)
		}

		// 5. Process Monitoring - Top RAM consumers
		if v.UsedPercent > ramThreshold {
			fmt.Printf("\n⚠️  WARNING: High RAM usage detected (%.2f%%)!\n", v.UsedPercent)
			fmt.Println("🔥 Top processes by RAM:")
			showTopProcessesByMemory(processCount)
		}

		// 6. Always show top consumers (for monitoring)
		fmt.Println("\n📈 Top Resource Consumers:")
		fmt.Println("\n  CPU Top 3:")
		showTopProcessesByCPU(3)
		fmt.Println("\n  RAM Top 3:")
		showTopProcessesByMemory(3)

		// GPU check
		checkGPU()

		fmt.Println("\n" + strings.Repeat("─", 64))
		fmt.Printf("Next update in 10 seconds... (Ctrl+C to stop)\n")
		time.Sleep(10 * time.Second)
	}
}

func checkGPU() {
	// Đây là placeholder vì GPU phụ thuộc vào phần cứng (NVIDIA/AMD)
	fmt.Println("\n🎮 GPU Usage:    N/A (Requires NVML/CGO for NVIDIA)")
}

// getStatus returns a status indicator based on threshold
func getStatus(usage, threshold float64) string {
	if usage >= threshold {
		return "🔴 CRITICAL"
	} else if usage >= threshold*0.75 {
		return "🟡 WARNING"
	}
	return "🟢 OK"
}

// ProcessInfo holds process information
type ProcessInfo struct {
	PID        int32
	Name       string
	CPUPercent float64
	MemoryMB   float64
	MemPercent float32
	Status     string
	Username   string
}

// showTopProcessesByCPU displays top processes by CPU usage
func showTopProcessesByCPU(count int) {
	procs, err := process.Processes()
	if err != nil {
		fmt.Printf("   Error getting processes: %v\n", err)
		return
	}

	var procInfos []ProcessInfo
	for _, p := range procs {
		cpuPercent, err := p.CPUPercent()
		if err != nil {
			continue
		}

		// Only include processes with measurable CPU usage
		if cpuPercent < 0.1 {
			continue
		}

		name, _ := p.Name()
		memInfo, _ := p.MemoryInfo()
		memPercent, _ := p.MemoryPercent()
		statusArr, _ := p.Status()
		username, _ := p.Username()

		statusStr := ""
		if len(statusArr) > 0 {
			statusStr = statusArr[0]
		}

		memMB := float64(0)
		if memInfo != nil {
			memMB = float64(memInfo.RSS) / 1024 / 1024
		}

		procInfos = append(procInfos, ProcessInfo{
			PID:        p.Pid,
			Name:       name,
			CPUPercent: cpuPercent,
			MemoryMB:   memMB,
			MemPercent: memPercent,
			Status:     statusStr,
			Username:   username,
		})
	}

	// Sort by CPU usage
	sort.Slice(procInfos, func(i, j int) bool {
		return procInfos[i].CPUPercent > procInfos[j].CPUPercent
	})

	// Display top N
	displayCount := count
	if len(procInfos) < count {
		displayCount = len(procInfos)
	}

	fmt.Printf("   %-8s %-25s %-10s %-12s %s\n", "PID", "NAME", "CPU%", "RAM(MB)", "USER")
	fmt.Println("   " + strings.Repeat("─", 70))
	for i := 0; i < displayCount; i++ {
		p := procInfos[i]
		userName := p.Username
		if len(userName) > 15 {
			userName = userName[:12] + "..."
		}
		fmt.Printf("   %-8d %-25s %-10.2f %-12.1f %s\n",
			p.PID, truncateString(p.Name, 25), p.CPUPercent, p.MemoryMB, userName)
	}
}

// showTopProcessesByMemory displays top processes by memory usage
func showTopProcessesByMemory(count int) {
	procs, err := process.Processes()
	if err != nil {
		fmt.Printf("   Error getting processes: %v\n", err)
		return
	}

	var procInfos []ProcessInfo
	for _, p := range procs {
		memInfo, err := p.MemoryInfo()
		if err != nil || memInfo == nil {
			continue
		}

		memMB := float64(memInfo.RSS) / 1024 / 1024

		// Only include processes using significant memory (> 10MB)
		if memMB < 10 {
			continue
		}

		name, _ := p.Name()
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()
		statusArr, _ := p.Status()
		username, _ := p.Username()

		statusStr := ""
		if len(statusArr) > 0 {
			statusStr = statusArr[0]
		}

		procInfos = append(procInfos, ProcessInfo{
			PID:        p.Pid,
			Name:       name,
			CPUPercent: cpuPercent,
			MemoryMB:   memMB,
			MemPercent: memPercent,
			Status:     statusStr,
			Username:   username,
		})
	}

	// Sort by memory usage
	sort.Slice(procInfos, func(i, j int) bool {
		return procInfos[i].MemoryMB > procInfos[j].MemoryMB
	})

	// Display top N
	displayCount := count
	if len(procInfos) < count {
		displayCount = len(procInfos)
	}

	fmt.Printf("   %-8s %-25s %-12s %-10s %s\n", "PID", "NAME", "RAM(MB)", "RAM%", "USER")
	fmt.Println("   " + strings.Repeat("─", 70))
	for i := 0; i < displayCount; i++ {
		p := procInfos[i]
		userName := p.Username
		if len(userName) > 15 {
			userName = userName[:12] + "..."
		}
		fmt.Printf("   %-8d %-25s %-12.1f %-10.2f %s\n",
			p.PID, truncateString(p.Name, 25), p.MemoryMB, p.MemPercent, userName)
	}
}

// truncateString truncates string to specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
