package utils

import (
	"go.uber.org/zap"
)

func SetupZapLogger(devMode bool, level string) (l *zap.Logger, al zap.AtomicLevel, err error) {
	al = zap.NewAtomicLevel()
	if err = al.UnmarshalText([]byte(level)); err != nil {
		return nil, zap.AtomicLevel{}, err
	}
	var cfg zap.Config
	if devMode {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.Level = al
	l, err = cfg.Build()
	return l, al, err
}
