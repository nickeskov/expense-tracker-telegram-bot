package kafka

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
	prometheusmetrics "github.com/deathowl/go-metrics-prometheus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rcrowley/go-metrics"
)

type runGroup struct {
	wg     sync.WaitGroup
	cancel func()
}

func (rg *runGroup) Go(fn func()) {
	rg.wg.Add(1)
	go func() {
		defer rg.wg.Done()
		fn()
	}()
}

func (rg *runGroup) GoWithCancel(fn func(done <-chan struct{})) {
	doneCh := make(chan struct{})
	markDoneFn := func() {
		close(doneCh)
	}
	if cancel := rg.cancel; cancel != nil {
		rg.cancel = func() {
			cancel()
			markDoneFn()
		}
	} else {
		rg.cancel = markDoneFn
	}
	rg.Go(func() {
		fn(doneCh)
	})
}

func (rg *runGroup) CancelAndWait() {
	if cancel := rg.cancel; cancel != nil {
		cancel()
		rg.cancel = nil
	}
	rg.wg.Wait()
}

type successHandler func(<-chan *sarama.ProducerMessage)

type Config struct {
	*sarama.Config
	successHandler successHandler
	prom           struct {
		provided          bool
		publisher         *prometheusmetrics.PrometheusConfig
		publishErrHandler func(err error)
	}
}

func NewConfig() *Config {
	c := sarama.NewConfig()
	c.Version = sarama.V2_8_0_0
	return &Config{Config: c}
}

func (c *Config) WithMetrics(
	registerer prometheus.Registerer,
	namespace, subsystem string,
	flushInterval time.Duration,
	publishErrHandler func(error),
) *Config {
	if c.MetricRegistry == nil {
		c.MetricRegistry = metrics.NewRegistry()
	}
	c.prom.provided = true
	c.prom.publisher = prometheusmetrics.NewPrometheusProvider(c.MetricRegistry, namespace, subsystem, registerer, flushInterval)
	c.prom.publishErrHandler = publishErrHandler
	return c
}

func (c *Config) WithSuccessHandler(h successHandler) *Config {
	c.Producer.Return.Successes = true
	c.successHandler = h
	return c
}

func (c *Config) tryRunProm(rg *runGroup) {
	prom := c.prom
	if !prom.provided {
		return
	}
	rg.GoWithCancel(func(done <-chan struct{}) {
		ticker := time.NewTicker(prom.publisher.FlushInterval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := prom.publisher.UpdatePrometheusMetricsOnce(); err != nil {
					if prom.publishErrHandler != nil {
						prom.publishErrHandler(err)
					}
				}
			}
		}
	})
}

func (c *Config) BuildAsyncProducer(brokersList []string, errHandler func(<-chan *sarama.ProducerError)) (*AsyncProducer, error) {
	if errHandler == nil {
		return nil, errors.New("<nil> errors handler")
	}
	c.Producer.Return.Errors = true
	inner, err := sarama.NewAsyncProducer(brokersList, c.Config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create new async producer")
	}
	rg := new(runGroup)
	rg.Go(func() {
		errHandler(inner.Errors())
	})
	if c.successHandler != nil {
		rg.Go(func() {
			c.successHandler(inner.Successes())
		})
	}
	c.tryRunProm(rg)
	return &AsyncProducer{
		inner: inner,
		rg:    rg,
	}, nil
}

func (c *Config) BuildSyncProducer(brokersList []string) (*SyncProducer, error) {
	inner, err := sarama.NewSyncProducer(brokersList, c.Config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create new sync producer")
	}
	if c.Producer.Idempotent {
		// idempotent producer has a unique producer ID and uses sequence IDs for each message,
		// allowing the broker to ensure, on a per-partition basis, that it is committing ordered messages with no duplication.
		c.Producer.Retry.Max = 1
		c.Net.MaxOpenRequests = 1
	}
	rg := new(runGroup)
	c.tryRunProm(rg)
	return &SyncProducer{
		inner: inner,
		rg:    rg,
	}, nil
}

func (c *Config) BuildConsumerGroup(brokersList []string, consumerGroup string, balanceStrategies ...string) (*ConsumerGroup, error) {
	var strategies []sarama.BalanceStrategy
	if len(balanceStrategies) == 0 {
		strategies = []sarama.BalanceStrategy{defaultBalanceStrategy}
	} else {
		for _, strategyName := range balanceStrategies {
			strategy, ok := supportedBalancingStrategies[strategyName]
			if !ok {
				return nil, errors.Errorf("unsupported balance strategy %q", strategyName)
			}
			strategies = append(strategies, strategy)
		}
	}
	c.Consumer.Group.Rebalance.GroupStrategies = strategies

	inner, err := sarama.NewConsumerGroup(brokersList, consumerGroup, c.Config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create consumer group %q", consumerGroup)
	}
	rg := new(runGroup)
	c.tryRunProm(rg)
	return &ConsumerGroup{
		inner: inner,
		rg:    rg,
	}, nil
}
