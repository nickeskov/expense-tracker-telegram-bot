package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

const (
	defaultFetchRequestTimeout = 5 * time.Second
	fetchByDateLayout          = "2006-01-02"
	exchangeRatesAPIURL        = "https://api.exchangerate.host"
)

const (
	fetchDateUnixMillisSpanTagKey = "fetch_date_unix_ms"
)

type ExchangeRatesWebProvider struct {
	apiURL              string
	baseCurrency        models.CurrencyCode
	supportedCurrencies map[models.CurrencyCode]struct{}
	client              *http.Client
}

func NewExchangeRatesWebProvider(base models.CurrencyCode, supported []models.CurrencyCode) (*ExchangeRatesWebProvider, error) {
	return newExchangeRatesWebProvider(exchangeRatesAPIURL, base, supported)
}

func newExchangeRatesWebProvider(apiURL string, base models.CurrencyCode, supported []models.CurrencyCode) (*ExchangeRatesWebProvider, error) {
	supportedCurrencies := make(map[models.CurrencyCode]struct{}, len(supported))
	for _, code := range supported {
		supportedCurrencies[code] = struct{}{}
	}
	return &ExchangeRatesWebProvider{
		apiURL:              apiURL,
		baseCurrency:        base,
		supportedCurrencies: supportedCurrencies,
		client:              http.DefaultClient,
	}, nil
}

func (e *ExchangeRatesWebProvider) FetchExchangeRates(ctx context.Context, date time.Time) (_ []models.ExchangeRate, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "FetchExchangeRates")
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()
	span.SetTag(fetchDateUnixMillisSpanTagKey, date.UnixMilli())

	data, err := e.fetchData(ctx, date)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch exchange rates data at time=%v", date)
	}
	rates := data.ToExchangeRates(date, func(rate *models.ExchangeRate) bool {
		_, ok := e.supportedCurrencies[rate.Code]
		return ok
	})
	return rates, nil
}

type exchangeRatesData struct {
	Rates map[models.CurrencyCode]decimal.Decimal `json:"rates"`
}

func (d *exchangeRatesData) ToExchangeRates(date time.Time, filter func(rate *models.ExchangeRate) bool) []models.ExchangeRate {
	var rates []models.ExchangeRate
	for code, rateValue := range d.Rates {
		rate := models.NewExchangeRate(code, rateValue, date)
		if filter(&rate) {
			rates = append(rates, rate)
		}
	}
	return rates
}

func makeExchangeRatesURL(baseURL string, date time.Time, base models.CurrencyCode) string {
	return fmt.Sprintf("%s/%s?base=%s", strings.TrimSuffix(baseURL, "/"), date.Format(fetchByDateLayout), base)
}

func (e *ExchangeRatesWebProvider) fetchData(ctx context.Context, date time.Time) (exchangeRatesData, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultFetchRequestTimeout)
	defer cancel()

	var (
		apiURL = makeExchangeRatesURL(e.apiURL, date, e.baseCurrency)
		method = http.MethodGet
	)
	req, err := http.NewRequestWithContext(ctx, method, apiURL, nil)
	if err != nil {
		return exchangeRatesData{}, errors.Wrapf(err, "failed to build HTTP %q request to %q", method, apiURL)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return exchangeRatesData{}, errors.Wrapf(err, "failed to do HTTP %q request to %q", method, apiURL)
	}
	defer resp.Body.Close()

	if c := resp.StatusCode; c != http.StatusOK {
		return exchangeRatesData{}, errors.Wrapf(err, "HTTP %q endpoint %q returned non-OK code (%d)", method, apiURL, c)
	}

	var data exchangeRatesData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return exchangeRatesData{}, errors.Wrap(err, "failed to decode response as JSON")
	}
	return data, nil
}
