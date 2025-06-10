package main

import "io"

type countingReader struct {
	r     io.Reader
	total int64
}

func (cr *countingReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	cr.total += int64(n)
	return n, err
}

func (cr *countingReader) Position() int64 {
	return cr.total
}
