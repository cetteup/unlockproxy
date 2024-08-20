package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"

	"github.com/cetteup/unlockproxy/internal/domain/player"
	"github.com/cetteup/unlockproxy/internal/domain/provider"
)

const (
	playerTable = "players"

	columnPID      = "pid"
	columnNick     = "nick"
	columnProvider = "provider"
	columnImported = "imported"
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
	query := sq.
		Insert(playerTable).
		Columns(
			columnPID,
			columnNick,
			columnProvider,
			columnImported,
		).
		Values(
			p.PID,
			p.Nick,
			p.Provider,
			p.Imported,
		)

	_, err := query.RunWith(r.db).ExecContext(ctx)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return player.ErrPlayerExists
		}
		return err
	}

	return nil
}

func (r *Repository) InsertMany(ctx context.Context, players []player.Player) error {
	query := sq.
		Insert(playerTable).
		Columns(
			columnPID,
			columnNick,
			columnProvider,
			columnImported,
		)

	for _, p := range players {
		query = query.Values(
			p.PID,
			p.Nick,
			p.Provider,
			p.Imported,
		)
	}

	_, err := query.RunWith(r.db).ExecContext(ctx)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return player.ErrPlayerExists
		}
		return err
	}

	return nil
}

func (r *Repository) FindByProviderBetweenPIDs(ctx context.Context, pv provider.Provider, lower, upper int) ([]player.Player, error) {
	query := sq.
		Select(
			columnPID,
			columnNick,
			columnProvider,
			columnImported,
		).
		From(playerTable).
		Where(sq.And{
			sq.Eq{columnProvider: pv},
			sq.GtOrEq{columnPID: lower},
			sq.LtOrEq{columnPID: upper},
		}).
		OrderBy(
			fmt.Sprintf("%s ASC", columnProvider),
		)

	rows, err := query.RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	players := make([]player.Player, 0)
	for rows.Next() {
		var p player.Player
		if err = rows.Scan(
			&p.PID,
			&p.Nick,
			&p.Provider,
			&p.Imported,
		); err != nil {
			return nil, err
		}

		players = append(players, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return players, nil
}
