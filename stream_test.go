package httpclient

import (
	"bytes"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testReader struct {
	buf    *bytes.Buffer
	err    error
	closed bool
}

func (r *testReader) Read(p []byte) (int, error) {
	return r.buf.Read(p)
}

func (r *testReader) Close() error {
	r.closed = true
	return r.err
}

func TestMultiStream(t *testing.T) {
	assert := assert.New(t)
	stream := newMuliReader()
	assert.Nil(stream.streams)
	assert.NoError(stream.Close())
	assert.Nil(stream.streams)
	r := &testReader{
		buf: &bytes.Buffer{},
	}
	r.buf.WriteString("hi")
	stream.Add(r)
	assert.NotNil(stream.streams)
	buf, err := ioutil.ReadAll(stream)
	assert.NoError(err)
	assert.Equal("hi", string(buf))
	assert.NoError(stream.Close())
	assert.Nil(stream.streams)
	for _, s := range stream.streams {
		assert.True(s.(*testReader).closed)
	}
}

func TestMultiStreamCloseError(t *testing.T) {
	assert := assert.New(t)
	stream := newMuliReader()
	assert.Nil(stream.streams)
	assert.NoError(stream.Close())
	assert.Nil(stream.streams)
	r := &testReader{
		buf: &bytes.Buffer{},
	}
	r.buf.WriteString("hi")
	stream.Add(r)
	re := &testReader{}
	re.err = errors.New("error")
	stream.Add(re)
	assert.NotNil(stream.streams)
	assert.Len(stream.streams, 2)
	assert.EqualError(stream.Close(), "error")
}
