package httpclient

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"time"
)

// ErrRequestTimeout is an error that's returned when the max timeout is reached for a request
var ErrRequestTimeout = errors.New("httpclient: timeout")

// ErrInvalidClientImpl is an error that's returned when the Client Do returns nil to both response and error
var ErrInvalidClientImpl = errors.New("httpclient: invalid response from Do")

// this is a catch all that will prevent a Retryable going over a predefined threshold in case it has a bug
const maxAttempts = 100

// Config is the configuration for the HTTPClient
type Config struct {
	Paginator Paginator
	Retryable Retryable
}

// NewConfig returns an empty Config by no pagination and no retry
func NewConfig() *Config {
	return &Config{
		Paginator: &noPaginator{},
		Retryable: &noRetry{},
	}
}

// HTTPClient is an implementation of the Client interface
type HTTPClient struct {
	config *Config
	ctx    context.Context
	c      Client
}

// ensure that we implement the Client interface
var _ Client = (*HTTPClient)(nil)

// NewHTTPClientDefault returns a default HTTPClient instance
func NewHTTPClientDefault() *HTTPClient {
	return &HTTPClient{
		config: NewConfig(),
		ctx:    context.Background(),
		c:      http.DefaultClient,
	}
}

// NewHTTPClient returns a configured HTTPClient instance
func NewHTTPClient(ctx context.Context, config *Config, client Client) *HTTPClient {
	if config == nil {
		config = NewConfig()
	}
	if client == nil {
		client = http.DefaultClient
	}
	if config.Paginator == nil {
		config.Paginator = NoPaginator()
	}
	if config.Retryable == nil {
		config.Retryable = NewNoRetry()
	}
	return &HTTPClient{
		config: config,
		ctx:    ctx,
		c:      client,
	}
}

// Default can be used to get a basic implementation of Client interface
var Default Client = NewHTTPClientDefault()

// Get is a convenience method for making a Get request to a url
func (c *HTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post is a convenience method for making a Post request to a url
func (c *HTTPClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// Do will invoke the http request
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	var count, page int
	var streams *multiReader
	started := time.Now()
	maxDuration := c.config.Retryable.RetryMaxDuration()
	req = req.WithContext(c.ctx)
	for time.Since(started) < maxDuration && count+1 < maxAttempts {
		count++
		page++
		resp, err := c.c.Do(req)
		if resp == nil && err == nil {
			return nil, ErrInvalidClientImpl
		}
		if err != nil {
			if !c.config.Retryable.RetryError(err) {
				return nil, err
			}
		} else {
			// if OK and a GET request type, see if we need to paginate
			if resp.StatusCode == http.StatusOK && req.Method == http.MethodGet {
				if ok, newreq := c.config.Paginator.HasMore(page, req, resp); ok {
					// reset our count and timestamp since we're going to loop and it's OK
					count = 0
					started = time.Now()
					if streams == nil {
						streams = newMuliReader()
					}
					// remember our stream since we're going to need to return it instead
					streams.Add(resp.Body)
					// don't reuse this request again
					req.Close = true
					// assign our new request for the loop
					req = newreq
					continue
				}
			}
			// if this request looks like a normal, non-retryable response
			// then just return it without attempting a retry
			if (resp.StatusCode >= 200 && resp.StatusCode < 300) ||
				resp.StatusCode == http.StatusUnauthorized ||
				resp.StatusCode == http.StatusPaymentRequired ||
				resp.StatusCode == http.StatusForbidden ||
				resp.StatusCode == http.StatusNotFound ||
				resp.StatusCode == http.StatusMethodNotAllowed ||
				resp.StatusCode == http.StatusPermanentRedirect ||
				resp.StatusCode == http.StatusTemporaryRedirect {
				// check to see if we have a multiple stream response (pagination)
				if streams != nil && resp.Body != nil {
					streams.Add(resp.Body)
					resp.Body = streams
				}
				return resp, nil
			}
			if !c.config.Retryable.RetryResponse(resp) {
				return resp, nil
			}
			// make sure we read all (if any) content and close the response stream as to not leak resources
			if resp.Body != nil {
				ioutil.ReadAll(resp.Body)
				resp.Body.Close()
			}
		}
		duration := c.config.Retryable.RetryDelay(count)
		if duration > 0 {
			remaining := math.Min(float64(maxDuration-time.Since(started)), float64(duration))
			select {
			case <-c.ctx.Done():
				{
					return nil, context.Canceled
				}
			case <-time.After(time.Duration(remaining)):
				{
					continue
				}
			}
		}
	}
	return nil, ErrRequestTimeout
}
