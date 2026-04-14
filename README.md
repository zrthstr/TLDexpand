# TLDexpand

**Fast, concurrent TLD scanner written in Go**

TLDexpand scans for all existing top-level domains for a given domain name using concurrent DNS lookups. Built for speed and simplicity.

## Features

- **Blazing Fast**: Concurrent DNS lookups using goroutines
- **Zero Dependencies**: Single static binary, no runtime required
- **Lightweight**: ~5-10MB Docker image
- **Pre-cached TLD List**: Includes full IANA TLD list (1400+ TLDs)
- **Multiple Output Formats**: JSON, JSON array, or plain text
- **Cross-platform**: Runs on Linux, macOS, Windows

## Quick Start

### Option 1: Build from Source

```bash
git clone https://github.com/l159375751/TLDexpand.git
cd TLDexpand
make build
./tldexpand -d google
```

### Option 2: Docker

```bash
git clone https://github.com/l159375751/TLDexpand.git
cd TLDexpand
make docker
docker run --rm tldexpand:latest -d google -i ccTLDs.txt
```

## Usage

```bash
tldexpand [options]

Options:
  -d string
        Domain name to scan (e.g., 'google')
  -o string
        Output file path
  -i string
        Custom TLD list file
  -m string
        Output mode: json, jsonarray, plain (default "json")
  -f    Use full IANA TLD list (default: uses ccTLDs.txt)
  -w int
        Number of concurrent workers (default 100)
```

## Examples

### Scan with country code TLDs (fast)
```bash
tldexpand -d google -i ccTLDs.txt
```

### Scan with full IANA TLD list
```bash
tldexpand -d google -f
```

### Save results to file
```bash
tldexpand -d github -f -o results.json
```

### Plain text output
```bash
tldexpand -d microsoft -i topTLDs.txt -m plain
```

### Increase concurrency
```bash
tldexpand -d amazon -f -w 200
```

## Output Modes

### JSON (default)
Returns a JSON object mapping domains to IP addresses:
```json
{
  "google.com": "172.217.168.46",
  "google.net": "172.217.168.4",
  "google.org": "216.239.32.27"
}
```

### JSON Array
Returns a JSON array of found domains:
```json
[
  "google.com",
  "google.net",
  "google.org"
]
```

### Plain Text
Returns one domain per line:
```
google.com
google.net
google.org
```

## TLD Lists Included

- **ccTLDs.txt**: Country code TLDs (~250 domains) - **Default**
- **sTLDs.txt**: Special TLDs (small set)
- **topTLDs.txt**: Top 24 most common TLDs
- **tld_scanner_list.txt**: Full IANA list (1400+ TLDs) - Use with `-f` flag

## Development

### Run Tests
```bash
make test
```

### Run Benchmarks
```bash
make bench
```

### Cross-Compile for All Platforms
```bash
make cross
```

This creates binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

### Build Docker Image
```bash
make docker
```

## Performance

TLDexpand uses goroutines for concurrent DNS lookups, making it significantly faster than sequential scanning:

- **100 workers (default)**: ~24 TLDs in 0.06s
- **200 workers**: Even faster for large TLD lists
- **Full IANA scan**: 1400+ TLDs in seconds

Adjust worker count with `-w` flag based on your network and system.

## License

See LICENSE file for details.

## Original Project

This is a modern Go rewrite of [tld_scanner](https://github.com/ozzi-/tld_scanner) by ozzi-.
