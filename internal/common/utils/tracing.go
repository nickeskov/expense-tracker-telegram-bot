package utils

import (
	"io"

	"github.com/pkg/errors"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
)

func InitTracing(serviceName string, logger *zap.Logger) (io.Closer, error) {
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
	}
	flusher, err := cfg.InitGlobalTracer(serviceName,
		config.Logger(zapWrapper{logger: logger}),
		config.Metrics(prometheus.New()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init global tracer")
	}
	return flusher, nil
}

type zapWrapper struct {
	logger *zap.Logger
}

func (w zapWrapper) Error(msg string) {
	w.logger.Error(msg)
}

func (w zapWrapper) Infof(msg string, args ...interface{}) {
	w.logger.Sugar().Infof(msg, args...)
}
