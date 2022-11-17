package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type ReportsRedisCache struct {
	redisDB *redis.Client
}

func NewReportsRedisCache(redisDB *redis.Client) (*ReportsRedisCache, error) {
	return &ReportsRedisCache{redisDB: redisDB}, nil
}

func makeUserReportsHashSetName(userID models.UserID) string {
	return "user_reports_" + strconv.FormatInt(int64(userID), 10)
}

func makeUserReportHashSetKey(since, till time.Time) string {
	return strconv.FormatInt(since.UTC().Unix(), 10) + ":" + strconv.FormatInt(till.UTC().Unix(), 10)
}

func (c *ReportsRedisCache) AddToCache(ctx context.Context, userID models.UserID, since, till time.Time, report expense.SummaryReport) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "AddToCache")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()

	data, err := cbor.Marshal(report)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal report of userID=(%d) since=%q till=%q", userID, since, till)
	}
	hashSetName := makeUserReportsHashSetName(userID)
	hashSetReportKey := makeUserReportHashSetKey(since, till)
	if err := c.redisDB.HSet(ctx, hashSetName, hashSetReportKey, data).Err(); err != nil {
		return errors.Wrapf(err, "failed to set data to redis hashSet=%q by key=%q", hashSetName, hashSetReportKey)
	}
	return nil
}

func (c *ReportsRedisCache) GetFromCache(ctx context.Context, userID models.UserID, since, till time.Time) (_ expense.SummaryReport, _ bool, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetFromCache")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()

	hashSetName := makeUserReportsHashSetName(userID)
	hashSetReportKey := makeUserReportHashSetKey(since, till)
	data, err := c.redisDB.HGet(ctx, hashSetName, hashSetReportKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, errors.Wrapf(err, "failed to get data from redis hashSet=%q by key=%q", hashSetName, hashSetReportKey)
	}
	report := make(expense.SummaryReport)
	if err := cbor.Unmarshal(data, &report); err != nil {
		return nil, false, errors.Wrapf(err, "failed to unmarshal report of userID=(%d) since=%q till=%q", userID, since, till)
	}
	return report, true, nil
}

func (c *ReportsRedisCache) DropCacheForUserID(ctx context.Context, userID models.UserID) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "DropCacheForUserID")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()

	hashSetName := makeUserReportsHashSetName(userID)
	if err := c.redisDB.Del(ctx, hashSetName).Err(); err != nil {
		return errors.Wrapf(err, "failed to del data by key=%q from redis", hashSetName)
	}
	return nil
}
