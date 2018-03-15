package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testClient struct {
	resp *http.Response
	err  error
	req  *http.Request
}

func (r *testClient) Do(req *http.Request) (*http.Response, error) {
	r.req = req
	return r.resp, r.err
}

func (r *testClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return r.Do(req)
}

func (r *testClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return r.Do(req)
}

type testClientRetry struct {
	resps []*http.Response
	count int
}

func (r *testClientRetry) Do(req *http.Request) (*http.Response, error) {
	resp := r.resps[r.count]
	r.count++
	return resp, nil
}

func (r *testClientRetry) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return r.Do(req)
}

func (r *testClientRetry) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return r.Do(req)
}

type retry struct {
	retryError    bool
	retryResponse bool
}

func (r *retry) RetryError(err error) bool {
	return r.retryError
}

func (r *retry) RetryResponse(resp *http.Response) bool {
	return r.retryResponse
}

func (r *retry) RetryDelay(retry int) time.Duration {
	return time.Duration(0)
}

func (r *retry) RetryMaxDuration() time.Duration {
	return time.Second
}

type paginate func(page int, req *http.Request, resp *http.Response) (bool, *http.Request)

type paginator struct {
	paginate paginate
}

func (p *paginator) HasMore(page int, req *http.Request, resp *http.Response) (bool, *http.Request) {
	return p.paginate(page, req, resp)
}

func TestNewHTTPClientInvalidClient(t *testing.T) {
	assert := assert.New(t)
	tc := &testClient{}
	config := NewConfig()
	client := NewHTTPClient(context.TODO(), config, tc)
	assert.NotNil(client)
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	assert.NoError(err)
	resp, err := client.Do(req)
	assert.EqualError(err, ErrInvalidClientImpl.Error())
	assert.Nil(resp)
}

func TestNewHTTPClient(t *testing.T) {
	assert := assert.New(t)
	tc := &testClient{}
	config := NewConfig()
	client := NewHTTPClient(context.TODO(), config, tc)
	assert.NotNil(client)
	tc.resp = &http.Response{
		StatusCode: http.StatusOK,
		Body:       &testReader{},
	}
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	assert.NoError(err)
	resp, err := client.Do(req)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
}

func TestNewHTTPClientGet(t *testing.T) {
	assert := assert.New(t)
	tc := &testClient{}
	config := NewConfig()
	client := NewHTTPClient(context.TODO(), config, tc)
	assert.NotNil(client)
	tc.resp = &http.Response{
		StatusCode: http.StatusOK,
		Body:       &testReader{},
	}
	resp, err := client.Get("/test")
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal(http.MethodGet, tc.req.Method)
}

func TestNewHTTPClientPost(t *testing.T) {
	assert := assert.New(t)
	tc := &testClient{}
	config := NewConfig()
	client := NewHTTPClient(context.TODO(), config, tc)
	assert.NotNil(client)
	tc.resp = &http.Response{
		StatusCode: http.StatusOK,
		Body:       &testReader{},
	}
	resp, err := client.Post("/test", "application/json", strings.NewReader("{}"))
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", tc.req.Header.Get("Content-Type"))
	assert.Equal(http.MethodPost, tc.req.Method)
	buf, err := ioutil.ReadAll(tc.req.Body)
	assert.NoError(err)
	assert.Equal("{}", string(buf))
}

func TestNewHTTPClientContext(t *testing.T) {
	assert := assert.New(t)
	tc := &testClient{}
	config := NewConfig()
	ctx := context.WithValue(context.Background(), ErrRequestTimeout, ErrInvalidClientImpl)
	client := NewHTTPClient(ctx, config, tc)
	assert.NotNil(client)
	tc.resp = &http.Response{
		StatusCode: http.StatusOK,
		Body:       &testReader{},
	}
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	assert.NotEqual(ctx, req.Context())
	assert.NoError(err)
	resp, err := client.Do(req)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal(ctx, tc.req.Context())
}

func TestNewHTTPClientNoRetry(t *testing.T) {
	assert := assert.New(t)
	tc := &testClient{}
	config := NewConfig()
	client := NewHTTPClient(context.TODO(), config, tc)
	assert.NotNil(client)
	tc.resp = &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       &testReader{},
	}
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	assert.NoError(err)
	resp, err := client.Do(req)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusNotFound, resp.StatusCode)
}

func TestNewHTTPClientRetryError(t *testing.T) {
	assert := assert.New(t)
	tc := &testClient{}
	tc.err = fmt.Errorf("error")
	config := NewConfig()
	config.Retryable = &retry{
		retryError: false,
	}
	client := NewHTTPClient(context.TODO(), config, tc)
	assert.NotNil(client)
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	assert.NoError(err)
	resp, err := client.Do(req)
	assert.EqualError(err, "error")
	assert.Nil(resp)
}

