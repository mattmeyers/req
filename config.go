package req

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
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

	return &c, nil
}

func defaultConfig() *Config {
	return &Config{}
}
