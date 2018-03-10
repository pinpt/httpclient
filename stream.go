package httpclient

import "io"

// borrowed from https://golang.org/src/io/multi.go

type multiReader struct {
	streams []io.ReadCloser
}

func newMuliReader() *multiReader {
	return &multiReader{}
}

func (r *multiReader) Add(rc io.ReadCloser) {
	if r.streams == nil {
		r.streams = make([]io.ReadCloser, 0)
	}
	r.streams = append(r.streams, rc)
}

func (r *multiReader) Close() error {
	if r.streams != nil {
		for _, s := range r.streams {
			if err := s.Close(); err != nil {
				return err
			}
		}
		r.streams = nil
	}
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
