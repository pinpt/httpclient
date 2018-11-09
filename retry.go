package httpclient

import (
	"math"
	"net/http"
	"strings"
	"time"
)

type noRetry struct {
}

var _ Retryable = (*noRetry)(nil)

func (r *noRetry) RetryError(err error) bool {
	return false
}

func (r *noRetry) RetryResponse(resp *http.Response) bool {
	return false
}

func (r *noRetry) RetryDelay(retry int) time.Duration {
	return time.Duration(0)
}

func (r *noRetry) RetryMaxDuration() time.Duration {
	return time.Second
}

// NewNoRetry will return a struct that implements Retryable but doesn't retry at all
func NewNoRetry() Retryable {
	return &noRetry{}
}

type backoffRetry struct {
	exponentFactor      float64
	initialTimeout      float64
	incrementingTimeout float64
	maxTimeout          float64
}

var _ Retryable = (*backoffRetry)(nil)

func (r *backoffRetry) RetryError(err error) bool {
	return true
}

func (r *backoffRetry) RetryResponse(resp *http.Response) bool {
	return true
}

func (r *backoffRetry) RetryDelay(retry int) time.Duration {
	return time.Duration((r.initialTimeout + math.Pow(r.exponentFactor, float64(retry))) * r.incrementingTimeout)
}

func (r *backoffRetry) RetryMaxDuration() time.Duration {
	return time.Duration(r.maxTimeout)
}

// NewBackoffRetry will return a Retryable that will support expotential backoff
func NewBackoffRetry(initialTimeout time.Duration, incrementingTimeout time.Duration, maxTimeout time.Duration, exponentFactor float64) Retryable {
	return &backoffRetry{
		exponentFactor:      exponentFactor,
		initialTimeout:      float64(initialTimeout / time.Millisecond),
		incrementingTimeout: float64(incrementingTimeout),
		maxTimeout:          float64(maxTimeout),
	}
}

type JiraRetry struct {
	backoffRetry Retryable
}

func (r *JiraRetry) RetryError(err error) bool {
	if strings.Contains(err.Error(), "This Jira instance is currently under heavy load") {
		return true
	}
	return r.backoffRetry.RetryError(err)
}

func (r *JiraRetry) RetryResponse(resp *http.Response) bool {
	return r.backoffRetry.RetryResponse(resp)
}

func (r *JiraRetry) RetryDelay(retry int) time.Duration {
	return r.backoffRetry.RetryDelay(retry)

}

func (r *JiraRetry) RetryMaxDuration() time.Duration {
	return r.backoffRetry.RetryMaxDuration()
}

func NewJiraRetry() *JiraRetry {
	return &JiraRetry{
		NewBackoffRetry(10*time.Millisecond, 100*time.Millisecond, 3*time.Minute, 2.0),
	}
}
