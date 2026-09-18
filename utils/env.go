package utils

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnv reads a .env file and sets environment variables if not already set.
func LoadEnv(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			// Only set if not already set in process environment
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
	}
	return scanner.Err()
}
