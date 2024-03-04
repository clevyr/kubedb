package util

import "io"

func ReadAndClose(r io.ReadCloser) {
	_, _ = io.Copy(io.Discard, r)
	_ = r.Close()
}
