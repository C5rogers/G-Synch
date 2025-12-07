package config

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var k = koanf.New(".")

var configurations map[string]string

func Load(configPath string) (map[string]string, error) {

	var err error
	path := file.Provider(configPath)

	configurations = make(map[string]string)

	err = k.Load(path, yaml.Parser())

	if err != nil {
		return nil, err
	}

	err = k.UnmarshalWithConf("", &configurations, koanf.UnmarshalConf{Tag: "koanf"})
	if err != nil {
		return nil, err
	}

	return configurations, nil
}
