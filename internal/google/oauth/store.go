package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RefreshTokenStore interface {
	Save(ctx context.Context, token string) error
	Load(ctx context.Context) (string, error)
}

type FileRefreshTokenStore struct {
	path string
}

// NewFileRefreshTokenStore creates a new file refresh token store.
//
// Summary:
// - Creates a new file refresh token store.
//
// Attributes:
// - path (string): Input parameter.
//
// Returns:
// - value1 (*FileRefreshTokenStore): Returned value.
func NewFileRefreshTokenStore(path string) *FileRefreshTokenStore {
	// TODO(prod): avoid local plaintext file; use a managed secret backend.
	return &FileRefreshTokenStore{path: strings.TrimSpace(path)}
}

type tokenFilePayload struct {
	RefreshToken string    `json:"refresh_token"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Save saves function behavior.
//
// Summary:
// - Saves function behavior.
//
// Attributes:
// - arg1 (context.Context): Input parameter.
// - token (string): Input parameter.
//
// Returns:
// - value1 (error): Returned value.
func (s *FileRefreshTokenStore) Save(_ context.Context, token string) error {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return errors.New("refresh token file path is required")
	}
	if strings.TrimSpace(token) == "" {
		return errors.New("refresh token is required")
	}

	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}

	payload := tokenFilePayload{
		RefreshToken: strings.TrimSpace(token),
		UpdatedAt:    time.Now().UTC(),
	}
	// TODO(prod): encrypt token payload at rest (KMS/envelope) if file storage is unavoidable.
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, raw, 0o600); err != nil {
		return err
	}

	return os.Rename(tmpPath, s.path)
}

// Load loads data.
//
// Summary:
// - Loads data.
//
// Attributes:
// - arg1 (context.Context): Input parameter.
//
// Returns:
// - value1 (string): Returned value.
// - value2 (error): Returned value.
func (s *FileRefreshTokenStore) Load(_ context.Context) (string, error) {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return "", errors.New("refresh token file path is required")
	}

	raw, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}

	var payload tokenFilePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", err
	}

	return strings.TrimSpace(payload.RefreshToken), nil
}
