package util

import "github.com/dustin/go-humanize"

type SizeWriter struct {
	size uint64
}

func (s *SizeWriter) Write(p []byte) (int, error) {
	s.size += uint64(len(p))
	return len(p), nil
}

func (s *SizeWriter) Size() uint64 {
	return s.size
}

func (s *SizeWriter) String() string {
	return humanize.IBytes(s.size)
}
