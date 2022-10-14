package config

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Values struct {
	LogUpdates bool    `yaml:"log-updates"`
	WhiteList  []int64 `yaml:"white-list"`
	BlackList  []int64 `yaml:"black-list"`
}

type config struct {
	Token  string `yaml:"token"`
	Values `yaml:",inline"`
}

type Service struct {
	config config
}

func New(path string) (*Service, error) {
	s := &Service{}

	rawYAML, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, errors.Wrap(err, "reading config file")
	}

	err = yaml.Unmarshal(rawYAML, &s.config)
	if err != nil {
		return nil, errors.Wrap(err, "parsing yaml")
	}

	return s, nil
}

func (s *Service) Values() *Values {
	return &s.config.Values
}

func (s *Service) Token() string {
	return s.config.Token
}
