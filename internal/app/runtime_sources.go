package app

import (
	"context"
	"strings"
	"time"

	"campus-room-status/internal/domain"
)

type staticInventorySource struct{}

func (staticInventorySource) LoadInventory(context.Context) (domain.InventorySnapshot, error) {
	return domain.InventorySnapshot{}, nil
}

type staticCalendarClient struct{}

func (staticCalendarClient) ListRoomEvents(_ context.Context, resourceEmail string, _, _ time.Time) ([]domain.Event, error) {
	_ = resourceEmail
	return nil, nil
}

type staticTokenProvider struct {
	token string
}

func (p staticTokenProvider) Token(context.Context) (string, error) {
	return p.token, nil
}

type oauthBootstrapInventorySource struct {
	primary domain.InventorySource
}

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

func (s unavailableInventorySource) LoadInventory(context.Context) (domain.InventorySnapshot, error) {
	return domain.InventorySnapshot{}, s.err
}

type unavailableCalendarClient struct {
	err error
}

func (c unavailableCalendarClient) ListRoomEvents(context.Context, string, time.Time, time.Time) ([]domain.Event, error) {
	return nil, c.err
}
