package countingreader

import "io"

type CountingReader struct {
	r     io.Reader
	total int64
}

func NewCountingReader(r io.Reader) *CountingReader {
	return &CountingReader{r: r}
}

func (cr *CountingReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	cr.total += int64(n)
	return n, err
}

func (cr *CountingReader) Position() int64 {
	return cr.total
}
