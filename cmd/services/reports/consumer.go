package main

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/generated/proto/api"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/generated/proto/types"
	"go.uber.org/zap"
)

const (
	topicKey     = "topic"
	offsetKey    = "offset"
	partitionKey = "partition"
)

const (
	userIDSpanTagKey          = "user_id"
	chatIDSpanTagKey          = "chat_id"
	sinceUnixMillisSpanTagKey = "since_unix_ms"
	tillUnixMillisSpanTagKey  = "till_unix_ms"
)

type ReportsConsumer struct {
	logger    *zap.Logger
	expenseUC expense.UseCase
	reporter  api.ReportsServiceClient
}

func NewReportsHandler(logger *zap.Logger, expenseUC expense.UseCase, reporter api.ReportsServiceClient) *ReportsConsumer {
	return &ReportsConsumer{logger: logger, expenseUC: expenseUC, reporter: reporter}
}

func (c *ReportsConsumer) Setup(session sarama.ConsumerGroupSession) error {
	c.logger.Info("Setting up reports consumer...",
		zap.String("member_id", session.MemberID()),
		zap.Int32("generation_id", session.GenerationID()),
	)
	return nil
}

func (c *ReportsConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	c.logger.Info("Cleaning up reports consumer...",
		zap.String("member_id", session.MemberID()),
		zap.Int32("generation_id", session.GenerationID()),
	)
	return nil
}

func (c *ReportsConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()
	for message := range claim.Messages() {
		session.MarkMessage(message, "")
		c.logger.Info("Message was claimed from kafka",
			zap.String(topicKey, message.Topic),
			zap.Int64(offsetKey, message.Offset),
			zap.Int32(partitionKey, message.Partition),
		)
		if err := c.handleMessage(ctx, message); err != nil {
			c.logger.Error("Failed to handle message", zap.Error(err),
				zap.String(topicKey, message.Topic),
				zap.Int64(offsetKey, message.Offset),
				zap.Int32(partitionKey, message.Partition),
			)
		}
	}
	return nil
}

func (c *ReportsConsumer) handleMessage(ctx context.Context, message *sarama.ConsumerMessage) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handleMessage")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(topicKey, message.Topic)
	span.SetTag(offsetKey, message.Offset)
	span.SetTag(partitionKey, message.Partition)

	event := new(expense.EventGenerateSummaryReportByCategories)
	if err := event.UnmarshalBinary(message.Value); err != nil {
		return errors.Wrapf(err, "failed to unmarshal binary (%T) from incomimg message value", event)
	}
	span.SetTag(userIDSpanTagKey, event.UserID)
	span.SetTag(chatIDSpanTagKey, event.ChatID)
	span.SetTag(sinceUnixMillisSpanTagKey, event.Since.UnixMilli())
	span.SetTag(tillUnixMillisSpanTagKey, event.Till.UnixMilli())

	report, err := c.expenseUC.GetExpensesSummaryByCategorySince(ctx, event.UserID, event.Since, event.Till)
	if err != nil {
		return errors.Wrapf(err, "failed to get expenses report by categories by event=%+v", event)
	}
	pbByCategories := make(map[string]*types.Decimal, len(report))
	for category, decimal := range report {
		var (
			exponent = decimal.Exponent()
			mantissa = decimal.Coefficient().Bytes()
		)
		pbByCategories[string(category)] = &types.Decimal{
			Exponent: exponent,
			Mantissa: mantissa,
		}
	}
	req := &api.SendReportRequest{
		ChatId: event.ChatID,
		UserId: (*int64)(&event.UserID),
		Report: &types.Report{Value: &types.Report_ByCategories_{
			ByCategories: &types.Report_ByCategories{Value: pbByCategories},
		}},
	}
	if _, err := c.reporter.SendReport(ctx, req); err != nil {
		return errors.Wrapf(err, "failed to send by gRPC report generated by event=%+v", event)
	}
	return nil
}
