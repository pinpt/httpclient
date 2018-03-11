package httpclient

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNoRetry(t *testing.T) {
	assert := assert.New(t)
	retry := &noRetry{}
	assert.False(retry.RetryError(ErrRequestTimeout))
	assert.False(retry.RetryResponse(&http.Response{}))
	assert.Equal(time.Duration(0), retry.RetryDelay(1))
	assert.Equal(time.Duration(0), retry.RetryDelay(2))
	assert.Equal(time.Second, retry.RetryMaxDuration())
}

func TestNoRetryNew(t *testing.T) {
	assert := assert.New(t)
	retry := NewNoRetry()
	assert.False(retry.RetryError(ErrRequestTimeout))
	assert.False(retry.RetryResponse(&http.Response{}))
	assert.Equal(time.Duration(0), retry.RetryDelay(1))
	assert.Equal(time.Duration(0), retry.RetryDelay(2))
	assert.Equal(time.Second, retry.RetryMaxDuration())
}

func TestBackoffRetry(t *testing.T) {
	assert := assert.New(t)
	retry := NewBackoffRetry(time.Millisecond, 10*time.Millisecond, time.Second, 1.5)
	assert.True(retry.RetryError(ErrRequestTimeout))
	assert.True(retry.RetryResponse(&http.Response{}))
	assert.Equal(time.Duration(25000000), retry.RetryDelay(1))
	assert.Equal(time.Duration(32500000), retry.RetryDelay(2))
	assert.Equal(time.Duration(43750000), retry.RetryDelay(3))
	assert.Equal(time.Duration(60625000), retry.RetryDelay(4))
	assert.Equal(time.Second, retry.RetryMaxDuration())
}
