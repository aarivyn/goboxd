package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Limits struct {
	WallTimeS    int `yaml:"wall_time_s"`
	MemoryKB     int `yaml:"memory_kb"`
	MaxProcesses int `yaml:"max_processes"`
}

type BuildConfig struct {
	Cmd           string   `yaml:"cmd"`
	Args          []string `yaml:"args"`
	Limits        Limits   `yaml:"limits"`
	FlagAllowlist []string `yaml:"flag_allowlist"`
}

type RunConfig struct {
	Cmd    string   `yaml:"cmd"`
	Args   []string `yaml:"args"`
	Limits Limits   `yaml:"limits"`
}

type Language struct {
	ID                     string      `yaml:"id"`
	Name                   string      `yaml:"name"`
	SourceFilename         string      `yaml:"source_filename"`
	SourceFilenameStrategy string      `yaml:"source_filename_strategy"`
	Artifact               string      `yaml:"artifact"`
	Build                  BuildConfig `yaml:"build"`
	Run                    RunConfig   `yaml:"run"`
}

type Config struct {
	Languages []Language `yaml:"languages"`
}

func Load(path string) (*Config, error) {
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

func (c *Config) FindLanguage(id string) *Language {
	for i := range c.Languages {
		if c.Languages[i].ID == id {
			return &c.Languages[i]
		}
	}
	return nil
}
