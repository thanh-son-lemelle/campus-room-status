package app

import (
	"context"
	"strings"
	"time"

	"campus-room-status/internal/domain"
)

type staticInventorySource struct{}

// LoadInventory loads inventory.
//
// Summary:
// - Loads inventory.
//
// Attributes:
// - arg1 (context.Context): Input parameter.
//
// Returns:
// - value1 (domain.InventorySnapshot): Returned value.
// - value2 (error): Returned value.
func (staticInventorySource) LoadInventory(context.Context) (domain.InventorySnapshot, error) {
	return domain.InventorySnapshot{}, nil
}

type staticCalendarClient struct{}

// ListRoomEvents lists room events.
//
// Summary:
// - Lists room events.
//
// Attributes:
// - arg1 (context.Context): Input parameter.
// - resourceEmail (string): Input parameter.
// - arg2 (time.Time): Input parameter.
// - arg3 (time.Time): Input parameter.
//
// Returns:
// - value1 ([]domain.Event): Returned value.
// - value2 (error): Returned value.
func (staticCalendarClient) ListRoomEvents(_ context.Context, resourceEmail string, _, _ time.Time) ([]domain.Event, error) {
	_ = resourceEmail
	return nil, nil
}

type staticTokenProvider struct {
	token string
}

// Token tokens function behavior.
//
// Summary:
// - Tokens function behavior.
//
// Attributes:
// - arg1 (context.Context): Input parameter.
//
// Returns:
// - value1 (string): Returned value.
// - value2 (error): Returned value.
func (p staticTokenProvider) Token(context.Context) (string, error) {
	return p.token, nil
}

type oauthBootstrapInventorySource struct {
	primary domain.InventorySource
}

// LoadInventory loads inventory.
//
// Summary:
// - Loads inventory.
//
// Attributes:
// - ctx (context.Context): Input parameter.
//
// Returns:
// - value1 (domain.InventorySnapshot): Returned value.
// - value2 (error): Returned value.
func (s oauthBootstrapInventorySource) LoadInventory(ctx context.Context) (domain.InventorySnapshot, error) {
	snapshot, err := s.primary.LoadInventory(ctx)
	if err == nil {
		return snapshot, nil
	}
	if !isMissingRefreshTokenError(err) {
		return domain.InventorySnapshot{}, err
	}

	// Keep API bootstrappable for OAuth consent flow without serving static fixtures.
	return domain.InventorySnapshot{}, nil
}

// isMissingRefreshTokenError is missing refresh token error.
//
// Summary:
// - Is missing refresh token error.
//
// Attributes:
// - err (error): Input parameter.
//
// Returns:
// - value1 (bool): Returned value.
func isMissingRefreshTokenError(err error) bool {
	if err == nil {
		return false
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "missing refresh token")
}

type unavailableInventorySource struct {
	err error
}

// LoadInventory loads inventory.
//
// Summary:
// - Loads inventory.
//
// Attributes:
// - arg1 (context.Context): Input parameter.
//
// Returns:
// - value1 (domain.InventorySnapshot): Returned value.
// - value2 (error): Returned value.
func (s unavailableInventorySource) LoadInventory(context.Context) (domain.InventorySnapshot, error) {
	return domain.InventorySnapshot{}, s.err
}

type unavailableCalendarClient struct {
	err error
}

// ListRoomEvents lists room events.
//
// Summary:
// - Lists room events.
//
// Attributes:
// - arg1 (context.Context): Input parameter.
// - arg2 (string): Input parameter.
// - arg3 (time.Time): Input parameter.
// - arg4 (time.Time): Input parameter.
//
// Returns:
// - value1 ([]domain.Event): Returned value.
// - value2 (error): Returned value.
func (c unavailableCalendarClient) ListRoomEvents(context.Context, string, time.Time, time.Time) ([]domain.Event, error) {
	return nil, c.err
}
