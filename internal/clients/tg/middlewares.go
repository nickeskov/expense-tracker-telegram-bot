package tg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

const (
	updateKeyName    = "update"
	endpointKeyName  = "endpoint"
	updateIDKeyName  = "update_id"
	messageIDKeyName = "message_id"
	senderIDKeyName  = "sender_id"
)

func createEndpointTracingMiddleware(endpoint string) func(endpointHandler) endpointHandler {
	operationName := fmt.Sprintf("bot handler '%s'", botEndpointString(endpoint))
	return func(next endpointHandler) endpointHandler {
		return func(ctx context.Context, teleCtx telebotReducedContext) (err error) {
			span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
			defer func() {
				ext.Error.Set(span, err != nil)
				span.Finish()
			}()

			updateID := teleCtx.Update().ID
			span.SetTag(updateIDKeyName, updateID)

			if msg := teleCtx.Message(); msg != nil {
				messageID := msg.ID
				span.SetTag(messageIDKeyName, messageID)
			}
			if sender := teleCtx.Sender(); sender != nil {
				senderID := sender.ID
				span.SetTag(senderIDKeyName, senderID)
			}
			return next(ctx, teleCtx)
		}
	}
}

func botEndpointString(endpoint string) string {
	return strings.Trim(strconv.Quote(endpoint), "\"")
}

func createEndpointMetricsMiddleware(endpoint string) telebot.MiddlewareFunc {
	stringEndpoint := botEndpointString(endpoint)
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(teleCtx telebot.Context) error {
			start := time.Now()
			err := next(teleCtx)
			duration := time.Since(start)
			errStatus := strconv.FormatBool(err != nil)
			inFlightRequests.WithLabelValues(stringEndpoint, errStatus).Inc()
			inFlightRequestsDuration.WithLabelValues(stringEndpoint, errStatus).Observe(duration.Seconds())
			return err
		}
	}
}

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
				zap.Int(updateIDKeyName, updateID),
				zap.Intp(messageIDKeyName, messageID),
				zap.Int64p(senderIDKeyName, senderID),
				zap.ByteString(updateKeyName, data),
			)
			return next(teleCtx)
		}
	}
}

func createTriggeredHandlerLoggerMiddleware(logger *zap.Logger, endpoint string) telebot.MiddlewareFunc {
	stringEndpoint := botEndpointString(endpoint)
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
				zap.String(endpointKeyName, stringEndpoint),
				zap.Int(updateIDKeyName, updateID),
				zap.Intp(messageIDKeyName, messageID),
				zap.Int64p(senderIDKeyName, senderID),
			)
			return next(teleCtx)
		}
	}
}

func debugMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(teleCtx telebot.Context) error {
		if err := next(teleCtx); err != nil {
			sendErr := teleCtx.Send(fmt.Sprintf("Oops, something went wrong: %v", err))
			if sendErr != nil {
				err = errors.Wrap(err, sendErr.Error())
			}
			return err
		}
		return nil
	}
}

func createRequireArgsCountMiddleware(minArgsCount, maxArgsCount int) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(teleCtx telebot.Context) error {
			args := teleCtx.Args()
			l := len(args)
			if l < minArgsCount {
				return teleCtx.Send(fmt.Sprintf("Not enough arguments: minumum required %d, provided %d", minArgsCount, l))
			}
			if l > maxArgsCount {
				return teleCtx.Send(fmt.Sprintf("Too many arguments: maximum allowed %d, provided %d", maxArgsCount, l))
			}
			return next(teleCtx)
		}
	}
}

func createIsUserExistsMiddleware(ctx context.Context, userUC user.UseCase) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(teleCtx telebot.Context) error {
			userID := models.UserID(teleCtx.Message().Sender.ID)
			exists, err := userUC.IsUserExists(ctx, userID)
			if err != nil {
				return errors.Wrapf(err, "failed to check in middleware whether the user with ID=%d exists", userID)
			}
			if exists {
				return next(teleCtx)
			}
			return teleCtx.Send(unknownUserMsg)
		}
	}
}
