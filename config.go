package main

import (
	"fmt"

	"github.com/itzg/go-flagsfiller"
	"github.com/joho/godotenv"
)

type config struct {
	Debug           bool    `usage:"Enable debug mode"`
	TalentaEmail    string  `usage:"Talenta email"`
	TalentaPassword string  `usage:"Talenta password"`
	Latitude        float64 `usage:"Clock In/Out Latitude"`
	Longitude       float64 `usage:"Clock In/Out Longitude"`
}

func parseConfig() (config, error) {
	godotenv.Load()
	var cfg config
	if err := flagsfiller.Parse(&cfg, flagsfiller.WithEnv("")); err != nil {
		return cfg, fmt.Errorf("parse flag and env: %w", err)
	}
	if cfg.TalentaEmail == "" {
		return cfg, fmt.Errorf("TALENTA_EMAIL is required")
	}
	if cfg.TalentaPassword == "" {
		return cfg, fmt.Errorf("TALENTA_PASSWORD is required")
	}
	if cfg.Latitude == 0 {
		return cfg, fmt.Errorf("LATITUDE is required")
	}
	if cfg.Longitude == 0 {
		return cfg, fmt.Errorf("LONGITUDE is required")
	}
	return cfg, nil
}
