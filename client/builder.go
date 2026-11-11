package client

import (
	"context"
	"net/url"
	"strconv"

	rpctypes "github.com/sodex-tech/sodex-go-sdk-public/client/types"
)

// BuilderFeeSummary returns combined Spot/Perps builder statistics, or nil when unavailable.
func (c *Client) BuilderFeeSummary(ctx context.Context, address string, filter HistoryFilter) (*rpctypes.BuilderFeeStats, error) {
	path := "/api/v1/user/" + address + "/builder/summary"
	if filter.AccountID > 0 {
		path += "?accountID=" + strconv.FormatUint(filter.AccountID, 10)
	}
	var result *rpctypes.BuilderFeeStats
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// BuilderFeeDailyChart returns daily builder statistics in ascending date order.
func (c *Client) BuilderFeeDailyChart(ctx context.Context, address string, filter HistoryFilter) ([]*rpctypes.BuilderFeeDailyStats, error) {
	path := "/api/v1/user/" + address + "/builder/daily-chart"
	query := url.Values{}
	if filter.AccountID > 0 {
		query.Set("accountID", strconv.FormatUint(filter.AccountID, 10))
	}
	if filter.Limit > 0 {
		query.Set("limit", strconv.Itoa(filter.Limit))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	var result []*rpctypes.BuilderFeeDailyStats
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}
