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
