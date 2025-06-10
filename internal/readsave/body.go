package readsave

import (
	"fmt"
	"io"

	"github.com/Maurits825/satisfactory-savefile-parser/internal/countingreader"
	"github.com/Maurits825/satisfactory-savefile-parser/internal/readsave/readfields"
	"github.com/Maurits825/satisfactory-savefile-parser/pkg/saveformat"
)

func readSaveFileBody(cr *countingreader.CountingReader, version uint32) *saveformat.SaveFileBody {
	var body saveformat.SaveFileBody

	readfields.ReadFields(cr,
		&body.UncompressedSize, &body.Value6, &body.NoneString1, &body.Value0,
		&body.Unknown1, &body.Value1, &body.NoneString2, &body.Unknown2,
	)

	if !body.IsValid() {
		panic("Invalid save file body")
	}

	for range 5 {
		grid := readLevelGroupingGrid(cr)
		body.LevelGroupingGrids = append(body.LevelGroupingGrids, *grid)
	}

	readfields.ReadFields(cr, &body.SubLevelCount)

	for range body.SubLevelCount {
		levelData := readLevelData(cr, version, false)
		body.Levels = append(body.Levels, *levelData)
	}

	fmt.Println("Reading persistent level data ...")
	levelData := readLevelData(cr, version, true)
	body.Levels = append(body.Levels, *levelData)

	//TODO zero field present here?

	readReferenceList(cr, &body)

	leftBytes, err := io.ReadAll(cr)
	if err != nil {
		panic("Reading left bytes after reading body: " + err.Error())
	}
	if len(leftBytes) != 0 {
		fmt.Printf("Warning - Left bytes after reading body: %d\n", len(leftBytes))
	}

	return &body
}

func readLevelGroupingGrid(r io.Reader) *saveformat.LevelGroupingGrid {
	var grid saveformat.LevelGroupingGrid

	readfields.ReadFields(r,
		&grid.GridName, &grid.Unknown1, &grid.Unknown2, &grid.LevelCount,
	)

	for range grid.LevelCount {
		var levelInfo saveformat.LevelInfo
		readfields.ReadFields(r, &levelInfo.StringValue, &levelInfo.IntValue)
		grid.LevelInfos = append(grid.LevelInfos, levelInfo)
	}

	return &grid
}

// TODO return ref list here? do we need this?
func readReferenceList(zr io.Reader, body *saveformat.SaveFileBody) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error reading ref list, skipping: ", r)
		}
	}()

	readfields.ReadFields(zr, &body.ReferenceListCount)
	for range body.ReferenceListCount {
		var reference saveformat.ObjectReference
		readfields.ReadFields(zr, &reference.LevelName, &reference.PathName)
		body.References = append(body.References, reference)
	}
}
