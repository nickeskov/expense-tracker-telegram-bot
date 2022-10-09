package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

type apiMock struct {
	t    *testing.T
	base models.CurrencyCode
	date time.Time
}

func newAPIMock(t *testing.T, base models.CurrencyCode, date time.Time) *httptest.Server {
	mock := &apiMock{
		t:    t,
		base: base,
		date: date,
	}
	s := httptest.NewServer(mock)
	t.Cleanup(s.Close)
	return s
}

func (m *apiMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const testdata = `
{
  "motd": {
    "msg": "If you or your company use this project or like what we doing, please consider backing us so we can continue maintaining and evolving this project.",
    "url": "https://exchangerate.host/#/donate"
  },
  "success": true,
  "historical": true,
  "base": "RUB",
  "date": "2022-10-09",
  "rates": {
    "AED": 0.058912,
    "EUR": 0.01648,
    "JEP": 0.014498,
    "JMD": 2.47361,
    "JPY": 2.331946,
    "KGS": 1.295059,
    "RSD": 1.931697,
    "RUB": 1,
    "RWF": 17.080029,
    "TRY": 0.298886,
    "TTD": 0.109609,
    "UAH": 0.595774,
    "UGX": 61.601773,
    "USD": 0.016044,
    "UYU": 0.65947
  }
}`
	u, err := url.Parse(makeExchangeRatesURL("", m.date, m.base))
	require.NoError(m.t, err)
	assert.Equal(m.t, u.Path, r.URL.Path)

	_, err = w.Write([]byte(testdata))
	require.NoError(m.t, err)
}

func TestExchangeRatesWebProvider_FetchExchangeRates(t *testing.T) {
	var (
		base      = models.CurrencyCode("RUB")
		supported = []models.CurrencyCode{"EUR", "USD", "TRY"}
		date      = time.Date(2022, 10, 9, 0, 0, 0, 0, time.UTC)
		expected  = []models.ExchangeRate{
			{Code: "USD", Rate: 0.016044, Date: date},
			{Code: "EUR", Rate: 0.01648, Date: date},
			{Code: "TRY", Rate: 0.298886, Date: date},
		}
	)
	s := newAPIMock(t, base, date)

	p, err := newExchangeRatesWebProvider(s.URL, base, supported)
	require.NoError(t, err)
	actual, err := p.FetchExchangeRates(context.Background(), date)
	require.NoError(t, err)
	assert.ElementsMatch(t, expected, actual)
}
