package sql

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"

	"github.com/cetteup/unlockproxy/internal/domain/player"
)

const (
	playerTable = "players"

	columnID                        = "id"
	columnPID                       = "pid"
	columnNick                      = "nick"
	columnCountry                   = "country"
	columnProvider                  = "provider"
	columnAdded                     = "added"
	columnUpdated                   = "updated"
	columnLastSeen                  = "last_seen"
	columnCountryLastChecked        = "country_last_checked"
	columnOtherProvidersLastChecked = "other_providers_last_checked"
	columnErrors                    = "err"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Insert(ctx context.Context, p player.Player) error {
	insert := sq.
		Insert(playerTable).
		Columns(
			columnPID,
			columnNick,
			columnProvider,
			columnAdded,
		).
		Values(
			p.PID,
			p.Nick,
			p.Provider,
			p.Added,
		)

	_, err := insert.RunWith(r.db).ExecContext(ctx)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return player.ErrPlayerExists
		}
		return err
	}

	return nil
}
