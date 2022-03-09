package gzips

import (
	"compress/gzip"
	"io"
)

func NewCompressReader(source io.Reader) io.Reader {
	r, w := io.Pipe()
	go func() {
		defer func(w *io.PipeWriter) {
			_ = w.Close()
		}(w)

		gzw := gzip.NewWriter(w)
		defer func(gzw *gzip.Writer) {
			_ = gzw.Close()
		}(gzw)

		_, err := io.Copy(gzw, source)
		if err != nil {
			_ = w.CloseWithError(err)
			return
		}
	}()
	return r
}
