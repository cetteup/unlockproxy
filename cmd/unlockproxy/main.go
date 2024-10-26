package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/cetteup/unlockproxy/cmd/unlockproxy/internal/config"
	"github.com/cetteup/unlockproxy/cmd/unlockproxy/internal/handler"
	"github.com/cetteup/unlockproxy/cmd/unlockproxy/internal/options"
	"github.com/cetteup/unlockproxy/internal/database"
	"github.com/cetteup/unlockproxy/internal/domain/player/sql"
)

var (
	buildVersion = "development"
	buildCommit  = "uncommitted"
	buildTime    = "unknown"
)

func main() {
	version := fmt.Sprintf("unlockproxy %s (%s) built at %s", buildVersion, buildCommit, buildTime)
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
	h := handler.NewHandler(repository, opts.Provider)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Second * 10,
	}))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogError:     true,
		LogRemoteIP:  true,
		LogMethod:    true,
		LogURI:       true,
		LogStatus:    true,
		LogLatency:   true,
		LogUserAgent: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info().
				Err(v.Error).
				Str("remote", v.RemoteIP).
				Str("method", v.Method).
				Str("URI", v.URI).
				Int("status", v.Status).
				Str("latency", v.Latency.Truncate(time.Millisecond).String()).
				Str("agent", v.UserAgent).
				Msg("request")

			return nil
		},
	}))

	// Requests handled locally
	e.GET("/ASP/getunlocksinfo.aspx", h.HandleGetUnlocksInfo)

	// Requests forwarded based on player provider
	e.GET("/ASP/getplayerinfo.aspx", h.HandleGetForward)
	e.GET("/ASP/getawardsinfo.aspx", h.HandleGetForward)
	e.GET("/ASP/getrankinfo.aspx", h.HandleGetForward)

	originURL, err := url.Parse(opts.Provider.BaseURL())
	if err != nil {
		e.Logger.Fatal(err)
	}

	g := e.Group("/ASP/*")
	g.Use(middleware.Proxy(middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
		{
			URL: originURL,
		},
	})))

	e.Logger.Fatal(e.Start(opts.ListenAddr))
}
