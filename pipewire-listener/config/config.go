package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type HassConfig struct {
	URL              string `yaml:"url"`
	Token            string `yaml:"token"`
	SpeakersDeviceID string `yaml:"speakersDeviceId"`
	SpeakersEntityID string `yaml:"speakersEntityId"`
}

type Config struct {
	Hass            *HassConfig `yaml:"hass"`
	Volt_AlsaOutput string      `yaml:"voltAlsaOutput"`
}

func MustRead(path string, c **Config) {
	if path == "" {
		path = "config.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().
			Str("path", path).
			Err(err).
			Msg("Failed to read config file")
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		log.Fatal().
			Str("path", path).
			Err(err).
			Msg("Invalid config file")
	}
}
