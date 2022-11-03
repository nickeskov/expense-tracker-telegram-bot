package tg

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

func createIncomingUpdatesLoggerMiddleware(logger *zap.Logger) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(teleCtx telebot.Context) error {
			data, err := json.Marshal(teleCtx.Update())
			if err != nil {
				return errors.Wrap(err, "failed to marshal incoming update")
			}
			var (
				updateID  = teleCtx.Update().ID
				messageID *int
				senderID  *int64
			)
			if msg := teleCtx.Message(); msg != nil {
				messageID = &msg.ID
			}
			if sender := teleCtx.Sender(); sender != nil {
				senderID = &sender.ID
			}
			logger.Info("Received incoming update",
				zap.Int("update_id", updateID),
				zap.Intp("message_id", messageID),
				zap.Int64p("sender_id", senderID),
				zap.ByteString("update", data),
			)
			return next(teleCtx)
		}
	}
}

func createTriggeredHandlerLoggerMiddleware(logger *zap.Logger, endpoint string) telebot.MiddlewareFunc {
	stringEndpoint := strings.Trim(strconv.Quote(endpoint), "\"")
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(teleCtx telebot.Context) error {
			var (
				updateID  = teleCtx.Update().ID
				messageID *int
				senderID  *int64
			)
			if msg := teleCtx.Message(); msg != nil {
				messageID = &msg.ID
			}
			if sender := teleCtx.Sender(); sender != nil {
				senderID = &sender.ID
			}
			logger.Info("Endpoint triggered with update",
				zap.String("endpoint", stringEndpoint),
				zap.Int("update_id", updateID),
				zap.Intp("message_id", messageID),
				zap.Int64p("sender_id", senderID),
			)
			return next(teleCtx)
		}
	}
}
