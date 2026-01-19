package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logging  Logging   `yaml:"logging"`
	Defaults Defaults  `yaml:"defaults"`
	Projects []Project `yaml:"projects"`
}

type Logging struct {
	Path string `yaml:"path"`
}

type Defaults struct {
	Provider     string `yaml:"provider"`
	Bucket       string `yaml:"bucket"`
	Cron         string `yaml:"cron"`
	RcloneRemote string `yaml:"rclone_remote"`
}

type Project struct {
	Name          string         `yaml:"name"`
	SourcePaths   []string       `yaml:"source_paths"`
	Cron          string         `yaml:"cron"`
	RcloneRemotes []RcloneRemote `yaml:"rclone_remotes"`
}

type RcloneRemote struct {
	Name   string `yaml:"name"`
	Bucket string `yaml:"bucket"`
}

func Load() (*Config, error) {
	path := envOr("CONFIG_PATH", "config1.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	err = cfg.check()
	if err != nil {
		return nil, fmt.Errorf("check config failed %s: %w", path, err)
	}
	cfg.applyDefaults()
	return cfg, nil
}

func (c *Config) applyDefaults() {
	for i := range c.Projects {
		p := &c.Projects[i]

		if p.Cron == "" {
			p.Cron = c.Defaults.Cron
		}
		if len(p.RcloneRemotes) == 0 && c.Defaults.RcloneRemote != "" {
			p.RcloneRemotes = []RcloneRemote{
				{
					Name:   c.Defaults.RcloneRemote,
					Bucket: c.Defaults.Bucket,
				},
			}
		}
	}
}

func (c *Config) check() error {
	if len(c.Projects) == 0 {
		return nil
	}
	for _, p := range c.Projects {
		if p.Name == "" {
			return fmt.Errorf("project name is required")
		}
		if len(p.SourcePaths) == 0 {
			return fmt.Errorf("project %s: source_paths is required", p.Name)
		}
		if len(p.RcloneRemotes) != 0 {
			for _, r := range p.RcloneRemotes {
				if r.Name == "" {
					return fmt.Errorf("project %s: rclone_remote name is required", p.Name)
				}
				if r.Bucket == "" {
					return fmt.Errorf("project %s: rclone_remote bucket is required", p.Name)
				}
			}
		}
	}
	return nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
