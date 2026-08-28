# fuffle
<img width="1024" height="559" alt="image" src="https://github.com/user-attachments/assets/60b8c168-1f42-4d70-aae7-1e3cc684e5d9" />


**File Shuffle** — A command-line tool for shuffling files and assessing directory structures.

## Modes

### `--mix` — Shuffle files

Takes one or more directories as input, shuffles all their files together, and writes them to a `mixed/` folder with randomized 6-character alphanumeric names. An `index.yaml` file is generated to map each anonymized name back to its original location.

```bash
fuffle --mix <dir1> [dir2] [dir3] ...
```

#### Example

```bash
fuffle --mix photos/ documents/ backups/
```

This will:

1. Recursively scan `photos/`, `documents/`, and `backups/` for files
2. Shuffle all files in random order
3. Write them to `./mixed/` with anonymized names (e.g., `a3x9Bf.png`, `kR7mZ2.pdf`)
4. Generate `./index.yaml` with the original path mapping

#### Output structure

```
./
├── mixed/
│   ├── a3x9Bf.png
│   ├── kR7mZ2.pdf
│   └── ...
└── index.yaml
```

#### Index format

```yaml
files:
    - anonymized: a3x9Bf.png
      original: photos/vacation/beach.png
      source_dir: photos
    - anonymized: kR7mZ2.pdf
      original: documents/report.pdf
      source_dir: documents
```

---

### `--assessment` — Check directory structure

Reads a text file where each line contains a file path and a directory separated by a space. Checks if each directory exists and prints a correctness report.

```bash
fuffle --assessment <file.txt>
```

#### Input file format

Each line: `<filepath> <directory>` (no spaces in paths)

```
photo1.jpg /home/user/photos
doc.pdf /home/user/documents
backup.zip /nonexistent/path
```

#### Example output

```
  [OK] photo1.jpg -> /home/user/photos
  [OK] doc.pdf -> /home/user/documents
  [MISSING] backup.zip -> /nonexistent/path

Assessment report for: listing.txt
  Total entries:  3
  Valid dirs:     2
  Missing dirs:   1
  Correctness:    66.7%
```

---

### `--report` — Generate report

Reads a YAML evaluation file and generates a report (SVG or HTML) with statistics, charts, and system metrics.

```bash
fuffle --report <evaluations.yaml> [--output <file>] [--serve [:port]]
```

#### Evaluation file format

```yaml
title: "My Evaluation"
extras:
  - type: text
    title: "Note"
    description: "Additional context"
    body: "This is extra text content"
  - type: image
    title: "Chart"
    description: "Supporting visualization"
    body: "./img/chart.png"
netdatafile: "path/to/netdata.csv"
entries:
  - file: "mixed/a3x9Bf.png"
    start_date: 1693123456789    # Unix timestamp in milliseconds
    end_date: 1693123457890      # Unix timestamp in milliseconds
    tags:
      - "prompt1"
  - file: "mixed/kR7mZ2.pdf"
    start_date: 1693123470000
    end_date: 1693123475000
    tags:
      - "prompt1"
```

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Report title |
| `extras` | []Extra | Extra content cards (text or image) |
| `netdatafile` | string | Path to netdata monitoring CSV file (optional) |
| `entries` | []EvaluationEntry | List of evaluation entries |

#### Extra fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `"text"` or `"image"` |
| `title` | string | Card title |
| `description` | string | Short description (optional) |
| `body` | string | Text content or image path (relative to CWD) |

#### Evaluation entry fields

| Field | Type | Description |
|-------|------|-------------|
| `file` | string | Path to the file being evaluated |
| `start_date` | int64 | Start timestamp in milliseconds (`date +%s%3N`) |
| `end_date` | int64 | End timestamp in milliseconds (`date +%s%3N`) |
| `tags` | []string | Tags for this entry (first tag used for OK/FAIL status) |

#### Netdata file format (CSV)

```csv
timestamp,cpu,ram,gpu,disk
1693123456,45.2,1024,78.5,12.3
1693123457,46.1,1028,79.2,12.5
```

#### Report contents

1. **Results table** - File, status (OK/FAIL), processing time, tag
2. **Statistics** - Total files, total time, mean time (+/- SD), median, mode
3. **Accuracy chart** - Accuracy over time (cumulative)
4. **System metrics** - If netdata file provided: CPU, RAM, GPU, Disk stats and time chart
5. **Extras** - Cards with text or image content

#### Output options

- `--output <file.svg>` - Generate SVG file (default: `report.svg`)
- `--output <file.html>` - Generate HTML file with embedded CSS
- `--serve [:port]` - Serve report via HTTP and open browser (default port: 8080)

#### Examples

```bash
# Generate SVG (default)
fuffle --report evaluation.yaml --output report.svg

# Generate HTML
fuffle --report evaluation.yaml --output report.html

# Serve and open in browser
fuffle --report evaluation.yaml --serve

# Serve on custom port
fuffle --report evaluation.yaml --serve :3000
```

---

## Behavior

- **Overwrite mode** (`--mix`): If `mixed/` or `index.yaml` already exist, they are replaced
- **Error on missing dirs** (`--mix`): Exits with an error if any provided directory does not exist
- **Error on empty** (`--mix`): Exits with an error if no files are found across all directories
- **Recursive** (`--mix`): Searches all subdirectories within the provided paths
- **Collision-safe** (`--mix`): Generates unique names; retries if a name collision occurs

## Build

```bash
go build -o fuffle .
```

## Dependencies

- Go standard library
- `gopkg.in/yaml.v3` (YAML marshaling)

## Report generation

The `--report` mode generates reports without any external dependencies. SVG is generated directly using Go's standard library. HTML reports include embedded CSS with a clean light-mode design and responsive layout.
