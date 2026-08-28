package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const sessionFile = "session.yaml"

type SessionInfo struct {
	Starttime int64 `yaml:"starttime,omitempty"`
	Endtime   int64 `yaml:"endtime,omitempty"`
}

type SessionFile struct {
	Title        string            `yaml:"title,omitempty"`
	BigfetchFile string            `yaml:"bigfetchfile,omitempty"`
	NetdataFile  string            `yaml:"netdatafile,omitempty"`
	Extras       []Extra           `yaml:"extras,omitempty"`
	Info         SessionInfo       `yaml:"info,omitempty"`
	Entries      []EvaluationEntry `yaml:"entries"`
}

func sessionNew() {
	if _, err := os.Stat(sessionFile); err == nil {
		fmt.Fprintf(os.Stderr, "Warning: %s already exists, overwriting\n", sessionFile)
	}

	session := SessionFile{
		Title:        "title",
		BigfetchFile: "./bigfetch.json",
		NetdataFile:  "./metrics.json",
		Extras:       []Extra{},
		Info:         SessionInfo{},
		Entries:      []EvaluationEntry{},
	}

	data, err := yaml.Marshal(session)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling session: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(sessionFile, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", sessionFile, err)
		os.Exit(1)
	}

	fmt.Printf("Created %s\n", sessionFile)
}

func sessionInsert(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: session insert requires flags\n")
		fmt.Fprintf(os.Stderr, "Usage: %s session insert -f file.py --start 123 --end 456 [--tags ok,tag2] [--result \"output\"]\n", os.Args[0])
		os.Exit(1)
	}

	var file, result string
	var start, end int64
	var tags []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-f":
			if i+1 < len(args) {
				file = args[i+1]
				i++
			}
		case "--result":
			if i+1 < len(args) {
				result = args[i+1]
				i++
			}
		case "--start":
			if i+1 < len(args) {
				v, err := strconv.ParseInt(args[i+1], 10, 64)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid start value: %v\n", err)
					os.Exit(1)
				}
				start = v
				i++
			}
		case "--end":
			if i+1 < len(args) {
				v, err := strconv.ParseInt(args[i+1], 10, 64)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid end value: %v\n", err)
					os.Exit(1)
				}
				end = v
				i++
			}
		case "--tags":
			if i+1 < len(args) {
				tags = strings.Split(args[i+1], ",")
				for j := range tags {
					tags[j] = strings.TrimSpace(tags[j])
				}
				i++
			}
		}
	}

	if file == "" {
		fmt.Fprintf(os.Stderr, "Error: -f is required\n")
		os.Exit(1)
	}
	if start == 0 || end == 0 {
		fmt.Fprintf(os.Stderr, "Error: --start and --end are required\n")
		os.Exit(1)
	}

	session, err := loadSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	entry := EvaluationEntry{
		File:      file,
		StartDate: start,
		EndDate:   end,
		Tags:      tags,
		Result:    result,
	}
	session.Entries = append(session.Entries, entry)

	if session.Info.Starttime == 0 || start < session.Info.Starttime {
		session.Info.Starttime = start
	}
	if end > session.Info.Endtime {
		session.Info.Endtime = end
	}

	data, err := yaml.Marshal(session)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling session: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(sessionFile, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", sessionFile, err)
		os.Exit(1)
	}

	fmt.Printf("Added entry to %s (%d total)\n", sessionFile, len(session.Entries))
}

func sessionTitle(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s session title \"new title\"\n", os.Args[0])
		os.Exit(1)
	}

	session, err := loadSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	session.Title = args[0]

	data, err := yaml.Marshal(session)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling session: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(sessionFile, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", sessionFile, err)
		os.Exit(1)
	}

	fmt.Printf("Title set to: %s\n", session.Title)
}

func sessionStarttime() {
	session, err := loadSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if session.Info.Starttime == 0 {
		fmt.Fprintf(os.Stderr, "No entries in session\n")
		os.Exit(1)
	}

	fmt.Printf("%d\n", session.Info.Starttime)
}

func sessionEndtime() {
	session, err := loadSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if session.Info.Endtime == 0 {
		fmt.Fprintf(os.Stderr, "No entries in session\n")
		os.Exit(1)
	}

	fmt.Printf("%d\n", session.Info.Endtime)
}

func loadSession() (*SessionFile, error) {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", sessionFile, err)
	}

	var session SessionFile
	if err := yaml.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", sessionFile, err)
	}

	return &session, nil
}
