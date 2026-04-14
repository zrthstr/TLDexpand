package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	ianaURL         = "https://data.iana.org/TLD/tlds-alpha-by-domain.txt"
	numRandomTests  = 3
	wildcardMinHits = 2
)

func updateTLDs(resolver string) error {
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
	fmt.Fprintf(os.Stderr, "Scanning for wildcard TLDs using resolver %s...\n", resolver)

	// Filter out wildcard TLDs
	cleanTLDs, wildcardTLDs := filterWildcards(rawTLDs, resolver)

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

	// Write clean TLD list to stdout
	fmt.Printf("# TLD list from IANA (wildcard TLDs removed)\n")
	fmt.Printf("# Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("# Source: %s\n", ianaURL)
	fmt.Printf("# Wildcards removed: %d\n", len(wildcardTLDs))

	for _, tld := range cleanTLDs {
		fmt.Println(strings.ToUpper(tld))
	}

	return nil
}

type tldResult struct {
	tld        string
	isWildcard bool
}

func filterWildcards(tlds []string, resolverAddr string) (clean []string, wildcards []string) {
	// Random test strings
	randomTests := []string{
		"this-domain-absolutely-does-not-exist-12345678",
		"completely-random-nonexistent-garbage-99999",
		"xyzabc123-fake-test-domain-should-not-resolve",
	}

	jobs := make(chan string, len(tlds))
	results := make(chan tldResult, workers*2)
	var wg sync.WaitGroup

	// Create custom resolver
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second * 10,
			}
			return d.DialContext(ctx, network, resolverAddr)
		},
	}

	// Start result collector
	var collectorWg sync.WaitGroup
	collectorWg.Add(1)
	processed := 0
	total := len(tlds)

	go func() {
		defer collectorWg.Done()
		for result := range results {
			processed++
			if processed%50 == 0 || processed == total {
				fmt.Fprintf(os.Stderr, "Progress: %d/%d TLDs checked\r", processed, total)
			}
			if result.isWildcard {
				wildcards = append(wildcards, result.tld)
			} else {
				clean = append(clean, result.tld)
			}
		}
	}()

	// Start worker pool
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go wildcardWorker(jobs, results, &wg, resolver, randomTests)
	}

	// Send jobs
	for _, tld := range tlds {
		jobs <- tld
	}
	close(jobs)

	wg.Wait()
	close(results)
	collectorWg.Wait()

	fmt.Fprintf(os.Stderr, "Progress: %d/%d TLDs checked\n", total, total)
	return clean, wildcards
}

func wildcardWorker(jobs <-chan string, results chan<- tldResult, wg *sync.WaitGroup, resolver *net.Resolver, randomTests []string) {
	defer wg.Done()

	for tld := range jobs {
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
		results <- tldResult{
			tld:        tld,
			isWildcard: resolveCount >= wildcardMinHits,
		}
	}
}
