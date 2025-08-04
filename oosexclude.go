package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"path"

	"github.com/spf13/pflag"
)

// prints the version message
const version = "v0.0.2"

func printVersion() {
	fmt.Printf("Current oosexclude version: %s\n", version)
}

const defaultExcludeListURL = "https://raw.githubusercontent.com/rix4uni/scope/refs/heads/main/outofscope.txt"

func main() {
	// Parse the exclude list file flag, with the default URL as fallback
	excludeListFile := pflag.StringP("exclude-list", "e", defaultExcludeListURL, "Path to exclude list file or URL")
	verbose := pflag.Bool("verbose", false, "enable verbose mode")
	version := pflag.BoolP("version", "v", false, "Print the version of the tool and exit.")
	pflag.Parse()

	// Print version and exit if -version flag is provided
	if *version {
		printVersion()
		return
	}

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
		if isExcluded(line, excludePatterns) {
			if *verbose {
				fmt.Printf("IGNORED: %s\n", line)
			}
		} else {
			if *verbose {
				fmt.Printf("NOT IGNORED: %s\n", line)
			} else {
				fmt.Println(line)
			}
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
		match, err := path.Match(pattern, url)
		if err != nil {
			// If the pattern is invalid, skip it
			continue
		}
		if match {
			return true
		}
	}
	return false
}
