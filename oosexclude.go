package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/pflag"
)

// prints the version message
const version = "v0.0.5"

func printVersion() {
	fmt.Printf("Current oosexclude version: %s\n", version)
}

const defaultExcludeListURL = "https://raw.githubusercontent.com/rix4uni/scope/refs/heads/main/data/outofscope.txt"

func main() {
	// Parse the exclude list file flag, with the default URL as fallback
	egrepFile := pflag.String("egrep", defaultExcludeListURL, "Path to exclude list file or URL")
	grepFile := pflag.String("grep", "", "Path to include list file or URL")
	ignoreCase := pflag.Bool("ignore-case", false, "Match patterns case-insensitively")
	stats := pflag.Bool("stats", false, "Print filtering stats to stderr after processing")
	version := pflag.Bool("version", false, "Print the version of the tool and exit.")
	pflag.Parse()

	// Print version and exit if -version flag is provided
	if *version {
		printVersion()
		return
	}

	// Mutually exclusive: -e and -i cannot be used together
	if pflag.CommandLine.Changed("egrep") && pflag.CommandLine.Changed("grep") {
		fmt.Fprintln(os.Stderr, "Error: --egrep and --grep cannot be used together")
		os.Exit(1)
	}

	var excludeRegexp *regexp.Regexp
	var includeRegexp *regexp.Regexp

	if pflag.CommandLine.Changed("grep") {
		// Include mode: only load include patterns, skip exclude entirely
		raw, err := readPatterns(*grepFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading include list: %v\n", err)
			os.Exit(1)
		}
		includeRegexp = compilePatterns(raw, *ignoreCase)
	} else {
		// Exclude mode: load exclude patterns (default URL or explicit -e)
		raw, err := readPatterns(*egrepFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading exclude list: %v\n", err)
			os.Exit(1)
		}
		excludeRegexp = compilePatterns(raw, *ignoreCase)
	}

	// Detect if stdout is a terminal for colored output
	colorEnabled := false
	if fi, err := os.Stdout.Stat(); err == nil {
		colorEnabled = (fi.Mode() & os.ModeCharDevice) != 0
	}

	// Filter input lines
	var inputCount, keptCount int
	out := bufio.NewWriter(os.Stdout)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1*1024*1024), 1*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		inputCount++
		if includeRegexp != nil {
			if loc := includeRegexp.FindStringIndex(line); loc != nil {
				fmt.Fprintln(out, highlightMatch(line, loc, colorEnabled))
				keptCount++
			}
		} else {
			if excludeRegexp == nil || !excludeRegexp.MatchString(line) {
				fmt.Fprintln(out, line)
				keptCount++
			}
		}
	}
	out.Flush()

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	if *stats {
		fmt.Fprintf(os.Stderr, "[stats] input: %d  kept: %d  removed: %d\n", inputCount, keptCount, inputCount-keptCount)
	}
}

// readPatterns reads patterns from a file or URL.
func readPatterns(source string) ([]string, error) {
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
	} else if _, err := os.Stat(source); err == nil {
		// Read patterns from a local file
		file, err := os.Open(source)
		if err != nil {
			return nil, fmt.Errorf("failed to open exclude list file: %v", err)
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	} else {
		// Inline: treat as comma-separated pattern string
		var patterns []string
		for _, p := range strings.Split(source, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				patterns = append(patterns, p)
			}
		}
		return patterns, nil
	}

	// Read patterns from the scanner
	var patterns []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

// globToRegex converts a glob-style pattern to a regex string.
// * -> .*, ? -> ., other regex special chars are escaped, [...] preserved as-is.
func globToRegex(pattern string) string {
	var sb strings.Builder
	inBracket := false
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		switch {
		case c == '[' && !inBracket:
			inBracket = true
			sb.WriteByte(c)
		case c == ']' && inBracket:
			inBracket = false
			sb.WriteByte(c)
		case inBracket:
			sb.WriteByte(c)
		case c == '*':
			sb.WriteString(".*")
		case c == '?':
			sb.WriteByte('.')
		case c == '.' || c == '+' || c == '(' || c == ')' || c == '{' || c == '}' || c == '^' || c == '$' || c == '|' || c == '\\':
			sb.WriteByte('\\')
			sb.WriteByte(c)
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

// compilePatterns joins all patterns into a single *regexp.Regexp using alternation.
// This gives O(M) matching regardless of pattern count N, same as grep internals.
// ignoreCase=true prepends (?i) to the combined regex.
func compilePatterns(patterns []string, ignoreCase bool) *regexp.Regexp {
	if len(patterns) == 0 {
		return nil
	}
	parts := make([]string, 0, len(patterns))
	for _, p := range patterns {
		converted := globToRegex(p)
		if _, err := regexp.Compile(converted); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: invalid pattern %q: %v\n", p, err)
			continue
		}
		parts = append(parts, "(?:"+converted+")")
	}
	if len(parts) == 0 {
		return nil
	}
	combined := strings.Join(parts, "|")
	if ignoreCase {
		combined = "(?i)" + combined
	}
	re, err := regexp.Compile(combined)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not compile combined pattern: %v\n", err)
		return nil
	}
	return re
}

// highlightMatch wraps the pre-found match location in bold-red ANSI codes.
func highlightMatch(line string, loc []int, color bool) string {
	if !color {
		return line
	}
	return line[:loc[0]] + "\033[01;31m" + line[loc[0]:loc[1]] + "\033[0m" + line[loc[1]:]
}
