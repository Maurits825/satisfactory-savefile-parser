package readsave

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
)

type CompressedSaveFileBody struct {
	Magic             uint32
	Hex2s             uint32
	Zero              byte
	MaxChunkSize      uint32
	Hex03             uint32
	CompressedSize1   uint64
	UncompressedSize1 uint64
	CompressedSize2   uint64
	UncompressedSize2 uint64
}

func (body *CompressedSaveFileBody) isValid() bool {
	return body.Magic == 0x9E2A83C1 &&
		body.Hex2s == 0x22222222 &&
		body.Zero == 0 &&
		body.MaxChunkSize == 512 && //TODO wiki says 131,072??
		body.Hex03 == 0x03000000 &&
		body.CompressedSize1 == body.CompressedSize2 &&
		body.UncompressedSize1 == body.UncompressedSize2
}

func readCompressedSaveFileBody(file io.Reader) (io.ReadCloser, uint64) {
	var compressedBody CompressedSaveFileBody
	if err := binary.Read(file, binary.LittleEndian, &compressedBody); err != nil {
		if err == io.EOF {
			return nil, 0
		}
		panic(err)
	}
	if !compressedBody.isValid() {
		panic("invalid compressed save file body")
	}

	compressedBytes := make([]byte, compressedBody.CompressedSize1)
	if _, err := io.ReadFull(file, compressedBytes); err != nil {
		panic("reading compressed body: " + err.Error())
	}

	b := bytes.NewReader(compressedBytes)
	zr, err := zlib.NewReader(b)
	if err != nil {
		panic("creating zlib reader: " + err.Error())
	}

	return zr, compressedBody.UncompressedSize1
}
