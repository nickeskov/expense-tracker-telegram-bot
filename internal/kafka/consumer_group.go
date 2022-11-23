package kafka

import (
	"context"

	"github.com/Shopify/sarama"
)

var (
	defaultBalanceStrategy       = sarama.BalanceStrategyRange
	supportedBalancingStrategies = map[string]sarama.BalanceStrategy{
		sarama.RangeBalanceStrategyName:      sarama.BalanceStrategyRange,
		sarama.RoundRobinBalanceStrategyName: sarama.BalanceStrategyRoundRobin,
		sarama.StickyBalanceStrategyName:     sarama.BalanceStrategySticky,
	}
)

type ConsumerGroup struct {
	inner sarama.ConsumerGroup
	rg    *runGroup
}

func (c *ConsumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	return c.inner.Consume(ctx, topics, handler)
}

func (c *ConsumerGroup) Close() error {
	defer c.rg.CancelAndWait()
	return c.inner.Close()
}
