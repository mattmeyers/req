package req

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Root    string            `toml:"root"`
	Aliases map[string]string `toml:"aliases"`
}

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
	return &Config{}
}
