package player

import (
	"context"
	"errors"
)

var (
	ErrPlayerExists = errors.New("player already exists")
)

type Repository interface {
	Insert(ctx context.Context, player Player) error
}
