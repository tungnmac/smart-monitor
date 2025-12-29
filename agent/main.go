package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type Stats struct {
	Hostname string  `json:"hostname"`
	CPU      float64 `json:"cpu"`
	RAM      float64 `json:"ram"`
}

func main() {
	ticker := time.NewTicker(2 * time.Second)
	for range ticker.C {
		c, _ := cpu.Percent(0, false)
		m, _ := mem.VirtualMemory()

		data := Stats{
			Hostname: "Server-01",
			CPU:      c[0],
			RAM:      m.UsedPercent,
		}

		jsonData, _ := json.Marshal(data)
		http.Post("http://localhost:8080/report", "application/json", bytes.NewBuffer(jsonData))
	}
}
