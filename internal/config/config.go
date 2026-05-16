package config

import (
	"bytes"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration.
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
		Interval int `yaml:"interval"` // seconds between detections (Linux/macOS ACRCloud)
	} `yaml:"detection"`

	Lyrics struct {
		LRCLibURL        string `yaml:"lrclib_url"`
		NeteaseSearchURL string `yaml:"netease_search_url"`
		NeteaseLyricsURL string `yaml:"netease_lyrics_url"`
	} `yaml:"lyrics"`
}

// Parse reads configuration from raw YAML bytes.
func Parse(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Load reads configuration from a YAML file path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// LoadOrDefault tries the external config file first; falls back to embedded
// default YAML if the file is missing or invalid.
func LoadOrDefault(path string, defaultYAML []byte) *Config {
	if data, err := os.ReadFile(path); err == nil {
		if cfg, err := Parse(data); err == nil {
			return cfg
		}
		log.Printf("config: %s is invalid, using embedded defaults", path)
	} else {
		log.Printf("config: %s not found, using embedded defaults", path)
	}
	cfg, err := Parse(defaultYAML)
	if err != nil {
		panic("config: failed to parse embedded defaults: " + err.Error())
	}
	return cfg
}
