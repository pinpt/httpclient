package httpclient

import (
	"io"
	"net/http"
	"time"
)

// Client is an interface that mimics http.Client
type Client interface {
	Get(url string) (*http.Response, error)
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}

// Retryable is an interface for retrying a request
type Retryable interface {
	RetryError(err error) bool
	RetryResponse(resp *http.Response) bool
	RetryDelay(retry int) time.Duration
	RetryMaxDuration() time.Duration
}

// Paginator is an interface for handling request pagination
type Paginator interface {
	HasMore(page int, req *http.Request, resp *http.Response) (bool, *http.Request)
}
