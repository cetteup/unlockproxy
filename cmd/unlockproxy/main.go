package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/cetteup/unlockproxy/cmd/unlockproxy/config"
	"github.com/cetteup/unlockproxy/cmd/unlockproxy/internal"
	"github.com/cetteup/unlockproxy/cmd/unlockproxy/options"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
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

	playerRepository := sql.NewRepository(db)

	w := internal.NewWorker(playerRepository)
	todo := make(chan internal.VerifyPlayerParams, 10)
	go w.Run(ctx, todo)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRemoteIP:  true,
		LogURI:       true,
		LogStatus:    true,
		LogLatency:   true,
		LogUserAgent: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info().
				Str("remote", v.RemoteIP).
				Str("URI", v.URI).
				Int("status", v.Status).
				Str("latency", v.Latency.Truncate(time.Millisecond).String()).
				Str("agent", v.UserAgent).
				Msg("request")

			return nil
		},
	}))

	e.GET("/ASP/VerifyPlayer.aspx", func(c echo.Context) error {
		var params internal.VerifyPlayerParams
		if err2 := c.Bind(&params); err2 != nil {
			msg := strings.Join([]string{"E\t216", "$\t4\t$"}, "\n")
			return c.String(http.StatusOK, msg)
		}
		if params.PID >= 10000000 && params.PID < 20000000 {
			todo <- params
		}

		// Mimic BF2Hub response (which does not support player verification)
		msg := strings.Join([]string{"E\t999", "$\t4\t$"}, "\n")
		return c.String(http.StatusOK, msg)
	})

	e.GET(fmt.Sprintf("/ASP/%s", opts.UnlocksEndpoint), func(c echo.Context) error {
		pid := c.QueryParam("pid")
		if _, err2 := strconv.Atoi(pid); err2 != nil {
			msg := strings.Join([]string{"E\t216", "$\t4\t$"}, "\n")
			return c.String(http.StatusOK, msg)
		}

		unlocks := strings.Join([]string{
			"O",
			"H\tpid\tnick\tasof",
			fmt.Sprintf("D\t%s\tunlockproxy\t%d", pid, time.Now().Unix()),
			"H\tenlisted\tofficer",
			"D\t0\t0",
			"H\tid\tstate",
			"D\t11\ts",
			"D\t22\ts",
			"D\t33\ts",
			"D\t44\ts",
			"D\t55\ts",
			"D\t66\ts",
			"D\t77\ts",
			"D\t88\ts",
			"D\t99\ts",
			"D\t111\ts",
			"D\t222\ts",
			"D\t333\ts",
			"D\t444\ts",
			"D\t555\ts",
			"$\t132\t$",
		}, "\n")
		return c.String(http.StatusOK, unlocks)
	})

	originURL, err := url.Parse(opts.OriginBaseURL)
	if err != nil {
		e.Logger.Fatal(err)
	}

	g := e.Group("/ASP/*")
	g.Use(middleware.Proxy(middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
		{
			URL: originURL,
		},
	})))

	go func() {
		if err2 := e.Start(opts.ListenAddr); !errors.Is(err2, http.ErrServerClosed) {
			e.Logger.Fatal(err2)
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	close(todo)
}
