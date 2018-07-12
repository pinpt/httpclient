package httpclient

import (
	"bytes"
	"io"
	"io/ioutil"
)

// borrowed from https://golang.org/src/io/multi.go

type multiReader struct {
	streams []io.Reader
}

func newMuliReader() *multiReader {
	return &multiReader{}
}

func (r *multiReader) Add(rc io.ReadCloser) error {
	if r.streams == nil {
		r.streams = make([]io.Reader, 0)
	}
	// NOTE: we read all in memory which is terrible _but_ with load testing
	// under windows, we get weird "wsasend: An existing connection was forcibly closed by the remote host."
	// messages by keeping multiple connections open (>300)
	buf, err := ioutil.ReadAll(rc)
	if err != nil {
		rc.Close()
		return err
	}
	rc.Close()
	r.streams = append(r.streams, bytes.NewReader(buf))
	return nil
}

func (r *multiReader) Close() error {
	r.streams = nil
	return nil
}

func (r *multiReader) Read(p []byte) (n int, err error) {
	for r.streams != nil && len(r.streams) > 0 {
		// Optimization to flatten nested multiReaders (Issue 13558).
		if len(r.streams) == 1 {
			if mr, ok := r.streams[0].(*multiReader); ok {
				r.streams = mr.streams
				continue
			}
		}
		n, err = r.streams[0].Read(p)
		if err == io.EOF {
			// Use eofReader instead of nil to avoid nil panic
			// after performing flatten (Issue 18232).
			r.streams[0] = eofReader{} // permit earlier GC
			r.streams = r.streams[1:]
		}
		if n > 0 || err != io.EOF {
			if err == io.EOF && len(r.streams) > 0 {
				// Don't return EOF yet. More readers remain.
				err = nil
			}
			return
		}
	}
	return 0, io.EOF
}

type eofReader struct{}

func (eofReader) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (eofReader) Close() error {
	return nil
}
