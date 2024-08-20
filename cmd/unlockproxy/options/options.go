package options

import (
	"flag"
)

type Options struct {
	Version bool

	ListenAddr   string
	Debug        bool
	ColorizeLogs bool

	ConfigPath string

	OriginBaseURL   string
	UnlocksEndpoint string
}

func Init() *Options {
	opts := new(Options)
	flag.BoolVar(&opts.Version, "v", false, "prints the version")
	flag.BoolVar(&opts.Version, "version", false, "prints the version")
	flag.BoolVar(&opts.Debug, "debug", false, "enable debug logging")
	flag.BoolVar(&opts.ColorizeLogs, "colorize-logs", false, "colorize log messages")
	flag.StringVar(&opts.ListenAddr, "address", ":8080", "server/bind address in format [host]:port")
	flag.StringVar(&opts.ConfigPath, "config", "config.yaml", "path to YAML config file")
	flag.StringVar(&opts.OriginBaseURL, "origin", "http://official.ranking.bf2hub.com", "origin to use for all other aspx endpoints, base URL without \"/ASP/\"")
	flag.StringVar(&opts.UnlocksEndpoint, "unlocks-endpoint", "getunlocksinfo.aspx", "path to use for fake getunlocksinfo endpoint, without leading \"/ASP/\"")
	flag.Parse()
	return opts
}
