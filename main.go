package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileInfo holds metadata about a file before it gets anonymized.
type FileInfo struct {
	OriginalPath string
	Extension    string
	SourceDir    string
}

// IndexEntry represents a single entry in the index.yaml output file.
type IndexEntry struct {
	Anonymized string `yaml:"anonymized"`
	Original   string `yaml:"original"`
	SourceDir  string `yaml:"source_dir"`
}

// IndexFile is the top-level structure written to index.yaml.
type IndexFile struct {
	Files []IndexEntry `yaml:"files"`
}

// AssessmentEntry represents a single line from the assessment input file.
type AssessmentEntry struct {
	File     string
	Dir      string
	DirValid bool
}

const (
	anonymizedNameLength = 6
	alphanumericChars    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

const version = "1.1.1"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	flag := os.Args[1]

	switch flag {
	case "--version", "-v":
		fmt.Printf("fuffle %s\n", version)
	case "--mix":
		runMix(os.Args[2:])
	case "--assessment":
		runAssessment(os.Args[2:])
	case "--report":
		runReport(os.Args[2:])
	case "session":
		runSession(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown option: %s\n\n", flag)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s --version, -v\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s --mix <dir1> [dir2] [dir3] ...\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s --assessment <file.txt>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s --report <evaluations.yaml> [--output <file>] [--serve [:port]]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s session new\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s session insert -f file.py --start 123 --end 456 [--tags ok,tag2]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s session starttime\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s session endtime\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  --version, -v  Print version\n")
	fmt.Fprintf(os.Stderr, "  --mix          Shuffle files from directories into ./mixed/ with anonymized names\n")
	fmt.Fprintf(os.Stderr, "  --assessment   Check folder existence from a file listing and print report\n")
	fmt.Fprintf(os.Stderr, "  --report       Generate report from evaluation YAML file\n")
	fmt.Fprintf(os.Stderr, "    --output       Output file (.svg or .html, default: report.svg)\n")
	fmt.Fprintf(os.Stderr, "    --serve [port] Serve report via HTTP and open browser (default: 8080)\n")
	fmt.Fprintf(os.Stderr, "  session        Manage evaluation sessions\n")
	fmt.Fprintf(os.Stderr, "    new            Create new session.yaml\n")
	fmt.Fprintf(os.Stderr, "    insert         Add entry to session.yaml\n")
	fmt.Fprintf(os.Stderr, "    starttime      Print session start time (date +%s format)\n")
	fmt.Fprintf(os.Stderr, "    endtime        Print session end time (date +%s format)\n")
}

// runMix handles the --mix mode: shuffles files from given directories into ./mixed/.
func runMix(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: --mix requires at least one directory argument\n")
		fmt.Fprintf(os.Stderr, "Usage: %s --mix <dir1> [dir2] [dir3] ...\n", os.Args[0])
		os.Exit(1)
	}

	dirs := args

	// Validate all directories exist and are actually directories.
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot access directory %q: %v\n", dir, err)
			os.Exit(1)
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "Error: %q is not a directory\n", dir)
			os.Exit(1)
		}
	}

	// Collect all files recursively from the given directories.
	var files []FileInfo
	for _, dir := range dirs {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			files = append(files, FileInfo{
				OriginalPath: path,
				Extension:    ext,
				SourceDir:    dir,
			})
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory %q: %v\n", dir, err)
			os.Exit(1)
		}
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no files found in the provided directories\n")
		os.Exit(1)
	}

	// Clean up existing output.
	mixedDir := "mixed"
	if _, err := os.Stat(mixedDir); err == nil {
		if err := os.RemoveAll(mixedDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing existing %q directory: %v\n", mixedDir, err)
			os.Exit(1)
		}
	}
	indexFile := "index.yaml"
	if err := os.Remove(indexFile); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error removing existing %q: %v\n", indexFile, err)
		os.Exit(1)
	}

	// Create the mixed directory.
	if err := os.MkdirAll(mixedDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating %q directory: %v\n", mixedDir, err)
		os.Exit(1)
	}

	// Shuffle the files using Fisher-Yates via the standard library.
	shuffleFiles(files)

	// Track used names to avoid collisions.
	usedNames := make(map[string]bool)
	var indexEntries []IndexEntry

	// Copy each file with an anonymized name.
	for _, f := range files {
		name, err := generateAnonymizedName(f.Extension, usedNames)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating anonymized name: %v\n", err)
			os.Exit(1)
		}

		dstPath := filepath.Join(mixedDir, name)
		if err := copyFile(f.OriginalPath, dstPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error copying %q to %q: %v\n", f.OriginalPath, dstPath, err)
			os.Exit(1)
		}

		indexEntries = append(indexEntries, IndexEntry{
			Anonymized: name,
			Original:   f.OriginalPath,
			SourceDir:  f.SourceDir,
		})
	}

	// Write index.yaml.
	idx := IndexFile{Files: indexEntries}
	data, err := yaml.Marshal(idx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling index: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(indexFile, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %q: %v\n", indexFile, err)
		os.Exit(1)
	}

	fmt.Printf("Shuffled %d files from %d directories into ./%s/\n", len(files), len(dirs), mixedDir)
	fmt.Printf("Index written to ./%s\n", indexFile)
}

