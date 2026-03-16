package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// loadDotEnvIfPresent loads environment variables from a local .env file when present.
// Existing process variables are preserved and never overwritten.
func loadDotEnvIfPresent() {
	candidates := []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join("..", "..", ".env"),
	}

	for _, candidate := range candidates {
		err := loadDotEnvFile(candidate)
		if err == nil {
			return
		}
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		return
	}
}

func loadDotEnvFile(path string) error {
	file, err := os.Open(path)
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

		line = strings.TrimPrefix(line, "export ")

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		value = strings.TrimSpace(value)
		value = unquoteDotEnvValue(value)

		if setErr := os.Setenv(key, value); setErr != nil {
			return fmt.Errorf("set env %s from %s: %w", key, path, setErr)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	return nil
}

func unquoteDotEnvValue(raw string) string {
	if len(raw) < 2 {
		return raw
	}

	if strings.HasPrefix(raw, "\"") && strings.HasSuffix(raw, "\"") {
		return strings.TrimSuffix(strings.TrimPrefix(raw, "\""), "\"")
	}
	if strings.HasPrefix(raw, "'") && strings.HasSuffix(raw, "'") {
		return strings.TrimSuffix(strings.TrimPrefix(raw, "'"), "'")
	}
	return raw
}
