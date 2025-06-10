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

func ReadAndYeet(cr *CountingReader, read func() uint32) {
	startPos := cr.Position()
	objectSize := read()
	endPos := cr.Position()

	bytesRead := uint32(endPos - startPos)
	if bytesRead > objectSize {
		panic("Read more bytes than expected")
	}

	trailingBytes := objectSize - bytesRead
	if trailingBytes != 0 {
		if _, err := io.CopyN(io.Discard, cr, int64(trailingBytes)); err != nil {
			panic(err)
		}
	}
}
