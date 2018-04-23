package main

import (
	"os"
	"fmt"
	"errors"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/wdxxs2z/helmi/pkg/config"
)

type Config struct {
	LogLevel 			string        		`yaml:"log_level"`
	Username 			string        		`yaml:"username"`
	Password 			string        		`yaml:"password"`
	Platform			string			`yaml:"platform"`
	HelmiConfig			config.Config		`yaml:"helmi_config"`
}

func LoadConfig(configFile string) (config *Config, err error) {
	if configFile == "" {
		return config, errors.New("Must provide a config file")
	}

	file, err := os.Open(configFile)
	if err != nil {
		return config, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return config, err
	}

	if err = yaml.Unmarshal(bytes, &config); err != nil {
		return config, err
	}

	if err = config.Validate(); err != nil {
		return config, fmt.Errorf("Validating config contents: %s", err)
	}

	return config, nil
}

func (c Config) Validate() error {

	if c.LogLevel == "" {
		return errors.New("Must provide a non-empty LogLevel")
	}

	if c.Username == "" {
		return errors.New("Must provide a non-empty Username")
	}

	if c.Password == "" {
		return errors.New("Must provide a non-empty Password")
	}

	return nil
}