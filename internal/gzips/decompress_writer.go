package gzips

import (
	"compress/gzip"
	"io"
)

func NewDecompressWriter(dest io.Writer) io.WriteCloser {
	r, w := io.Pipe()
	go func() {
		defer func(r *io.PipeReader) {
			_ = r.Close()
		}(r)

		gzr, err := gzip.NewReader(r)
		if err != nil {
			_ = r.CloseWithError(err)
			return
		}

		_, err = io.Copy(dest, gzr)
		if err != nil {
			_ = r.CloseWithError(err)
			return
		}

		err = gzr.Close()
		if err != nil {
			_ = r.CloseWithError(err)
			return
		}
	}()
	return w
}
