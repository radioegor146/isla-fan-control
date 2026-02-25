package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type FanConfig struct {
	GPUs         []string        `yaml:"gpus"`
	Curve        map[uint32]byte `yaml:"curve"`
	DefaultSpeed byte            `yaml:"default_speed"`
	Hysteresis   uint32          `yaml:"hysteresis"`
}

type Config struct {
	Fans           map[byte]FanConfig `yaml:"fans"`
	UpdateInterval time.Duration      `yaml:"update_interval"`
}

func ParseConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	for id, fan := range cfg.Fans {
		if fan.DefaultSpeed > 100 {
			return nil, fmt.Errorf("fan %d default_speed must be 0-100", id)
		}
		for _, speed := range fan.Curve {
			if speed > 100 {
				return nil, fmt.Errorf("fan %d curve speed must be 0-100", id)
			}
		}
	}

	return &cfg, nil
}
