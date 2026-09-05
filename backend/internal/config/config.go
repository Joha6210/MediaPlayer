package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP           HTTPConfig           `yaml:"http"`
	MPV            MPVConfig            `yaml:"mpv"`
	Plexamp        PlexampConfig        `yaml:"plexamp"`
	SourceDefaults SourceDefaultsConfig `yaml:"sourceDefaults"`
	Runtime        RuntimeConfig        `yaml:"runtime"`
}

type HTTPConfig struct {
	ListenAddr string `yaml:"listenAddr"`
}

type MPVConfig struct {
	SocketPath string `yaml:"socketPath"`
}

type PlexampConfig struct {
	BaseURL string `yaml:"baseURL"`
	Token   string `yaml:"token"`
}

type SourceDefaultsConfig struct {
	DefaultVolume int `yaml:"defaultVolume"`
}

type RuntimeConfig struct {
	TestMode bool `yaml:"testMode"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.HTTP.ListenAddr == "" {
		return errors.New("http.listenAddr is required")
	}
	if c.MPV.SocketPath == "" {
		return errors.New("mpv.socketPath is required")
	}
	if c.SourceDefaults.DefaultVolume < 0 || c.SourceDefaults.DefaultVolume > 100 {
		return errors.New("sourceDefaults.defaultVolume must be between 0 and 100")
	}
	return nil
}
