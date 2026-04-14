package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	ianaURL         = "https://data.iana.org/TLD/tlds-alpha-by-domain.txt"
	numRandomTests  = 3
	wildcardMinHits = 2
)

func updateTLDs(outputFile string) error {
	fmt.Fprintf(os.Stderr, "Fetching TLD list from IANA...\n")

	// Fetch IANA list
	resp, err := http.Get(ianaURL)
	if err != nil {
		return fmt.Errorf("failed to fetch IANA TLD list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("IANA returned status %d", resp.StatusCode)
	}

	// Parse TLD list
	var rawTLDs []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rawTLDs = append(rawTLDs, strings.ToLower(line))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read IANA response: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Fetched %d TLDs from IANA\n", len(rawTLDs))
	fmt.Fprintf(os.Stderr, "Scanning for wildcard TLDs (false positives)...\n")

	// Filter out wildcard TLDs
	cleanTLDs, wildcardTLDs := filterWildcards(rawTLDs)

	fmt.Fprintf(os.Stderr, "\nResults:\n")
	fmt.Fprintf(os.Stderr, "  Total TLDs fetched: %d\n", len(rawTLDs))
	fmt.Fprintf(os.Stderr, "  Wildcard TLDs (removed): %d\n", len(wildcardTLDs))
	fmt.Fprintf(os.Stderr, "  Clean TLDs (kept): %d\n\n", len(cleanTLDs))

	if len(wildcardTLDs) > 0 {
		fmt.Fprintf(os.Stderr, "Removed wildcard TLDs:\n")
		for _, tld := range wildcardTLDs {
			fmt.Fprintf(os.Stderr, "  .%s\n", tld)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Write clean TLD list
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	// Write header comment
	fmt.Fprintf(f, "# TLD list from IANA (wildcard TLDs removed)\n")
	fmt.Fprintf(f, "# Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "# Source: %s\n", ianaURL)
	fmt.Fprintf(f, "# Wildcards removed: %d\n", len(wildcardTLDs))

	// Write TLDs
	for _, tld := range cleanTLDs {
		fmt.Fprintln(f, strings.ToUpper(tld))
	}

	fmt.Fprintf(os.Stderr, "Wrote clean TLD list to: %s\n", outputFile)
	return nil
}

func filterWildcards(tlds []string) (clean []string, wildcards []string) {
	// Use Google DNS for consistent wildcard detection
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second * 10,
			}
			return d.DialContext(ctx, network, "8.8.8.8:53")
		},
	}

	// Random test strings
	randomTests := []string{
		"this-domain-absolutely-does-not-exist-12345678",
		"completely-random-nonexistent-garbage-99999",
		"xyzabc123-fake-test-domain-should-not-resolve",
	}

	total := len(tlds)
	for idx, tld := range tlds {
		if idx%50 == 0 {
			fmt.Fprintf(os.Stderr, "Progress: %d/%d TLDs checked\r", idx, total)
		}

		resolveCount := 0

		// Test random domains with individual timeout per lookup
		for _, random := range randomTests {
			testDomain := random + "." + tld
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err := resolver.LookupHost(ctx, testDomain)
			cancel()
			if err == nil {
				resolveCount++
			}
		}

		// If 2+ random domains resolve, it's wildcarding
		if resolveCount >= wildcardMinHits {
			wildcards = append(wildcards, tld)
		} else {
			clean = append(clean, tld)
		}
	}

	fmt.Fprintf(os.Stderr, "Progress: %d/%d TLDs checked\n", total, total)
	return clean, wildcards
}