// runAssessment handles the --assessment mode: reads a file and checks directory existence.
func runAssessment(args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Error: --assessment requires exactly one file argument\n")
		fmt.Fprintf(os.Stderr, "Usage: %s --assessment <file.txt>\n", os.Args[0])
		os.Exit(1)
	}

	filePath := args[0]

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot open file %q: %v\n", filePath, err)
		os.Exit(1)
	}
	defer f.Close()

	// Parse each line: <filepath> <directory>
	var entries []AssessmentEntry
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "Warning: line %d has invalid format, skipping: %s\n", lineNum, line)
			continue
		}

		file := parts[0]
		dir := parts[1]

		// Check if the directory exists.
		info, err := os.Stat(dir)
		valid := err == nil && info.IsDir()

		entries = append(entries, AssessmentEntry{
			File:     file,
			Dir:      dir,
			DirValid: valid,
		})
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %q: %v\n", filePath, err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no valid entries found in %q\n", filePath)
		os.Exit(1)
	}

	// Print report.
	var validCount int
	for _, e := range entries {
		status := "OK"
		if !e.DirValid {
			status = "MISSING"
		} else {
			validCount++
		}
		fmt.Printf("  [%s] %s -> %s\n", status, e.File, e.Dir)
	}

	total := len(entries)
	percentage := float64(validCount) / float64(total) * 100

	fmt.Println()
	fmt.Printf("Assessment report for: %s\n", filePath)
	fmt.Printf("  Total entries:  %d\n", total)
	fmt.Printf("  Valid dirs:     %d\n", validCount)
	fmt.Printf("  Missing dirs:   %d\n", total-validCount)
	fmt.Printf("  Correctness:    %.1f%%\n", percentage)
}

// runSession handles the session subcommand: new or insert.
func runSession(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: session requires a subcommand\n")
		fmt.Fprintf(os.Stderr, "Usage: %s session new\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s session insert -r \"result\" -f file.py --start 123 --end 456 [--tags ok,tag2]\n", os.Args[0])
		os.Exit(1)
	}

	sub := args[0]
	switch sub {
	case "new":
		sessionNew()
	case "insert":
		sessionInsert(args[1:])
	case "starttime":
		sessionStarttime()
	case "endtime":
		sessionEndtime()
	default:
		fmt.Fprintf(os.Stderr, "Unknown session subcommand: %s\n", sub)
		os.Exit(1)
	}
}

// runReport handles the --report mode: generates SVG/HTML report from evaluation YAML.
func runReport(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: --report requires at least one argument\n")
		fmt.Fprintf(os.Stderr, "Usage: %s --report <evaluations.yaml> [--output <file>] [--serve [:port]]\n", os.Args[0])
		os.Exit(1)
	}

	inputPath := args[0]
	outputPath := ""
	serveAddr := ""
	serveMode := false

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--output":
			if i+1 < len(args) {
				outputPath = args[i+1]
				i++
			}
		case "--serve":
			serveMode = true
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				serveAddr = args[i+1]
				i++
			}
		}
	}

	eval, err := loadEvaluation(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var netdata *NetdataMetrics
	if eval.NetdataFile != "" {
		netdata, err = loadNetdata(eval.NetdataFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot load netdata file: %v\n", err)
		}
	}

	var hardware *HardwareInfo
	if eval.BigfetchFile != "" {
		hardware, err = loadBigfetch(eval.BigfetchFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot load bigfetch file: %v\n", err)
		}
	}

	if serveMode {
		html := generateHTML(eval, netdata, hardware)
		if err := serveHTML(html, serveAddr); err != nil {
			fmt.Fprintf(os.Stderr, "Error serving report: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if outputPath == "" {
		outputPath = "report.svg"
	}

	if strings.HasSuffix(outputPath, ".html") {
		html := generateHTML(eval, netdata, hardware)
		if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing HTML: %v\n", err)
			os.Exit(1)
		}
	} else {
		svg := generateSVG(eval, netdata)
		if err := os.WriteFile(outputPath, []byte(svg), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing SVG: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Report generated: %s\n", outputPath)
}

// shuffleFiles randomizes the order of the slice in place.
func shuffleFiles(files []FileInfo) {
	n := len(files)
	for i := n - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			panic(fmt.Sprintf("crypto/rand failed: %v", err))
		}
		j := int(jBig.Int64())
		files[i], files[j] = files[j], files[i]
	}
}

// generateAnonymizedName creates a random 6-character alphanumeric name with the given extension.
// It retries on collisions up to a reasonable limit.
func generateAnonymizedName(ext string, used map[string]bool) (string, error) {
	const maxAttempts = 1000
	for attempt := 0; attempt < maxAttempts; attempt++ {
		b := make([]byte, anonymizedNameLength)
		for i := range b {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphanumericChars))))
			if err != nil {
				return "", fmt.Errorf("generating random char: %w", err)
			}
			b[i] = alphanumericChars[idx.Int64()]
		}
		name := string(b) + ext
		if !used[name] {
			used[name] = true
			return name, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique name after %d attempts", maxAttempts)
}

// copyFile copies a single file from src to dst, preserving permissions.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
