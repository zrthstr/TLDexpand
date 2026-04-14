package main

import (
	"context"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestLoadTLDs(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_tlds_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testTLDs := "# Comment line\nCOM\nNET\nORG\n\nEDU\n"
	if _, err := tmpFile.Write([]byte(testTLDs)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	tlds := loadTLDs(tmpFile.Name())

	if len(tlds) != 4 {
		t.Errorf("Expected 4 TLDs, got %d", len(tlds))
	}

	expected := []string{"com", "net", "org", "edu"}
	for i, tld := range tlds {
		if tld != expected[i] {
			t.Errorf("Expected TLD %s, got %s", expected[i], tld)
		}
	}
}

func TestWorker(t *testing.T) {
	jobs := make(chan string, 1)
	results := make(chan string, 1)
	var wg sync.WaitGroup

	// Use default resolver for test
	resolver := &net.Resolver{}

	jobs <- "invalid-tld-that-should-not-exist-12345"
	close(jobs)

	wg.Add(1)
	go worker("test", jobs, results, &wg, resolver)

	wg.Wait()
	close(results)

	count := 0
	for range results {
		count++
	}

	if count > 0 {
		t.Error("Should not get results for non-existent domain")
	}
}

func TestLoadRealTLDFile(t *testing.T) {
	// Test loading the actual tlds file
	if _, err := os.Stat("tlds"); err != nil {
		t.Skip("tlds file not found, skipping test")
	}

	tlds := loadTLDs("tlds")

	if len(tlds) < 1400 {
		t.Errorf("Expected at least 1400 TLDs in tlds file, got %d", len(tlds))
	}

	// Check for some known TLDs
	knownTLDs := []string{"com", "net", "org", "uk", "de", "app", "dev"}
	tldMap := make(map[string]bool)
	for _, tld := range tlds {
		tldMap[tld] = true
	}

	for _, known := range knownTLDs {
		if !tldMap[known] {
			t.Errorf("Expected TLD '%s' not found in tlds file", known)
		}
	}
}

func TestWildcardTLDs(t *testing.T) {
	// Find TLDs that wildcard (respond to ANY query)
	// These are false positives - they shouldn't be in our results
	//
	// Run with: go test -v -run TestWildcard
	// Full scan: go test -v -run TestWildcard -args -full

	if _, err := os.Stat("tlds"); err != nil {
		t.Skip("tlds file not found, skipping wildcard test")
	}

	// Check for -full flag
	fullScan := false
	for _, arg := range os.Args {
		if arg == "-full" {
			fullScan = true
			break
		}
	}

	tlds := loadTLDs("tlds")
	resolver := &net.Resolver{}
	ctx := context.Background()

	// Test with highly unlikely random strings
	randomTests := []string{
		"this-domain-absolutely-does-not-exist-12345678",
		"completely-random-nonexistent-garbage-99999",
		"xyzabc123-fake-test-domain-should-not-resolve",
	}

	wildcardTLDs := []string{}

	// Sample rate: every 20th TLD for quick test, all for full scan
	step := 20
	if fullScan {
		step = 1
		t.Logf("Running FULL wildcard scan on %d TLDs (this will take a while)...", len(tlds))
	} else {
		t.Logf("Running sample wildcard scan on ~%d TLDs (use -args -full for complete scan)...", len(tlds)/20)
	}

	for i := 0; i < len(tlds); i += step {
		tld := tlds[i]
		resolveCount := 0

		// Test 3 random domains on this TLD
		for _, random := range randomTests {
			testDomain := random + "." + tld
			_, err := resolver.LookupHost(ctx, testDomain)
			if err == nil {
				resolveCount++
			}
		}

		// If 2+ random domains resolve, it's likely wildcarding
		if resolveCount >= 2 {
			wildcardTLDs = append(wildcardTLDs, tld)
			t.Logf("WILDCARD DETECTED: .%s (%d/3 random domains resolved)", tld, resolveCount)
		}
	}

	if len(wildcardTLDs) > 0 {
		t.Errorf("Found %d wildcard TLDs (FALSE POSITIVES):", len(wildcardTLDs))
		for _, tld := range wildcardTLDs {
			t.Errorf("  .%s", tld)
		}
	} else {
		t.Logf("✓ No wildcard TLDs detected")
	}
}

func BenchmarkLoadTLDs(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "bench_tlds_*.txt")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	var tlds strings.Builder
	for i := 0; i < 1000; i++ {
		tlds.WriteString("TLD")
		tlds.WriteString(string(rune('A' + i%26)))
		tlds.WriteString("\n")
	}
	os.WriteFile(tmpFile.Name(), []byte(tlds.String()), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loadTLDs(tmpFile.Name())
	}
}
