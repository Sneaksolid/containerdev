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
	Name         string   `yaml:"name" validate:"required"`
	Image        string   `yaml:"image" validate:"required"`
	Stdin        bool     `yaml:"stdin"`
	AsUser       bool     `yaml:"as_user"`
	MountWorkdir bool     `yaml:"mount_workdir"`
	Mounts       []string `yaml:"mounts"`
	Cmd          []string `yaml:"cmd"`
}

func (c *ContainerConfig) getRunOptions() (*RunOptions, error) {
	opts := RunOptions{
		Name:   c.Name,
		Image:  c.Image,
		Stdin:  c.Stdin,
		AsUser: c.AsUser,

		Volumes: make(map[string]string),
	}

	if c.MountWorkdir {
		pwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("error getting current directory: %w", err)
		}

		opts.Volumes[pwd] = pwd
		opts.WorkDir = pwd
	}

	for _, mount := range c.Mounts {
		opts.Volumes[mount] = mount
	}

	if len(c.Cmd) > 0 {
		opts.Cmd = c.Cmd
	}

	return &opts, nil
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

		runOpts, err := containerCfg.getRunOptions()
		if err != nil {
			return fmt.Errorf("error getting run options: %w", err)
		}

		err = Run(ctx, *runOpts)
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
