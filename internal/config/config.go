package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	ACRCloud struct {
		AccessKey string `yaml:"access_key"`
		SecretKey string `yaml:"secret_key"`
		Host      string `yaml:"host"`
	} `yaml:"acrcloud"`

	Overlay struct {
		FontSize   int     `yaml:"font_size"`
		FontFamily string  `yaml:"font_family"`
		Position   string  `yaml:"position"` // "bottom", "top", "center"
		Opacity    float64 `yaml:"opacity"`
	} `yaml:"overlay"`

	Detection struct {
		Interval int `yaml:"interval"` // seconds between detections
	} `yaml:"detection"`
}

// Load reads configuration from a YAML file
func Load(path string) (*Config, error) {
	// TODO: Step 2 - Implement configuration loading
	// Read the file, unmarshal YAML into Config struct

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
