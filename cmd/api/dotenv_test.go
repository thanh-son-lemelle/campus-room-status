package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnvFile_LoadsMissingVariablesWithoutOverridingExistingOnes(t *testing.T) {
	existingKey := "DOTENV_TEST_EXISTING_KEY"
	newKey := "DOTENV_TEST_NEW_KEY"
	quotedKey := "DOTENV_TEST_QUOTED"
	exportedKey := "DOTENV_TEST_EXPORTED"

	t.Setenv(existingKey, "from-process")
	_ = os.Unsetenv(newKey)
	_ = os.Unsetenv(quotedKey)
	_ = os.Unsetenv(exportedKey)

	path := filepath.Join(t.TempDir(), ".env")
	content := "" +
		"# comment\n" +
		existingKey + "=from-file\n" +
		newKey + "=from-file\n" +
		quotedKey + "=\"value with spaces\"\n" +
		"export " + exportedKey + "=from-export\n" +
		"   \n"

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write dotenv file: %v", err)
	}

	if err := loadDotEnvFile(path); err != nil {
		t.Fatalf("load dotenv file: %v", err)
	}

	if got := os.Getenv(existingKey); got != "from-process" {
		t.Fatalf("expected %s to stay from-process, got %q", existingKey, got)
	}
	if got := os.Getenv(newKey); got != "from-file" {
		t.Fatalf("expected %s=from-file, got %q", newKey, got)
	}
	if got := os.Getenv(quotedKey); got != "value with spaces" {
		t.Fatalf("expected %s=value with spaces, got %q", quotedKey, got)
	}
	if got := os.Getenv(exportedKey); got != "from-export" {
		t.Fatalf("expected %s=from-export, got %q", exportedKey, got)
	}
}

func TestLoadDotEnvFile_IgnoresMalformedLines(t *testing.T) {
	okKey := "DOTENV_TEST_OK"
	_ = os.Unsetenv(okKey)

	path := filepath.Join(t.TempDir(), ".env")
	content := "" +
		"INVALID_LINE\n" +
		"=missing_key\n" +
		okKey + "=1\n"

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write dotenv file: %v", err)
	}

	if err := loadDotEnvFile(path); err != nil {
		t.Fatalf("load dotenv file: %v", err)
	}

	if got := os.Getenv(okKey); got != "1" {
		t.Fatalf("expected %s=1, got %q", okKey, got)
	}
}

func TestLoadDotEnvFile_ReturnsNotExistWhenFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.env")

	err := loadDotEnvFile(path)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}
