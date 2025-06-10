package readsave

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Maurits825/satisfactory-savefile-parser/internal/countingreader"
	"github.com/Maurits825/satisfactory-savefile-parser/internal/readsave/readfields"
	"github.com/Maurits825/satisfactory-savefile-parser/pkg/saveformat"
)

func ReadSave(file *os.File) *saveformat.SaveFileBody {
	header := readHeader(file)
	fmt.Printf("Save name: %s, Version: %d-%d-%d\n", header.SessionName, header.SaveVersion, header.SaveHeaderVersion, header.BuildVersion)
	fmt.Println("")

	readers := make([]io.Reader, 0)
	totalSize := uint64(0)
	fmt.Println("Decompressing save file body")
	for {
		zr, size := readCompressedSaveFileBody(file)
		if size == 0 {
			break
		}
		totalSize += size
		readers = append(readers, zr)
	}

	multiZr := io.MultiReader(readers...)
	cr := countingreader.NewCountingReader(multiZr)

	startTime := time.Now()
	statusUpdate := newStatusTicker(1*time.Second, func() { statusPrint(cr, totalSize, startTime) })
	statusUpdate.start()

	body := readSaveFileBody(cr, header.SaveVersion)

	statusUpdate.stop()
	tDiff := float64(time.Since(startTime).Seconds())
	fmt.Printf("Done reading in %.2fs\n", tDiff)
	fmt.Println("")
	return body
}

func readHeader(file io.Reader) *saveformat.SaveFileHeader {
	header := &saveformat.SaveFileHeader{}

	readfields.ReadFields(file, &header.SaveHeaderVersion, &header.SaveVersion, &header.BuildVersion)
	readfields.ReadFields(file,
		readfields.ConditionalFields(header.SaveHeaderVersion >= 14, &header.SaveName),
		&header.MapName, &header.MapOptions, &header.SessionName,
		&header.PlayedSeconds, &header.SaveTimestampTicks, &header.SessionVisibility, &header.EditorObjectVersion,
		&header.ModMetadata, &header.ModFlags, &header.SaveIdentifier,
		readfields.ConditionalFields(header.SaveHeaderVersion >= 13, &header.Unknown1, &header.Unknown2, &header.SessionRandom1, &header.SessionRandom2, &header.CheatFlag),
	)

	return header
}

func statusPrint(cr *countingreader.CountingReader, total uint64, startTime time.Time) {
	pos := cr.Position()
	percent := float64(pos) / float64(total) * 100.0

	tDiff := float64(time.Since(startTime).Seconds())
	percentRate := float64(percent) / tDiff
	eta := (100.0 - percent) / percentRate
	fmt.Printf("Progress %.2f%% ETA: %.2fs\n", percent, eta)
}
