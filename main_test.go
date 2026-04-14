package main

import (
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