func TestNewHTTPClientRetryResponseError(t *testing.T) {
	assert := assert.New(t)
	tc := &testClient{}
	tc.resp = &http.Response{
		StatusCode: http.StatusInternalServerError,
	}
	config := NewConfig()
	config.Retryable = &retry{}
	client := NewHTTPClient(context.TODO(), config, tc)
	assert.NotNil(client)
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	assert.NoError(err)
	resp, err := client.Do(req)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusInternalServerError, resp.StatusCode)
}

type mockBody struct {
	read   bool
	closed bool
}

func (m *mockBody) Read(buf []byte) (int, error) {
	m.read = true
	return len(buf), io.EOF
}
func (m *mockBody) Close() error {
	m.closed = true
	return nil
}

func TestNewHTTPClientRetry(t *testing.T) {
	assert := assert.New(t)
	tc := &testClientRetry{
		resps: []*http.Response{
			&http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       &mockBody{},
			},
			&http.Response{
				StatusCode: http.StatusGatewayTimeout,
				Body:       &mockBody{},
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Body: &testReader{
					buf: bytes.NewBuffer([]byte("hi")),
				},
			},
		},
	}
	config := NewConfig()
	config.Retryable = NewBackoffRetry(time.Millisecond, 10*time.Millisecond, time.Second, 2)
	client := NewHTTPClient(context.TODO(), config, tc)
	assert.NotNil(client)
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	assert.NoError(err)
	resp, err := client.Do(req)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	buf, err := ioutil.ReadAll(resp.Body)
	assert.NoError(err)
	assert.Equal("hi", string(buf))
	assert.True(tc.resps[0].Body.(*mockBody).read)
	assert.True(tc.resps[0].Body.(*mockBody).closed)
	assert.True(tc.resps[1].Body.(*mockBody).read)
	assert.True(tc.resps[1].Body.(*mockBody).closed)
}

func TestNewHTTPClientRetryCancelled(t *testing.T) {
	assert := assert.New(t)
	tc := &testClientRetry{
		resps: []*http.Response{
			&http.Response{
				StatusCode: http.StatusInternalServerError,
			},
			&http.Response{
				StatusCode: http.StatusGatewayTimeout,
			},
		},
	}
	config := NewConfig()
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	cancel()
	config.Retryable = NewBackoffRetry(time.Millisecond, time.Second, 10*time.Second, 4)
	client := NewHTTPClient(ctx, config, tc)
	assert.NotNil(client)
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	assert.NoError(err)
	resp, err := client.Do(req)
	assert.EqualError(err, context.Canceled.Error())
	assert.Nil(resp)
}

func TestNewHTTPClientRetryTimeout(t *testing.T) {
	assert := assert.New(t)
	tc := &testClientRetry{
		resps: []*http.Response{
			&http.Response{
				StatusCode: http.StatusInternalServerError,
			},
			&http.Response{
				StatusCode: http.StatusGatewayTimeout,
			},
			&http.Response{
				StatusCode: http.StatusGatewayTimeout,
			},
			&http.Response{
				StatusCode: http.StatusGatewayTimeout,
			},
		},
	}
	config := NewConfig()
	config.Retryable = NewBackoffRetry(time.Millisecond, time.Millisecond*500, time.Second, 4)
	client := NewHTTPClient(context.Background(), config, tc)
	assert.NotNil(client)
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	assert.NoError(err)
	resp, err := client.Do(req)
	assert.EqualError(err, ErrRequestTimeout.Error())
	assert.Nil(resp)
}

func TestNewHTTPClientPagination(t *testing.T) {
	assert := assert.New(t)
	tc := &testClient{}
	var paged bool
	p := &paginator{
		paginate: func(page int, req *http.Request, resp *http.Response) (bool, *http.Request) {
			paged = true
			if page > 1 {
				return false, nil
			}
			return true, &http.Request{}
		},
	}
	config := NewConfig()
	config.Paginator = p
	client := NewHTTPClient(context.TODO(), config, tc)
	assert.NotNil(client)
	tc.resp = &http.Response{
		StatusCode: http.StatusOK,
		Body:       &testReader{},
	}
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	assert.NoError(err)
	resp, err := client.Do(req)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.True(paged)
}

func TestNewHTTPClientPaginationNone(t *testing.T) {
	assert := assert.New(t)
	tc := &testClient{}
	var paged bool
	p := &paginator{
		paginate: func(page int, req *http.Request, resp *http.Response) (bool, *http.Request) {
			paged = true
			if page > 1 {
				return false, nil
			}
			return true, &http.Request{}
		},
	}
	config := NewConfig()
	config.Paginator = p
	client := NewHTTPClient(context.TODO(), config, tc)
	assert.NotNil(client)
	tc.resp = &http.Response{
		StatusCode: http.StatusOK,
		Body:       &testReader{},
	}
	req, err := http.NewRequest(http.MethodPost, "/test", nil)
	assert.NoError(err)
	resp, err := client.Do(req)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.False(paged)
}
