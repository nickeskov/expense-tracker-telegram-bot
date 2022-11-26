package usecase

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/kafka"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

const (
	chatIDSpanTagKey = "chat_id"
)

type ExtendedUseCase struct {
	uc           *UseCase
	reportsTopic string
	sender       kafka.MessageSender
}

func NewExtendedUseCase(uc *UseCase, reportsTopic string, sender kafka.MessageSender) (*ExtendedUseCase, error) {
	return &ExtendedUseCase{uc: uc, reportsTopic: reportsTopic, sender: sender}, nil
}

func (u *ExtendedUseCase) AddExpense(ctx context.Context, userID models.UserID, expense models.Expense) (models.Expense, error) {
	return u.uc.AddExpense(ctx, userID, expense)
}

func (u *ExtendedUseCase) GetExpensesSummaryByCategorySince(ctx context.Context, userID models.UserID, since, till time.Time) (expense.SummaryReport, error) {
	return u.uc.GetExpensesSummaryByCategorySince(ctx, userID, since, till)
}

func (u *ExtendedUseCase) GetExpensesAscendSinceTill(ctx context.Context, userID models.UserID, since, till time.Time, max int) ([]models.Expense, error) {
	return u.uc.GetExpensesAscendSinceTill(ctx, userID, since, till, max)
}

func (u *ExtendedUseCase) SendGetExpensesSummaryByCategorySinceRequest(ctx context.Context, chatID int64, userID models.UserID, since, till time.Time) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "SendGetExpensesSummaryByCategorySinceRequest")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(userIDSpanTagKey, userID)
	span.SetTag(chatIDSpanTagKey, chatID)
	span.SetTag(sinceUnixMillisSpanTagKey, since.UnixMilli())
	span.SetTag(tillUnixMillisSpanTagKey, till.UnixMilli())

	event := expense.EventGenerateSummaryReportByCategories{
		ChatID: chatID,
		UserID: userID,
		Since:  since,
		Till:   till,
	}
	data, err := event.MarshalBinary()
	if err != nil {
		return errors.Wrapf(err, "failed to binary marshal event (%T) for chatID=%d and userID=%d", event, chatID, userID)
	}
	msg := &kafka.ProducerMessage{
		Topic: u.reportsTopic,
		Value: kafka.ByteEncoder(data),
	}
	if err := u.sender.SendMessage(ctx, msg); err != nil {
		return errors.Wrapf(err, "failed to send message to topic %q for chatID=%d and userID=%d", u.reportsTopic, chatID, userID)
	}
	return nil
}
