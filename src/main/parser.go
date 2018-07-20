package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// EnvParser represents the actions for parsing a file
// It will read hosts information from the config file(hc.config.yml and address.yml)
type EnvParser interface {
	GetParsedData() interface{}
	Parse() error
}

// EnvFinder abstracts what EnvParser should do to find configuration
type EnvFinder interface {
	FindEnv(string) []string
	FindHost(string, string) string
	FindGroup(string) []string
}

// YamlEnvFinder do the work on data read from the yaml file
type YamlEnvFinder struct {
	parser EnvParser
}

// FindEnv finds target env ip configs
func (y YamlEnvFinder) FindEnv(env string) []string {
	hosts := y.parser.GetParsedData().(envConfig).EnvRule[env]
	return hosts
}

// FindHost finds host
func (y YamlEnvFinder) FindHost(env, host string) string {
	conf := y.parser.GetParsedData().(envConfig)
	matchingRule, exists := conf.EnvRule[env]
	if !exists {
		return ""
	}

	ips, exists := conf.Address[host]
	if !exists {
		return ""
	}

	for _, ip := range ips {
		for _, rule := range matchingRule {
			if strings.HasPrefix(ip, rule) {
				return ip
			}
		}
	}
	return ""
}

// FindGroup finds group information
func (y YamlEnvFinder) FindGroup(group string) []string {
	result, exists := y.parser.GetParsedData().(envConfig).Group[group]
	if !exists {
		return []string{}
	}
	return result
}

// NewYamlEnvFinder creates a new YamlEnvFinder instance
func NewYamlEnvFinder(parser EnvParser) YamlEnvFinder {
	return YamlEnvFinder{
		parser: parser,
	}
}

// YamlEnvParser parses the yml file of env config file
// It has dependency of gopkg.in/ymal.v2
type YamlEnvParser struct {
	base     string
	filename string
	ec       envConfig
}

// GetParsedData returns the parsed data
func (y YamlEnvParser) GetParsedData() interface{} {
	return y.ec
}

// Parse the yaml file & and store results using the library
func (y YamlEnvParser) Parse() error {
	var data []byte
	var err error
	data, err = ioutil.ReadFile(y.base + y.filename)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	err = yaml.Unmarshal(data, &y.ec)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
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
