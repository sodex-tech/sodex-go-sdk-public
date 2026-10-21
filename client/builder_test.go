package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const builderOwner = "0x1111111111111111111111111111111111111111"
const builderSummaryJSON = `{"address":"0x1111111111111111111111111111111111111111","name":"Named Builder","builderCode":"9007199254740993","userCount":"18446744073709551615","builderFee":"0.300000000000000001","tradeCount":"9007199254740993","tradeVolume":"300.000000000000000001"}`
const builderDailyJSON = `[{"date":"2026-10-20","userCount":"1","builderFee":"0.2","tradeCount":"4","tradeVolume":"200"},{"date":"2026-10-21","userCount":"18446744073709551615","builderFee":"0.100000000000000001","tradeCount":"9007199254740989","tradeVolume":"100.000000000000000001"}]`

// TestBuilderFeeQueries verifies distinct summary/daily payloads, routes, filters and lossless uint64/decimal round trips.
func TestBuilderFeeQueries(t *testing.T) {
	for _, filter := range []HistoryFilter{{}, {AccountID: 9007199254740993, Limit: 300, Symbol: "ignored", StartTime: 1, EndTime: 2}} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Empty(t, r.Header.Get("X-API-Sign"))
			want := url.Values{}
			if filter.AccountID > 0 {
				want.Set("accountID", strconv.FormatUint(filter.AccountID, 10))
			}
			switch r.URL.Path {
			case "/api/v1/user/" + builderOwner + "/builder/summary":
				_, _ = io.WriteString(w, `{"code":0,"data":`+builderSummaryJSON+`}`)
			case "/api/v1/user/" + builderOwner + "/builder/daily-chart":
				if filter.Limit > 0 {
					want.Set("limit", strconv.Itoa(filter.Limit))
				}
				_, _ = io.WriteString(w, `{"code":0,"data":`+builderDailyJSON+`}`)
			default:
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			assert.Equal(t, want, r.URL.Query())
		}))
		c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
		summary, err := c.BuilderFeeSummary(context.Background(), builderOwner, filter)
		require.NoError(t, err)
		require.NotNil(t, summary)
		require.EqualValues(t, 9007199254740993, summary.BuilderCode)
		require.Equal(t, ^uint64(0), summary.UserCount)
		require.EqualValues(t, 9007199254740993, summary.TradeCount)
		require.Equal(t, "Named Builder", summary.Name)
		require.Equal(t, builderOwner, summary.Address)
		require.Equal(t, "0.300000000000000001", summary.BuilderFee)
		require.Equal(t, "300.000000000000000001", summary.TradeVolume)
		encoded, err := json.Marshal(summary)
		require.NoError(t, err)
		require.JSONEq(t, builderSummaryJSON, string(encoded))
		chart, err := c.BuilderFeeDailyChart(context.Background(), builderOwner, filter)
		require.NoError(t, err)
		require.Len(t, chart, 2)
		require.Equal(t, "2026-10-20", chart[0].Date)
		require.Equal(t, "2026-10-21", chart[1].Date)
		require.EqualValues(t, 1, chart[0].UserCount)
		require.Equal(t, ^uint64(0), chart[1].UserCount)
		require.EqualValues(t, 9007199254740989, chart[1].TradeCount)
		require.Equal(t, "0.100000000000000001", chart[1].BuilderFee)
		require.Equal(t, "100.000000000000000001", chart[1].TradeVolume)
		encoded, err = json.Marshal(chart)
		require.NoError(t, err)
		require.JSONEq(t, builderDailyJSON, string(encoded))
		srv.Close()
	}
}

// TestBuilderFeeEmptyAndError preserves null summaries, empty charts and application errors from both routes.
func TestBuilderFeeEmptyAndError(t *testing.T) {
	for _, failure := range []bool{false, true} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Empty(t, r.URL.RawQuery)
			if failure {
				_, _ = io.WriteString(w, `{"code":-1,"error":"invalid parameter: accountID"}`)
			} else if r.URL.Path == "/api/v1/user/"+builderOwner+"/builder/summary" {
				_, _ = io.WriteString(w, `{"code":0,"data":null}`)
			} else {
				_, _ = io.WriteString(w, `{"code":0,"data":[]}`)
			}
		}))
		c := New(Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
		summary, err := c.BuilderFeeSummary(context.Background(), builderOwner, HistoryFilter{})
		if failure {
			require.ErrorContains(t, err, "invalid parameter: accountID")
		} else {
			require.NoError(t, err)
			require.Nil(t, summary)
		}
		chart, err := c.BuilderFeeDailyChart(context.Background(), builderOwner, HistoryFilter{})
		if failure {
			require.ErrorContains(t, err, "invalid parameter: accountID")
		} else {
			require.NoError(t, err)
			require.Empty(t, chart)
		}
		srv.Close()
	}
}
