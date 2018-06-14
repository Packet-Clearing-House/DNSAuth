package main

import (
	"github.com/shafreeck/configo"
	"github.com/Packet-Clearing-House/DNSAuth/DNSAuth/bgp"
)

type Config struct {
	BGP *bgp.BGPConf
	CustomerDB string `cfg:"customer-db; required; "`
	CustomerRefresh int `cfg:"customer-refresh; 24; "`
	InfluxDB string `cfg:"influx-db; required; "`
	WatchDir string `cfg:"watch-dir; required; "`
}

func LoadConfig(path string) (*Config, error) {

	conf := &Config{}
	if err := configo.Load(path, &conf); err != nil {
		return nil, err
	}
	return conf, nil
}
