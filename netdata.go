package main

import (
	"bufio"
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
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open netdata file %q: %w", path, err)
	}
	defer f.Close()

	metrics := &NetdataMetrics{}
	scanner := bufio.NewScanner(f)
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
