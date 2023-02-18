package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/cetteup/unlockproxy/cmd/unlockproxy/config"
)

var (
	buildVersion = "development"
	buildCommit  = "uncommitted"
	buildTime    = "unknown"
)

func main() {
	version := fmt.Sprintf("unlockproxy %s (%s) built at %s", buildVersion, buildCommit, buildTime)
	cfg := config.Init()

	// Print version and exit
	if cfg.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		NoColor:    !cfg.ColorizeLogs,
		TimeFormat: time.RFC3339,
	})
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	e := echo.New()
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

	e.GET("/ASP/getunlocksinfo.aspx", func(c echo.Context) error {
		pid := c.QueryParam("pid")
		if _, err := strconv.Atoi(pid); err != nil {
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

	originURL, err := url.Parse("http://official.ranking.bf2hub.com/")
	if err != nil {
		e.Logger.Fatal(err)
	}

	g := e.Group("/ASP/*")
	g.Use(middleware.Proxy(middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
		{
			URL: originURL,
		},
	})))

	e.Logger.Fatal(e.Start(cfg.ListenAddr))
}
