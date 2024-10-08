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

	"github.com/cetteup/unlockproxy/cmd/unlockproxy/options"
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

	e.GET(fmt.Sprintf("/ASP/%s", opts.UnlocksEndpoint), func(c echo.Context) error {
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

	e.Logger.Fatal(e.Start(opts.ListenAddr))
}
