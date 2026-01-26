// Package collector handles system metrics collection
package collector

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// Metrics holds collected system metrics
type Metrics struct {
	// Basic metrics
	CPUPercent  float64
	RAMPercent  float64
	DiskPercent float64

	// Extended metrics
	CPUCores     int
	RAMTotal     uint64
	RAMUsed      uint64
	DiskTotal    uint64
	DiskUsed     uint64
	LoadAverage  []float64
	Uptime       uint64
	ProcessCount uint64

	// Network metrics
	NetworkSent uint64
	NetworkRecv uint64

	Timestamp time.Time
}

// Collector collects system metrics
type Collector struct {
	// Configuration
	diskPath string
}

// NewCollector creates a new metrics collector
func NewCollector(diskPath string) *Collector {
	if diskPath == "" {
		diskPath = "/"
	}
	return &Collector{
		diskPath: diskPath,
	}
}

// Collect gathers current system metrics
func (c *Collector) Collect() (*Metrics, error) {
	metrics := &Metrics{
		Timestamp: time.Now(),
	}

	// CPU metrics
	if err := c.collectCPU(metrics); err != nil {
		return nil, fmt.Errorf("failed to collect CPU metrics: %w", err)
	}

	// Memory metrics
	if err := c.collectMemory(metrics); err != nil {
		return nil, fmt.Errorf("failed to collect memory metrics: %w", err)
	}

	// Disk metrics
	if err := c.collectDisk(metrics); err != nil {
		return nil, fmt.Errorf("failed to collect disk metrics: %w", err)
	}

	// Load average
	if err := c.collectLoadAverage(metrics); err != nil {
		// Non-critical, log but continue
		metrics.LoadAverage = []float64{0, 0, 0}
	}

	// Host info
	if err := c.collectHostInfo(metrics); err != nil {
		// Non-critical
	}

	// Network metrics
	if err := c.collectNetwork(metrics); err != nil {
		// Non-critical
	}

	return metrics, nil
}

// collectCPU collects CPU metrics
func (c *Collector) collectCPU(metrics *Metrics) error {
	// CPU usage percentage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return err
	}
	if len(cpuPercent) > 0 {
		metrics.CPUPercent = cpuPercent[0]
	}

	// CPU cores
	cores, err := cpu.Counts(true)
	if err != nil {
		return err
	}
	metrics.CPUCores = cores

	return nil
}

// collectMemory collects memory metrics
func (c *Collector) collectMemory(metrics *Metrics) error {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	metrics.RAMPercent = memInfo.UsedPercent
	metrics.RAMTotal = memInfo.Total
	metrics.RAMUsed = memInfo.Used

	return nil
}

// collectDisk collects disk metrics
func (c *Collector) collectDisk(metrics *Metrics) error {
	diskInfo, err := disk.Usage(c.diskPath)
	if err != nil {
		return err
	}

	metrics.DiskPercent = diskInfo.UsedPercent
	metrics.DiskTotal = diskInfo.Total
	metrics.DiskUsed = diskInfo.Used

	return nil
}

// collectLoadAverage collects system load average
func (c *Collector) collectLoadAverage(metrics *Metrics) error {
	loadInfo, err := load.Avg()
	if err != nil {
		return err
	}

	metrics.LoadAverage = []float64{
		loadInfo.Load1,
		loadInfo.Load5,
		loadInfo.Load15,
	}

	return nil
}

// collectHostInfo collects host information
func (c *Collector) collectHostInfo(metrics *Metrics) error {
	hostInfo, err := host.Info()
	if err != nil {
		return err
	}

	metrics.Uptime = hostInfo.Uptime
	metrics.ProcessCount = hostInfo.Procs

	return nil
}

// collectNetwork collects network metrics
func (c *Collector) collectNetwork(metrics *Metrics) error {
	netIO, err := net.IOCounters(false)
	if err != nil {
		return err
	}

	if len(netIO) > 0 {
		metrics.NetworkSent = netIO[0].BytesSent
		metrics.NetworkRecv = netIO[0].BytesRecv
	}

	return nil
}
