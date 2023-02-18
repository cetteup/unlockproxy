package config

import (
	"flag"
)

type Config struct {
	Version bool

	ListenAddr   string
	Debug        bool
	ColorizeLogs bool
}

func Init() *Config {
	cfg := new(Config)
	flag.BoolVar(&cfg.Version, "v", false, "prints the version")
	flag.BoolVar(&cfg.Version, "version", false, "prints the version")
	flag.BoolVar(&cfg.Debug, "debug", false, "enable debug logging")
	flag.BoolVar(&cfg.ColorizeLogs, "colorize-logs", false, "colorize log messages")
	flag.StringVar(&cfg.ListenAddr, "address", ":8080", "server/bind address in format [host]:port")
	flag.Parse()
	return cfg
}
