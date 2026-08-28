package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type BigfetchData struct {
	Fastfetch []FastfetchEntry `json:"fastfetch"`
	Dmidecode []DmidecodeEntry `json:"dmidecode"`
}

type FastfetchEntry struct {
	Type   string          `json:"type"`
	Result json.RawMessage `json:"result"`
}

type DmidecodeEntry struct {
	Description string                 `json:"description"`
	Values      map[string]interface{} `json:"values"`
}

type CPUInfo struct {
	Model   string
	Cores   int
	Threads int
}

type GPUInfo struct {
	Model string
	Count int
	VRAM  string
}

type RAMInfo struct {
	Model      string
	Clock      string
	Technology string
}

type DiskInfo struct {
	Model string
	Type  string
	Speed string
}

type HardwareInfo struct {
	CPU  CPUInfo
	GPU  []GPUInfo
	RAM  []RAMInfo
	Disk []DiskInfo
}

func loadBigfetch(path string) (*HardwareInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read bigfetch file %q: %w", path, err)
	}

	var bigfetch BigfetchData
	if err := json.Unmarshal(data, &bigfetch); err != nil {
		return nil, fmt.Errorf("cannot parse bigfetch JSON: %w", err)
	}

	info := &HardwareInfo{}

	for _, entry := range bigfetch.Fastfetch {
		switch entry.Type {
		case "CPU":
			var cpu struct {
				CPU   string `json:"cpu"`
				Cores struct {
					Physical int `json:"physical"`
					Logical  int `json:"logical"`
				} `json:"cores"`
			}
			if err := json.Unmarshal(entry.Result, &cpu); err == nil {
				info.CPU = CPUInfo{
					Model:   cpu.CPU,
					Cores:   cpu.Cores.Physical,
					Threads: cpu.Cores.Logical,
				}
			}
		case "GPU":
			var gpus []struct {
				Name   string `json:"name"`
				Vendor string `json:"vendor"`
				Memory struct {
					Dedicated struct {
						Total *int64 `json:"total"`
					} `json:"dedicated"`
					Shared struct {
						Total *int64 `json:"total"`
					} `json:"shared"`
				} `json:"memory"`
			}
			if err := json.Unmarshal(entry.Result, &gpus); err == nil {
				for _, g := range gpus {
					gpu := GPUInfo{
						Model: g.Name,
					}
					if g.Memory.Dedicated.Total != nil {
						gpu.VRAM = formatBytes(*g.Memory.Dedicated.Total)
					} else if g.Memory.Shared.Total != nil {
						gpu.VRAM = formatBytes(*g.Memory.Shared.Total) + " (shared)"
					}
					info.GPU = append(info.GPU, gpu)
				}
			}
		case "Disk":
			var disks []struct {
				MountFrom string `json:"mountFrom"`
				Bytes     struct {
					Total int64 `json:"total"`
				} `json:"bytes"`
				Filesystem string `json:"filesystem"`
			}
			if err := json.Unmarshal(entry.Result, &disks); err == nil {
				for _, d := range disks {
					disk := DiskInfo{
						Model: d.MountFrom,
						Type:  detectDiskType(d.MountFrom),
					}
					info.Disk = append(info.Disk, disk)
				}
			}
		}
	}

	for _, entry := range bigfetch.Dmidecode {
		if entry.Description == "Memory Device" {
			ram := RAMInfo{}
			if size, ok := entry.Values["size"].(string); ok {
				ram.Model = size
			}
			if speed, ok := entry.Values["speed"].(string); ok {
				ram.Clock = speed
			}
			if typ, ok := entry.Values["type"].(string); ok {
				ram.Technology = typ
			}
			if locator, ok := entry.Values["locator"].(string); ok {
				if partNum, ok := entry.Values["part_number"].(string); ok {
					ram.Model = strings.TrimSpace(partNum) + " (" + strings.TrimSpace(locator) + ")"
				}
			}
			info.RAM = append(info.RAM, ram)
		}
	}

	return info, nil
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.0f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.0f MB", float64(bytes)/float64(MB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

func detectDiskType(mountFrom string) string {
	if strings.Contains(mountFrom, "nvme") {
		return "NVMe"
	}
	if strings.Contains(mountFrom, "sd") {
		return "SSD"
	}
	if strings.Contains(mountFrom, "hd") {
		return "HDD"
	}
	return "Unknown"
}
