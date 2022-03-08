package gzips

import (
	"compress/gzip"
	"io"
)

func NewDecompressWriter(dest io.Writer, ch chan error) io.WriteCloser {
	r, w := io.Pipe()
	go func() {
		gzr, err := gzip.NewReader(r)
		if err != nil {
			_ = r.CloseWithError(err)
			return
		}
		defer func(gzr *gzip.Reader) {
			_ = gzr.Close()
		}(gzr)

		_, err = io.Copy(dest, gzr)
		ch <- err
	}()
	return w
}
