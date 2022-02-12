package req

import "github.com/BurntSushi/toml"

type Config struct {
	Aliases map[string]string `toml:"aliases"`
}

func ParseConfig(path string) (*Config, error) {
	if path == "" {
		path = "./.reqrc"
	}

	var c Config
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
