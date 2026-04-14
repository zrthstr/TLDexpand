# TLDexpand

Concurrent TLD scanner in Go. Tests which top-level domains exist for a given domain name.

Vibe-coded, if not mindlessly.

## Usage

```bash
tldexpand <domain> <tld-file> <resolver>
```

**Arguments:**
- `domain` - Domain name to scan (e.g., 'google')
- `tld-file` - TLD list file ('tlds' or 'cctlds')
- `resolver` - DNS resolver (e.g., '8.8.8.8:53')

**Output:** One domain per line to stdout

## Examples

```bash
# Scan with country code TLDs
./tldexpand google cctlds 8.8.8.8:53

# Scan with full IANA list
./tldexpand github tlds 1.1.1.1:53

# Save to file
./tldexpand amazon tlds 8.8.8.8:53 > results.txt
```

## Update TLD List

```bash
# Fetch IANA list and remove wildcard TLDs
./tldexpand --update > tlds
```

Fetches from IANA, tests all TLDs for wildcarding, removes false positives.

## Install

```bash
go install github.com/zrthstr/TLDexpand@latest
```

Or build from source:

```bash
git clone https://github.com/zrthstr/TLDexpand.git
cd TLDexpand
go build -o tldexpand
```

## TLD Lists

- **tlds** - Full IANA list (1429 TLDs, wildcards removed)
- **cctlds** - Country code TLDs (251 TLDs)

## Notes

- 150 concurrent workers
- Runtime wildcard filtering (auto-detects false positives)
- Pure stdlib (no external dependencies)
- ~3.4MB binary

## Original

Go rewrite of [tld_scanner](https://github.com/ozzi-/tld_scanner) by ozzi-.
