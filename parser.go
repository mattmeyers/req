package req

import (
	"os"

	"gopkg.in/yaml.v3"
)

func ParseReqfile(path string) (Reqfile, error) {
	f, err := os.Open(path)
	if err != nil {
		return Reqfile{}, err
	}

	var reqfile Reqfile
	err = yaml.NewDecoder(f).Decode(&reqfile)
	if err != nil {
		return Reqfile{}, err
	}

	return reqfile, nil
}
