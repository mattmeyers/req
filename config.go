package req

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Root         string            `toml:"root"`
	DefaultEnv   string            `toml:"default_env"`
	Aliases      map[string]string `toml:"aliases"`
	Environments map[string]Env    `toml:"environments"`
}

type Env map[string]string

func ParseConfig(path string) (*Config, error) {
	if path == "" {
		path = "./.reqrc"
	}

	var c Config
	_, err := toml.DecodeFile(path, &c)
	if os.IsNotExist(err) {
		return defaultConfig(), nil
	} else if err != nil {
		return nil, err
	}

	for k, v := range c.Aliases {
		c.Aliases[k] = filepath.Clean(v)
	}

	return &c, nil
}

func defaultConfig() *Config {
	return &Config{
		Aliases:      map[string]string{},
		Environments: map[string]Env{},
	}
}

func (c *Config) NewEnv(env string) error {
	if _, ok := c.Environments[env]; ok {
		return errors.New("env already exists")
	}

	c.Environments[env] = make(Env)

	return nil
}

func (c *Config) SetEnvValue(env, key, value string) error {
	envMap, ok := c.Environments[env]
	if !ok {
		return errors.New("unknown env")
	}

	envMap[key] = value

	return nil
}

func (c *Config) DeleteEnvValue(env, key string) error {
	envMap, ok := c.Environments[env]
	if !ok {
		return errors.New("unknown env")
	}

	delete(envMap, key)

	return nil
}
