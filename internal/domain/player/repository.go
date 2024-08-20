package player

import (
	"context"
	"errors"

	"github.com/cetteup/unlockproxy/internal/domain/provider"
)

var (
	ErrPlayerExists   = errors.New("player already exists")
	ErrPlayerNotFound = errors.New("player not found")
)

type Repository interface {
	Insert(ctx context.Context, player Player) error
	InsertMany(ctx context.Context, players []Player) error
	FindByProviderBetweenPIDs(ctx context.Context, pv provider.Provider, lower, upper int) ([]Player, error)
}
