// Package client provides an HTTP client for the Sodex REST API.
package client

// APIResponse is the standard JSON envelope returned by Sodex REST endpoints.
// Code == 0 means success; any non-zero value is an application-level error.
type APIResponse[T any] struct {
	Code      int     `json:"code"`
	Timestamp uint64  `json:"timestamp"`
	Data      T       `json:"data"`
	Error     *string `json:"error,omitempty"`
}

// HistoryFilter captures the shared optional pagination and time filters.
type HistoryFilter struct {
	Symbol    string
	StartTime int64
	EndTime   int64
	Limit     int
}
