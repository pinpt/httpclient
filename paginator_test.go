package httpclient

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginator(t *testing.T) {
	assert := assert.New(t)
	p := &noPaginator{}
	ok, r := p.HasMore(1, &http.Request{}, &http.Response{})
	assert.False(ok)
	assert.Nil(r)
}

func TestPaginatorNew(t *testing.T) {
	assert := assert.New(t)
	p := NoPaginator()
	ok, r := p.HasMore(1, &http.Request{}, &http.Response{})
	assert.False(ok)
	assert.Nil(r)
}

func TestLinkPaginator(t *testing.T) {
	assert := assert.New(t)
	p := &linkPaginator{}
	u, _ := url.Parse("https://foo.com/bar?page=1")
	ok, r := p.HasMore(1, &http.Request{URL: u}, &http.Response{
		Header: http.Header{"Link": []string{`link <https://foo.com/bar?page=2>; rel="next"`}},
	})
	assert.True(ok)
	assert.NotNil(r)
	assert.Equal("https://foo.com/bar?page=2", r.URL.String())
}

func TestLinkPaginatorNone(t *testing.T) {
	assert := assert.New(t)
	p := NewLinkPaginator()
	u, _ := url.Parse("https://foo.com/bar?page=1")
	ok, r := p.HasMore(1, &http.Request{URL: u}, &http.Response{})
	assert.False(ok)
	assert.Nil(r)
}
