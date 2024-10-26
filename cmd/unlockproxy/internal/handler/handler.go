package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/cetteup/unlockproxy/cmd/unlockproxy/internal/asp"
	"github.com/cetteup/unlockproxy/internal/domain/player"
	"github.com/cetteup/unlockproxy/internal/domain/provider"
	"github.com/cetteup/unlockproxy/internal/trace"
)

type repository interface {
	FindByPID(ctx context.Context, pid int) (player.Player, error)
}

type params struct {
	PID int `query:"pid"`
}

type UpstreamResponse struct {
	StatusCode int
	Header     map[string][]string
	Body       []byte
}

type Handler struct {
	repository repository
	provider   provider.Provider

	client *http.Client
}

func NewHandler(repository repository, provider provider.Provider) *Handler {
	return &Handler{
		repository: repository,
		provider:   provider,
		client: &http.Client{
			// Don't follow redirects, just return first response (mimic proxy behaviour)
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (h *Handler) HandleGetUnlocksInfo(c echo.Context) error {
	p := params{}
	if err2 := c.Bind(&p); err2 != nil {
		return c.String(http.StatusOK, asp.NewSyntaxErrorResponse().Serialize())
	}

	unlocks := asp.NewOKResponse().
		WriteHeader("pid", "nick", "asof").
		WriteData(strconv.Itoa(p.PID), "unlockproxy", asp.Timestamp()).
		WriteHeader("enlisted", "officer").
		WriteData("0", "0").
		WriteHeader("id", "state").
		WriteData("11", "s").
		WriteData("22", "s").
		WriteData("33", "s").
		WriteData("44", "s").
		WriteData("55", "s").
		WriteData("66", "s").
		WriteData("77", "s").
		WriteData("88", "s").
		WriteData("99", "s").
		WriteData("111", "s").
		WriteData("222", "s").
		WriteData("333", "s").
		WriteData("444", "s").
		WriteData("555", "s")

	return c.String(http.StatusOK, unlocks.Serialize())
}

func (h *Handler) HandleGetForward(c echo.Context) error {
	p := params{}
	if err2 := c.Bind(&p); err2 != nil {
		return c.String(http.StatusOK, asp.NewSyntaxErrorResponse().Serialize())
	}

	pv, err2 := h.determineProvider(c.Request().Context(), p.PID)
	if err2 != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err2)
	}

	log.Debug().
		Stringer(trace.LogProvider, pv).
		Str("URI", c.Request().RequestURI).
		Msg("Forwarding request")

	res, err2 := h.forwardRequest(c.Request().Context(), pv, c.Request())
	if err2 != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err2)
	}

	// Copy all upstream header to ensure response can be handled correctly downstream
	for key, values := range res.Header {
		for _, value := range values {
			c.Response().Header().Add(key, value)
		}
	}

	return c.String(res.StatusCode, string(res.Body))
}

func (h *Handler) determineProvider(ctx context.Context, pid int) (provider.Provider, error) {
	p, err := h.repository.FindByPID(ctx, pid)
	if err != nil {
		if errors.Is(err, player.ErrPlayerNotFound) {
			log.Warn().
				Int(trace.LogPlayerPID, pid).
				Msg("Player not found, falling back to default provider")
			return h.provider, nil
		}
		if errors.Is(err, player.ErrMultiplePlayersFound) {
			log.Warn().
				Int(trace.LogPlayerPID, pid).
				Msg("Found multiple players, falling back to default provider")
			return h.provider, nil
		}
		return 0, err
	}

	return p.Provider, nil
}

func (h *Handler) forwardRequest(ctx context.Context, pv provider.Provider, incoming *http.Request) (*UpstreamResponse, error) {
	u, err := url.Parse(pv.BaseURL())
	if err != nil {
		return nil, err
	}
	u = u.JoinPath(incoming.URL.Path)
	u.RawQuery = incoming.URL.RawQuery

	req, err := http.NewRequestWithContext(ctx, incoming.Method, u.String(), incoming.Body)
	if err != nil {
		return nil, err
	}

	// Host differs from URL
	req.Host = incoming.Host
	req.Header = incoming.Header.Clone()

	res, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = res.Body.Close()
	if err != nil {
		return nil, err
	}

	return &UpstreamResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       bytes,
	}, nil
}
