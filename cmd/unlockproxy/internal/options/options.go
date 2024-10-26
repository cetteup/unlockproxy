package options

import (
	"flag"

	"github.com/cetteup/unlockproxy/internal/domain/provider"
)

type Options struct {
	Version bool

	ListenAddr   string
	Debug        bool
	ColorizeLogs bool

	ConfigPath string

	Provider provider.Provider
}

func Init() *Options {
	opts := new(Options)
	flag.BoolVar(&opts.Version, "v", false, "prints the version")
	flag.BoolVar(&opts.Version, "version", false, "prints the version")
	flag.BoolVar(&opts.Debug, "debug", false, "enable debug logging")
	flag.BoolVar(&opts.ColorizeLogs, "colorize-logs", false, "colorize log messages")
	flag.StringVar(&opts.ListenAddr, "address", ":8080", "server/bind address in format [host]:port")
	flag.StringVar(&opts.ConfigPath, "config", "config.yaml", "path to YAML config file")
	flag.TextVar(&opts.Provider, "provider", provider.ProviderB2BF2, "primary provider to use as fallback/for directly proxied endpoints (bf2hub|playbf2|openspy|b2bf2)")
	flag.Parse()
	return opts
}
