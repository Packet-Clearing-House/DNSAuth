package main

import (
	"github.com/shafreeck/configo"
)

type Config struct {
	CleanupAction   string `cfg:"cleanup-action; none; /(none|move|delete)/; Action to take after processing a file"`
	CleanupDir      string `cfg:"cleanup-dir; required; path; Path to move processed files when cleanup-action=move"`
	CustomerDB      string `cfg:"customer-db; required; "`
	CustomerRefresh int    `cfg:"customer-refresh; 24; "`
	InfluxDB        string `cfg:"influx-db; required; "`
	WatchDir        string `cfg:"watch-dir; required; "`
}

func LoadConfig(path string) (*Config, error) {

	conf := &Config{}
	if err := configo.Load(path, &conf); err != nil {
		return nil, err
	}
	return conf, nil
}
