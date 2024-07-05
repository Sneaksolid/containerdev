package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v3"
)

var (
	configFile = xdg.ConfigHome + "/containerdev.yaml"
)

type ContainerConfig struct {
	Name  string `yaml:"name" validate:"required"`
	Image string `yaml:"image" validate:"required"`
}

type Config struct {
	Containers []ContainerConfig `yaml:"containers" validate:"required"`
}

func (c *Config) GetContainerConfig(name string) *ContainerConfig {
	for _, containerConfig := range c.Containers {
		if containerConfig.Name == name {
			return &containerConfig
		}
	}

	return nil
}

func main() {
	run(func(ctx context.Context) error {
		if len(os.Args) < 2 {
			return fmt.Errorf("missing container name")
		}

		cfg, err := readConfig()
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}

		containerCfg := cfg.GetContainerConfig(os.Args[1])
		if containerCfg == nil {
			return fmt.Errorf("container not found in config")
		}

		err = Run(ctx, RunOptions{
			Image:      containerCfg.Image,
			EntryPoint: "sh",
			Cmd:        []string{""},
		})
		if err != nil {
			return fmt.Errorf("error running container: %w", err)
		}

		return nil
	})
}

func readConfig() (*Config, error) {
	f, err := os.Open(configFile)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			f, err = writeEmptyConfig(configFile)
			if err != nil {
				return nil, fmt.Errorf("error writing empty config: %w", err)
			}
		default:
			return nil, fmt.Errorf("error opening config file: %w", err)
		}
	}
	defer f.Close()

	cfg := Config{}
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, fmt.Errorf("error decoding config: %w", err)
	}

	return &cfg, nil
}

func writeEmptyConfig(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating config dir: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("error creating config file: %w", err)
	}

	empty := Config{}
	err = yaml.NewEncoder(f).Encode(&empty)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("error writing empty config: %w", err)
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("error seeking to beginning of config file: %w", err)
	}
	return f, nil
}
