package main

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// EnvParser represents the actions for parsing a file
// It will read hosts information from the config file(hc.config.yml and address.yml)
type EnvParser interface {
	Parse() (interface{}, error)
}

// YamlEnvParser parses the yml file of env config file
// It has dependency of gopkg.in/ymal.v2
type YamlEnvParser struct {
	base     string
	filename string
}

// Parse the yaml file & and store results using the library
func (y YamlEnvParser) Parse() (interface{}, error) {
	var data []byte
	var err error
	data, err = ioutil.ReadFile(y.base + "\\" + y.filename)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	var out envConfig
	err = yaml.Unmarshal(data, &out)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return out, nil
}

// NewYamlEnvParser creates a new YamlParser instance
func NewYamlEnvParser(name string, base string) YamlEnvParser {
	return YamlEnvParser{
		base:     base,
		filename: name,
	}
}

type envConfig struct {
	EnvRule map[string][]string `yaml:"envRule"`
	Group   map[string][]string `yaml:"group"`
	Address map[string][]string `yaml:"address"`
}
