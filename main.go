package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const workers = 150

func main() {
	var updateMode bool

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <domain> <tld-file> <resolver>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "   or: %s --update\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Scan mode:\n")
		fmt.Fprintf(os.Stderr, "  domain      Domain name to scan (e.g., 'google')\n")
		fmt.Fprintf(os.Stderr, "  tld-file    TLD list file (e.g., 'tlds', 'cctlds')\n")
		fmt.Fprintf(os.Stderr, "  resolver    DNS resolver (e.g., '8.8.8.8:53', '1.1.1.1:53')\n\n")
		fmt.Fprintf(os.Stderr, "Update mode:\n")
		fmt.Fprintf(os.Stderr, "  --update    Fetch IANA TLD list and remove wildcard TLDs\n")
		fmt.Fprintf(os.Stderr, "              Outputs to stdout\n\n")
		fmt.Fprintf(os.Stderr, "Output:\n")
		fmt.Fprintf(os.Stderr, "  One domain per line to stdout\n")
		fmt.Fprintf(os.Stderr, "  Use > to redirect output\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s google tlds 8.8.8.8:53 > results\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --update > tlds\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Common resolvers:\n")
		fmt.Fprintf(os.Stderr, "  Google:     8.8.8.8:53 / 8.8.4.4:53\n")
		fmt.Fprintf(os.Stderr, "  Cloudflare: 1.1.1.1:53 / 1.0.0.1:53\n")
		fmt.Fprintf(os.Stderr, "  Quad9:      9.9.9.9:53\n\n")
	}

	flag.BoolVar(&updateMode, "update", false, "Update TLD list from IANA (removes wildcards)")
	flag.Parse()
	args := flag.Args()

	// Update mode
	if updateMode {
		if err := updateTLDs(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Scan mode
	if len(args) != 3 {
		flag.Usage()
		os.Exit(1)
	}

	domain := args[0]
	tldFile := args[1]
	resolver := args[2]

	tlds := loadTLDs(tldFile)
	scan(domain, tlds, resolver)
}

func loadTLDs(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var tlds []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		tlds = append(tlds, strings.ToLower(line))
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading TLD file: %v\n", err)
		os.Exit(1)
	}

	return tlds
}

func scan(domain string, tlds []string, resolver string) {
	jobs := make(chan string, len(tlds))
	results := make(chan string, workers*2)
	var wg sync.WaitGroup

	// Create custom resolver
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second * 5,
			}
			return d.DialContext(ctx, network, resolver)
		},
	}

	// Start result printer
	var printerWg sync.WaitGroup
	printerWg.Add(1)
	go func() {
		defer printerWg.Done()
		for foundDomain := range results {
			fmt.Println(foundDomain)
		}
	}()

	// Start worker pool
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go worker(domain, jobs, results, &wg, r)
	}

	// Send jobs
	for _, tld := range tlds {
		jobs <- tld
	}
	close(jobs)

	wg.Wait()
	close(results)
	printerWg.Wait()
}

func worker(domain string, jobs <-chan string, results chan<- string, wg *sync.WaitGroup, resolver *net.Resolver) {
	defer wg.Done()

	ctx := context.Background()

	for tld := range jobs {
		fullDomain := fmt.Sprintf("%s.%s", domain, tld)

		_, err := resolver.LookupHost(ctx, fullDomain)
		if err == nil {
			// Verify not a wildcard by testing a random nonsense domain
			randomDomain := fmt.Sprintf("xyzabc-nonexistent-test-99999.%s", tld)
			_, wildcardErr := resolver.LookupHost(ctx, randomDomain)

			// Only output if random domain does NOT resolve (not a wildcard)
			if wildcardErr != nil {
				results <- fullDomain
			}
		}
	}
}
