package cache

import (
	"context"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-redis/redismock/v8"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

func newReportsRedisTestCache(t *testing.T) (*ReportsRedisCache, redismock.ClientMock) {
	db, mock := redismock.NewClientMock()
	t.Cleanup(func() {
		assert.NoError(t, db.Close())
	})
	cache, err := NewReportsRedisCache(db)
	require.NoError(t, err)
	return cache, mock
}

func marshalCBOR(t *testing.T, report expense.SummaryReport) []byte {
	data, err := cbor.Marshal(report)
	require.NoError(t, err)
	return data
}

func TestReportsRedisCache_AddToCache(t *testing.T) {
	var (
		userID = models.UserID(111)
		since  = time.Now().Truncate(24 * time.Hour).UTC()
		till   = since.AddDate(0, 0, 1)
		report = expense.SummaryReport{"test": decimal.NewFromInt32(222)}
	)
	cache, mock := newReportsRedisTestCache(t)
	var (
		hashSetName      = makeUserReportsHashSetName(userID)
		hashSetReportKey = makeUserReportHashSetKey(since, till)
		data             = marshalCBOR(t, report)
	)
	mock.ExpectHSet(hashSetName, hashSetReportKey, data).SetVal(1)

	err := cache.AddToCache(context.Background(), userID, since, till, report)
	assert.NoError(t, err)
}

func TestReportsRedisCache_GetFromCache_Hit(t *testing.T) {
	var (
		userID = models.UserID(111)
		since  = time.Now().Truncate(24 * time.Hour).UTC()
		till   = since.AddDate(0, 0, 1)
		report = expense.SummaryReport{"test": decimal.NewFromInt32(222)}
	)
	cache, mock := newReportsRedisTestCache(t)
	var (
		hashSetName      = makeUserReportsHashSetName(userID)
		hashSetReportKey = makeUserReportHashSetKey(since, till)
		data             = marshalCBOR(t, report)
	)
	mock.ExpectHGet(hashSetName, hashSetReportKey).SetVal(string(data))
	returnedReport, ok, err := cache.GetFromCache(context.Background(), userID, since, till)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, report, returnedReport)
}

func TestReportsRedisCache_GetFromCache_Miss(t *testing.T) {
	var (
		userID = models.UserID(111)
		since  = time.Now().Truncate(24 * time.Hour).UTC()
		till   = since.AddDate(0, 0, 1)
	)
	cache, mock := newReportsRedisTestCache(t)
	var (
		hashSetName      = makeUserReportsHashSetName(userID)
		hashSetReportKey = makeUserReportHashSetKey(since, till)
	)
	mock.ExpectHGet(hashSetName, hashSetReportKey).RedisNil()
	returnedReport, ok, err := cache.GetFromCache(context.Background(), userID, since, till)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, returnedReport)
}

func TestReportsRedisCache_DropCacheForUserID(t *testing.T) {
	var (
		userID      = models.UserID(111)
		hashSetName = makeUserReportsHashSetName(userID)
	)
	cache, mock := newReportsRedisTestCache(t)
	mock.ExpectDel(hashSetName).SetVal(1)

	err := cache.DropCacheForUserID(context.Background(), userID)
	require.NoError(t, err)
}
