package httpclient

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginator(t *testing.T) {
	assert := assert.New(t)
	p := &noPaginator{}
	ok, r := p.HasMore(1, &http.Response{})
	assert.False(ok)
	assert.Nil(r)
}
