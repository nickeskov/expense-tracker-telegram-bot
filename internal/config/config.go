package config

import (
	"io"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gopkg.in/yaml.v3"
)

type Values struct {
	LogUpdates                  bool                  `yaml:"log-updates"`
	Debug                       bool                  `yaml:"debug"`
	LogLevel                    string                `yaml:"log-level"`
	DevLogger                   bool                  `yaml:"dev-logger"`
	BaseCurrency                models.CurrencyCode   `yaml:"base-currency"`
	SupportedCurrencies         []models.CurrencyCode `yaml:"supported-currencies,flow"`
	ExchangeRatesUpdateInterval time.Duration         `yaml:"exchange-rates-update-interval"`
	WhiteList                   []int64               `yaml:"white-list,flow"`
	BlackList                   []int64               `yaml:"black-list,flow"`
	DBConnectionString          string                `yaml:"db-connection-string"`
}

type config struct {
	Token  string `yaml:"token"`
	Values `yaml:",inline"`
}

type Service struct {
	config config
}

func NewFromReader(r io.Reader) (*Service, error) {
	s := &Service{}

	err := yaml.NewDecoder(r).Decode(&s.config)
	if err != nil {
		return nil, errors.Wrap(err, "parsing yaml")
	}

	if s.config.Token == "" {
		return nil, errors.New("'token' parameter is required")
	}
	if s.config.BaseCurrency == "" {
		return nil, errors.New("'base-currency' parameter is required")
	}

	var baseSupported bool
	for _, currency := range s.config.SupportedCurrencies {
		if currency == s.config.BaseCurrency {
			baseSupported = true
			break
		}
	}
	if !baseSupported {
		s.config.SupportedCurrencies = append(s.config.SupportedCurrencies, s.config.BaseCurrency)
	}

	return s, nil
}

func (s *Service) Values() *Values {
	return &s.config.Values
}

func (s *Service) Token() string {
	return s.config.Token
}
