package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database DatabaseConfig `yaml:"db"`
}

type DatabaseConfig struct {
	Hostname     string `yaml:"host"`
	DatabaseName string `yaml:"dbname"`
	Username     string `yaml:"user"`
	Password     string `yaml:"passwd"`
}

func LoadConfig(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
