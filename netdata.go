package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type NetdataRow struct {
	Timestamp int64
	CPU       float64
	RAM       float64
	GPU       float64
	Disk      float64
}

type NetdataMetrics struct {
	Timestamps []int64
	CPU        []float64
	RAM        []float64
	GPU        []float64
	Disk       []float64
}

func loadNetdata(path string) (*NetdataMetrics, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open netdata file %q: %w", path, err)
	}

	if len(data) > 0 && data[0] == '{' {
		var probe struct {
			Labels []interface{} `json:"labels"`
			Data   []interface{} `json:"data"`
		}
		if err := json.Unmarshal(data, &probe); err == nil && len(probe.Labels) > 0 && len(probe.Data) > 0 {
			return parseNetdataLabelsJSON(data)
		}
		return parseNetdataJSON(data)
	}
	return parseNetdataCSV(data)
}

func parseNetdataJSON(data []byte) (*NetdataMetrics, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("cannot parse netdata JSON: %w", err)
	}

	metrics := &NetdataMetrics{}

	getValue := func(chartName string) float64 {
		entry, ok := raw[chartName]
		if !ok {
			return 0
		}
		var obj struct {
			Dimensions map[string]struct {
				Value float64 `json:"value"`
			} `json:"dimensions"`
		}
		if err := json.Unmarshal(entry, &obj); err != nil {
			return 0
		}
		total := 0.0
		for _, d := range obj.Dimensions {
			total += d.Value
		}
		return total
	}

	cpu := getValue("system.cpu")
	ram := getValue("system.ram")
	gpu := getValue("gpu.cuda_gpu")
	disk := getValue("system.disk")

	metrics.Timestamps = append(metrics.Timestamps, 0)
	metrics.CPU = append(metrics.CPU, cpu)
	metrics.RAM = append(metrics.RAM, ram)
	metrics.GPU = append(metrics.GPU, gpu)
	metrics.Disk = append(metrics.Disk, disk)

	return metrics, nil
}

func parseNetdataLabelsJSON(data []byte) (*NetdataMetrics, error) {
	var raw struct {
		Labels []string    `json:"labels"`
		Data   [][]float64 `json:"data"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("cannot parse netdata labels JSON: %w", err)
	}

	if len(raw.Labels) == 0 || len(raw.Data) == 0 {
		return nil, fmt.Errorf("empty netdata labels data")
	}

	colIndex := make(map[string]int)
	for i, l := range raw.Labels {
		colIndex[l] = i
	}

	timeIdx, ok := colIndex["time"]
	if !ok {
		return nil, fmt.Errorf("no 'time' column in netdata labels")
	}

	cpuCols := []string{"user", "system", "nice", "iowait", "irq", "softirq", "steal", "guest", "guest_nice"}
	ramIdx, hasRam := colIndex["ram"]
	gpuIdx, hasGpu := colIndex["gpu"]
	diskIdx, hasDisk := colIndex["disk"]

	metrics := &NetdataMetrics{}
	for _, row := range raw.Data {
		if len(row) <= int(timeIdx) {
			continue
		}
		ts := int64(row[timeIdx])
		metrics.Timestamps = append(metrics.Timestamps, ts)

		cpu := 0.0
		for _, c := range cpuCols {
			if idx, ok := colIndex[c]; ok && int(idx) < len(row) {
				cpu += row[idx]
			}
		}
		metrics.CPU = append(metrics.CPU, cpu)

		ram := 0.0
		if hasRam && int(ramIdx) < len(row) {
			ram = row[ramIdx]
		}
		metrics.RAM = append(metrics.RAM, ram)

		gpu := 0.0
		if hasGpu && int(gpuIdx) < len(row) {
			gpu = row[gpuIdx]
		}
		metrics.GPU = append(metrics.GPU, gpu)

		disk := 0.0
		if hasDisk && int(diskIdx) < len(row) {
			disk = row[diskIdx]
		}
		metrics.Disk = append(metrics.Disk, disk)
	}

	return metrics, nil
}

func parseNetdataCSV(data []byte) (*NetdataMetrics, error) {
	metrics := &NetdataMetrics{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || lineNum == 1 {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 5 {
			continue
		}
		ts, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}
		cpu, _ := strconv.ParseFloat(parts[1], 64)
		ram, _ := strconv.ParseFloat(parts[2], 64)
		gpu, _ := strconv.ParseFloat(parts[3], 64)
		disk, _ := strconv.ParseFloat(parts[4], 64)

		metrics.Timestamps = append(metrics.Timestamps, ts)
		metrics.CPU = append(metrics.CPU, cpu)
		metrics.RAM = append(metrics.RAM, ram)
		metrics.GPU = append(metrics.GPU, gpu)
		metrics.Disk = append(metrics.Disk, disk)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading netdata file: %w", err)
	}
	return metrics, nil
}
