package main

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

func main() {
	for {
		fmt.Println("--- System Monitor ---")

		// 1. CPU Usage
		percent, _ := cpu.Percent(time.Second, false)
		fmt.Printf("CPU Usage:    %.2f%%\n", percent[0])

		// 2. Memory (RAM) Usage
		v, _ := mem.VirtualMemory()
		fmt.Printf("RAM Usage:    %.2f%% (Used: %vMB / Total: %vMB)\n",
			v.UsedPercent, v.Used/1024/1024, v.Total/1024/1024)

		// 3. Disk Usage (Lấy ổ đĩa gốc /)
		d, _ := disk.Usage("/")
		fmt.Printf("Disk Usage:   %.2f%% (Free: %vGB)\n",
			d.UsedPercent, d.Free/1024/1024/1024)

		// Lưu ý về GPU: gopsutil không hỗ trợ trực tiếp GPU.
		// Bạn cần dùng các lệnh hệ thống hoặc thư viện riêng (như NVML cho NVIDIA).
		checkGPU()

		fmt.Println("----------------------")
		time.Sleep(2 * time.Second)
	}
}

func checkGPU() {
	// Đây là placeholder vì GPU phụ thuộc vào phần cứng (NVIDIA/AMD)
	fmt.Println("GPU Usage:    N/A (Requires NVML/CGO for NVIDIA)")
}
