package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const defaultExcludeListURL = "https://raw.githubusercontent.com/rix4uni/scope/refs/heads/main/outofscope.txt"

func main() {
	// Parse the exclude list file flag, with the default URL as fallback
	excludeListFile := flag.String("exclude-list", defaultExcludeListURL, "Path to exclude list file or URL")
	flag.Parse()

	// Read exclude patterns from file or URL
	excludePatterns, err := readExcludePatterns(*excludeListFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading exclude list: %v\n", err)
		os.Exit(1)
	}

	// Filter input lines against the exclude patterns
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if !isExcluded(line, excludePatterns) {
			fmt.Println(line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}

// readExcludePatterns reads exclusion patterns from a file or URL.
func readExcludePatterns(source string) ([]string, error) {
	var scanner *bufio.Scanner

	// Check if source is a URL
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// Fetch the exclude list from the URL
		resp, err := http.Get(source)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch exclude list from URL: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
		}

		scanner = bufio.NewScanner(resp.Body)
	} else {
		// Read exclude list from a local file
		file, err := os.Open(source)
		if err != nil {
			return nil, fmt.Errorf("failed to open exclude list file: %v", err)
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	// Read patterns from the scanner
	var patterns []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			patterns = append(patterns, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

// isExcluded checks if the URL matches any pattern in the exclude list.
// Patterns with `*` should match only subdomains, not exact domain names.
func isExcluded(url string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "*.") {
			// Match subdomains only (e.g., "*.arubanetworks.com" matches "admin.arubanetworks.com" but not "arubanetworks.com")
			subdomainPattern := pattern[2:]
			if strings.HasSuffix(url, subdomainPattern) && url != subdomainPattern {
				return true
			}
		} else {
			// Exact match
			if url == pattern {
				return true
			}
		}
	}
	return false
}
