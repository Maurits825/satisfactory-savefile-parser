package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

type SaveFileHeader struct {
	SaveHeaderVersion   uint32
	SaveVersion         uint32
	BuildVersion        uint32
	SessionName         string
	MapName             string
	MapOptions          string
	SaveName            string
	PlayedSeconds       uint32
	SaveTimestampTicks  uint64
	SessionVisibility   byte
	EditorObjectVersion uint32
	ModMetadata         string
	ModFlags            uint32
	SaveIdentifier      string
	Unknown1            uint32 // always 1
	Unknown2            uint32 // always 1
	SessionRandom1      uint64
	SessionRandom2      uint64
	CheatFlag           uint32
}

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

func readHeader(file io.Reader) *SaveFileHeader {
	header := &SaveFileHeader{}

	readFields(file, &header.SaveHeaderVersion, &header.SaveVersion, &header.BuildVersion)
	readFields(file,
		conditionalFields(header.SaveHeaderVersion >= 14, &header.SaveName),
		&header.MapName, &header.MapOptions, &header.SessionName,
		&header.PlayedSeconds, &header.SaveTimestampTicks, &header.SessionVisibility, &header.EditorObjectVersion,
		&header.ModMetadata, &header.ModFlags, &header.SaveIdentifier,
		conditionalFields(header.SaveHeaderVersion >= 13, &header.Unknown1, &header.Unknown2, &header.SessionRandom1, &header.SessionRandom2, &header.CheatFlag),
	)

	return header
}

func readCompressedSaveFileBody(file io.Reader) (io.ReadCloser, uint64, error) {
	var compressedBody CompressedSaveFileBody
	if err := binary.Read(file, binary.LittleEndian, &compressedBody); err != nil {
		return nil, 0, err
	}
	if !compressedBody.isValid() {
		return nil, 0, fmt.Errorf("invalid compressed save file body header: %+v", compressedBody)
	}

	compressedBytes := make([]byte, compressedBody.CompressedSize1)
	if _, err := io.ReadFull(file, compressedBytes); err != nil {
		return nil, 0, fmt.Errorf("reading compressed body: %w", err)
	}

	b := bytes.NewReader(compressedBytes)
	zr, err := zlib.NewReader(b)
	if err != nil {
		return nil, 0, fmt.Errorf("creating zlib reader: %w", err)
	}
	return zr, compressedBody.UncompressedSize1, nil
}

func readSaveFile(saveFile string) *SaveFileBody {
	file, err := os.Open(saveFile)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	defer file.Close()

	header := readHeader(file)
	fmt.Printf("Save name: %s, Version: %d-%d-%d\n", header.SessionName, header.SaveVersion, header.SaveHeaderVersion, header.BuildVersion)
	fmt.Println("")

	readers := make([]io.Reader, 0)
	totalSize := uint64(0)
	fmt.Println("Decompressing save file body")
	for {
		zr, size, err := readCompressedSaveFileBody(file)
		totalSize += size
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error reading compressed save file body:", err)
			return nil
		}
		readers = append(readers, zr)
	}

	multiZr := io.MultiReader(readers...)
	cr := &countingReader{r: multiZr}

	startTime := time.Now()
	statusUpdate := newStatusTicker(1*time.Second, func() { statusPrint(cr, totalSize, startTime) })
	statusUpdate.start()

	body, err := readSaveFileBody(cr, header.SaveVersion)
	if err != nil {
		fmt.Println("reading save file body: %w", err)
		return nil
	}

	statusUpdate.stop()
	tDiff := float64(time.Since(startTime).Seconds())
	fmt.Printf("Done reading in %.2fs\n", tDiff)
	fmt.Println("")
	return body
}

func statusPrint(cr *countingReader, total uint64, startTime time.Time) {
	pos := cr.Position()
	percent := float64(pos) / float64(total) * 100.0

	tDiff := float64(time.Since(startTime).Seconds())
	percentRate := float64(percent) / tDiff
	eta := (100.0 - percent) / percentRate
	fmt.Printf("Progress %.2f%% ETA: %.2fs\n", percent, eta)
}
