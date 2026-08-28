package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type EvaluationEntry struct {
	File      string   `yaml:"file"`
	StartDate int64    `yaml:"start_date"`
	EndDate   int64    `yaml:"end_date"`
	Tags      []string `yaml:"tags"`
	Result    string   `yaml:"result,omitempty"`
}

type Extra struct {
	Type        string `yaml:"type"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Body        string `yaml:"body"`
}

type EvaluationFile struct {
	Title        string            `yaml:"title"`
	Extras       []Extra           `yaml:"extras"`
	NetdataFile  string            `yaml:"netdatafile"`
	BigfetchFile string            `yaml:"bigfetchfile"`
	Entries      []EvaluationEntry `yaml:"entries"`
}

func loadEvaluation(path string) (*EvaluationFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %q: %w", path, err)
	}
	var eval EvaluationFile
	if err := yaml.Unmarshal(data, &eval); err != nil {
		return nil, fmt.Errorf("cannot parse YAML: %w", err)
	}
	if len(eval.Entries) == 0 {
		return nil, fmt.Errorf("no entries found in %q", path)
	}
	return &eval, nil
}

func entryDuration(e EvaluationEntry) float64 {
	return float64(e.EndDate - e.StartDate)
}

func totalDuration(entries []EvaluationEntry) (int64, int64) {
	if len(entries) == 0 {
		return 0, 0
	}
	min := entries[0].StartDate
	max := entries[0].EndDate
	for _, e := range entries[1:] {
		if e.StartDate < min {
			min = e.StartDate
		}
		if e.EndDate > max {
			max = e.EndDate
		}
	}
	return min, max
}

func calcAccuracyOverTime(entries []EvaluationEntry) []float64 {
	if len(entries) == 0 {
		return nil
	}
	sorted := make([]EvaluationEntry, len(entries))
	copy(sorted, entries)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].StartDate < sorted[i].StartDate {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	result := make([]float64, len(sorted))
	correct := 0
	for i, e := range sorted {
		if len(e.Tags) > 0 {
			correct++
		}
		result[i] = float64(correct) / float64(i+1) * 100
	}
	return result
}
