package kafka

import (
	"context"

	"github.com/Shopify/sarama"
)

type ProducerMessage sarama.ProducerMessage

type MessageSender interface {
	SendMessage(context.Context, *ProducerMessage) error
}

type MessagesSender interface {
	MessageSender
	SendMessages(context.Context, []*ProducerMessage) error
}

type (
	StringEncoder = sarama.StringEncoder
	ByteEncoder   = sarama.ByteEncoder
)

type AsyncProducer struct {
	inner sarama.AsyncProducer
	rg    *runGroup
}

func (p *AsyncProducer) SendMessage(ctx context.Context, msg *ProducerMessage) error {
	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}
	p.inner.Input() <- (*sarama.ProducerMessage)(msg)
	return nil
}

func (p *AsyncProducer) SendMessages(ctx context.Context, msgs []*ProducerMessage) error {
	for _, msg := range msgs {
		if err := p.SendMessage(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

func (p *AsyncProducer) Close() error {
	defer p.rg.CancelAndWait()
	return p.inner.Close()
}

type SyncProducer struct {
	inner sarama.SyncProducer
	rg    *runGroup
}

func (p *SyncProducer) SendMessage(ctx context.Context, msg *ProducerMessage) error {
	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}
	_, _, err := p.inner.SendMessage((*sarama.ProducerMessage)(msg))
	return err
}

func (p *SyncProducer) SendMessages(ctx context.Context, msgs []*ProducerMessage) error {
	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}
	saramaMsgs := make([]*sarama.ProducerMessage, len(msgs))
	for i := range msgs {
		saramaMsgs[i] = (*sarama.ProducerMessage)(msgs[i])
	}
	return p.inner.SendMessages(saramaMsgs)
}

func (p *SyncProducer) Close() error {
	defer p.rg.CancelAndWait()
	return p.inner.Close()
}
