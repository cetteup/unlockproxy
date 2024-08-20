package internal

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/cetteup/unlockproxy/internal/domain/player"
	"github.com/cetteup/unlockproxy/internal/domain/provider"
	"github.com/cetteup/unlockproxy/internal/trace"
)

type VerifyPlayerParams struct {
	PID  int    `query:"pid"`
	Nick string `query:"SoldierNick"`
	Auth string `query:"auth"`
}

type Worker struct {
	repository player.Repository
}

func NewWorker(repository player.Repository) *Worker {
	return &Worker{repository: repository}

}

func (w *Worker) Run(ctx context.Context, check <-chan VerifyPlayerParams) {
	for params := range check {
		valid, err := w.verifyPlayer(params)
		if err != nil {
			log.Error().
				Err(err).
				Int(trace.LogPlayerPID, params.PID).
				Str(trace.LogPlayerNick, params.Nick).
				Msg("Failed to verify player")
			continue
		}

		log.Debug().
			Int(trace.LogPlayerPID, params.PID).
			Str(trace.LogPlayerNick, params.Nick).
			Bool("valid", valid).
			Msg("Verified player")

		if valid {
			p := player.Player{
				PID:      params.PID,
				Nick:     params.Nick,
				Provider: provider.ProviderOpenSpy,
				Added:    time.Now().UTC(),
			}

			err = w.repository.Insert(ctx, p)
			if err != nil {
				if errors.Is(err, player.ErrPlayerExists) {
					log.Debug().
						Int(trace.LogPlayerPID, p.PID).
						Str(trace.LogPlayerNick, p.Nick).
						Msg("Player already exists")
				} else {
					log.Error().
						Err(err).
						Int(trace.LogPlayerPID, p.PID).
						Str(trace.LogPlayerNick, p.Nick).
						Msg("Failed to insert player")
				}
			} else {
				log.Info().
					Int(trace.LogPlayerPID, p.PID).
					Str(trace.LogPlayerNick, p.Nick).
					Msg("Inserted new player")
			}
		}
	}
}

func (w *Worker) verifyPlayer(params VerifyPlayerParams) (bool, error) {
	u, err := url.Parse("http://bf2web.openspy.net/ASP/VerifyPlayer.aspx")
	if err != nil {
		return false, err
	}

	q := u.Query()
	q.Add("auth", params.Auth)
	q.Add("SoldierNick", params.Nick)
	q.Add("pid", strconv.Itoa(params.PID))
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return false, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if bytes.Contains(body, []byte("H\tresult\t\nD\tOk")) {
		return true, nil
	}

	return false, nil
}
