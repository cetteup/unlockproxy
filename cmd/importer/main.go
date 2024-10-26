package main

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"slices"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/cetteup/unlockproxy/cmd/importer/internal/config"
	"github.com/cetteup/unlockproxy/cmd/importer/internal/opendata"
	"github.com/cetteup/unlockproxy/cmd/importer/internal/options"
	"github.com/cetteup/unlockproxy/internal/database"
	"github.com/cetteup/unlockproxy/internal/domain/player"
	"github.com/cetteup/unlockproxy/internal/domain/player/sql"
	"github.com/cetteup/unlockproxy/internal/domain/provider"
	"github.com/cetteup/unlockproxy/internal/trace"
)

var (
	buildVersion = "development"
	buildCommit  = "uncommitted"
	buildTime    = "unknown"
)

func main() {
	version := fmt.Sprintf("importer %s (%s) built at %s", buildVersion, buildCommit, buildTime)
	opts := options.Init()

	// Print version and exit
	if opts.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		NoColor:    !opts.ColorizeLogs,
		TimeFormat: time.RFC3339,
	})
	if opts.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	cfg, err := config.LoadConfig(opts.ConfigPath)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("config", opts.ConfigPath).
			Msg("Failed to read config file")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db := database.Connect(
		cfg.Database.Hostname,
		cfg.Database.DatabaseName,
		cfg.Database.Username,
		cfg.Database.Password,
	)
	defer func() {
		err2 := db.Close()
		if err2 != nil {
			log.Error().
				Err(err2).
				Msg("Failed to close database connection")
		}
	}()

	repository := sql.NewRepository(db)

	err = load(ctx, repository, opts.OpendataPath, opts.BatchSize)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to import players from bf2opendata")
	}
}

func load(ctx context.Context, repository player.Repository, basePath string, batchSize int) error {
	var providers = []provider.Provider{
		provider.ProviderBF2Hub,
		provider.ProviderPlayBF2,
		provider.ProviderOpenSpy,
		provider.ProviderB2BF2,
	}
	for _, pv := range providers {
		stats := struct {
			processed int
			imported  int
		}{}
		name := path.Join(basePath, fmt.Sprintf("v_%s.dat", pv.String()))
		batch := make([]player.Player, 0, batchSize)
		err := opendata.LoadPlayersFromFile(ctx, name, func(ctx context.Context, p opendata.Player) error {
			stats.processed++
			batch = append(batch, player.Player{
				PID:      p.PID,
				Nick:     p.Nick,
				Provider: pv,
				Imported: time.Now().UTC(),
			})

			if len(batch) == cap(batch) {
				imported, err2 := insert(ctx, repository, pv, batch)
				if err2 != nil {
					return err2
				}
				stats.imported += imported
				batch = batch[:0]
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to import players from %s: %w", pv, err)
		}

		// Upsert any remaining, incomplete batch
		if len(batch) > 0 {
			imported, err2 := insert(ctx, repository, pv, batch)
			if err2 != nil {
				return err2
			}
			stats.imported += imported
		}

		log.Info().
			Int("processed", stats.processed).
			Msgf("Imported %d players from %s", stats.imported, pv)
	}

	return nil
}

func insert(ctx context.Context, repository player.Repository, pv provider.Provider, players []player.Player) (int, error) {
	if len(players) == 0 {
		return 0, nil
	}

	// Ensure players are sorted ascending by PID
	slices.SortFunc(players, func(a, b player.Player) int {
		return cmp.Compare(a.PID, b.PID)
	})

	existing, err := repository.FindByProviderBetweenPIDs(ctx, pv, players[0].PID, players[len(players)-1].PID)
	if err != nil {
		return 0, fmt.Errorf("failed to find existing players: %w", err)
	}

	// Create map for consistently fast lookups
	catalog := make(map[int]struct{}, len(existing))
	for _, p := range existing {
		catalog[p.PID] = struct{}{}
	}

	nonexistent := make([]player.Player, 0, len(players))
	for _, p := range players {
		if _, exists := catalog[p.PID]; !exists {
			nonexistent = append(nonexistent, p)
		}
	}

	// Insert cannot handle empty slices, so return if there's nothing to insert
	if len(nonexistent) == 0 {
		return 0, nil
	}

	err = repository.InsertMany(ctx, nonexistent)
	if err != nil {
		return 0, fmt.Errorf("failed to insert new players: %w", err)
	}

	for _, p := range nonexistent {
		log.Debug().
			Int(trace.LogPlayerPID, p.PID).
			Str(trace.LogPlayerNick, p.Nick).
			Stringer(trace.LogProvider, pv).
			Msg("Imported player")
	}

	return len(nonexistent), nil
}
