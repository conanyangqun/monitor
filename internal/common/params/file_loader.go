package params

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// loadLinesFromFile loads lines from a text file
func loadLinesFromFile(filename string) []string {
	// Get the path relative to the current package directory
	filePath := filepath.Join("internal", "common", "params", "data", filename)

	file, err := os.Open(filePath)
	if err != nil {
		// Fallback to empty slice if file not found
		return []string{}
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			fmt.Println("Error closing file:", err)
		}
	}(file)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// LoadUserAgents loads user agent strings from file
func LoadUserAgents() []string {
	agents := loadLinesFromFile("user_agents.txt")
	if len(agents) == 0 {
		// Fallback data if file not found
		return []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Gecko/20100101 Firefox/89.0",
		}
	}
	return agents
}

// LoadFirstNames loads first names from file
func LoadFirstNames() []string {
	names := loadLinesFromFile("first_names.txt")
	if len(names) == 0 {
		// Fallback data if file not found
		return []string{"John", "Jane", "Alex", "Emily", "Michael", "Sarah"}
	}
	return names
}

// LoadLastNames loads last names from file
func LoadLastNames() []string {
	names := loadLinesFromFile("last_names.txt")
	if len(names) == 0 {
		// Fallback data if file not found
		return []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia"}
	}
	return names
}

// LoadFakeDomains loads fake domains from file
func LoadFakeDomains() []string {
	domains := loadLinesFromFile("fake_domains.txt")
	if len(domains) == 0 {
		// Fallback data if file not found
		return []string{"example.com", "test.com", "demo.com", "sample.com"}
	}
	return domains
}
