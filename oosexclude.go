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
const version = "v0.0.4"

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

	var excludeRegexps []*regexp.Regexp
	var includeRegexps []*regexp.Regexp

	if pflag.CommandLine.Changed("grep") {
		// Include mode: only load include patterns, skip exclude entirely
		raw, err := readPatterns(*grepFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading include list: %v\n", err)
			os.Exit(1)
		}
		includeRegexps = compilePatterns(raw, false, *ignoreCase)
	} else {
		// Exclude mode: load exclude patterns (default URL or explicit -e)
		raw, err := readPatterns(*egrepFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading exclude list: %v\n", err)
			os.Exit(1)
		}
		excludeRegexps = compilePatterns(raw, false, *ignoreCase)
	}

	// Detect if stdout is a terminal for colored output
	colorEnabled := false
	if fi, err := os.Stdout.Stat(); err == nil {
		colorEnabled = (fi.Mode() & os.ModeCharDevice) != 0
	}

	// Filter input lines
	var inputCount, keptCount int
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		inputCount++
		if len(includeRegexps) > 0 {
			if re := findIncludeMatch(line, includeRegexps); re != nil {
				fmt.Println(highlightMatch(line, re, colorEnabled))
				keptCount++
			}
		} else {
			if !isExcluded(line, excludeRegexps) {
				fmt.Println(line)
				keptCount++
			}
		}
	}

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
// When anchored is true, wraps the result in ^...$ for full-line matching.
// When ignoreCase is true, prepends (?i) for case-insensitive matching.
func globToRegex(pattern string, anchored bool, ignoreCase bool) string {
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
	result := sb.String()
	if anchored {
		result = "^" + result + "$"
	}
	if ignoreCase {
		return "(?i)" + result
	}
	return result
}

// compilePatterns compiles glob/regex pattern strings into *regexp.Regexp.
// anchored=true adds ^...$ for full-line matching (used by --egrep).
// ignoreCase=true prepends (?i) for case-insensitive matching.
func compilePatterns(patterns []string, anchored bool, ignoreCase bool) []*regexp.Regexp {
	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		re, err := regexp.Compile(globToRegex(pattern, anchored, ignoreCase))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: invalid pattern %q: %v\n", pattern, err)
			continue
		}
		compiled = append(compiled, re)
	}
	return compiled
}

// isExcluded checks if the URL matches any compiled exclude pattern.
func isExcluded(url string, regexps []*regexp.Regexp) bool {
	for _, re := range regexps {
		if re.MatchString(url) {
			return true
		}
	}
	return false
}

// findIncludeMatch returns the first compiled pattern that matches url, or nil.
func findIncludeMatch(url string, regexps []*regexp.Regexp) *regexp.Regexp {
	for _, re := range regexps {
		if re.MatchString(url) {
			return re
		}
	}
	return nil
}

// highlightMatch wraps the first match in url with bold-red ANSI codes.
func highlightMatch(line string, re *regexp.Regexp, color bool) string {
	if !color {
		return line
	}
	loc := re.FindStringIndex(line)
	if loc == nil {
		return line
	}
	return line[:loc[0]] + "\033[01;31m" + line[loc[0]:loc[1]] + "\033[0m" + line[loc[1]:]
}
