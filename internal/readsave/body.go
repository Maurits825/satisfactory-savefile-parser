package readsave

import (
	"errors"
	"fmt"
	"io"

	"github.com/Maurits825/satisfactory-savefile-parser/internal/countingreader"
	. "github.com/Maurits825/satisfactory-savefile-parser/internal/readsave/readfields"
	"github.com/Maurits825/satisfactory-savefile-parser/pkg/saveformat"
)

func readSaveFileBody(zr io.Reader, version uint32) (*saveformat.SaveFileBody, error) {
	var body saveformat.SaveFileBody

	ReadFields(zr,
		&body.UncompressedSize, &body.Value6, &body.NoneString1, &body.Value0,
		&body.Unknown1, &body.Value1, &body.NoneString2, &body.Unknown2,
	)

	if !body.IsValid() {
		return nil, errors.New("invalid save file body header")
	}

	for range 5 {
		grid := readLevelGroupingGrid(zr)
		body.LevelGroupingGrids = append(body.LevelGroupingGrids, *grid)
	}

	ReadFields(zr, &body.SubLevelCount)

	for range body.SubLevelCount {
		levelData := readLevelData(zr, version, false)
		body.Levels = append(body.Levels, *levelData)
	}

	fmt.Println("Reading persistent level data ...")
	levelData := readLevelData(zr, version, true)
	body.Levels = append(body.Levels, *levelData)

	//TODO zero field present here?

	readReferenceList(zr, &body)

	leftBytes, err := io.ReadAll(zr)
	if err != nil {
		return nil, err
	}
	if len(leftBytes) != 0 {
		fmt.Printf("Warning - Left bytes after reading body: %d\n", len(leftBytes))
	}

	return &body, err
}

func readLevelGroupingGrid(r io.Reader) *saveformat.LevelGroupingGrid {
	var grid saveformat.LevelGroupingGrid

	ReadFields(r,
		&grid.GridName, &grid.Unknown1, &grid.Unknown2, &grid.LevelCount,
	)

	for range grid.LevelCount {
		var levelInfo saveformat.LevelInfo
		ReadFields(r, &levelInfo.StringValue, &levelInfo.IntValue)
		grid.LevelInfos = append(grid.LevelInfos, levelInfo)
	}

	return &grid
}

// todo split into read header/object funcs?
func readLevelData(r io.Reader, version uint32, isPersistentLevel bool) *saveformat.LevelData {
	var levelData saveformat.LevelData

	if !isPersistentLevel {
		ReadFields(r, &levelData.Name)
	}

	ReadFields(r, &levelData.Size, &levelData.HeaderCount)

	headerTypes := make([]uint32, 0, levelData.HeaderCount)
	startPos := r.(*countingreader.CountingReader).Position()
	for range levelData.HeaderCount {
		var headerType uint32
		ReadFields(r, &headerType)
		headerTypes = append(headerTypes, headerType)
		if headerType == 0 {
			var componentHeader saveformat.ComponentHeader
			ReadFields(r,
				&componentHeader.TypePath, &componentHeader.Root,
				&componentHeader.Name, ConditionalFields(version >= 51, &componentHeader.Flags),
				&componentHeader.ParentActorName,
			)
			levelData.ComponentHeaders = append(levelData.ComponentHeaders, componentHeader)

		} else if headerType == 1 {
			var actorHeader saveformat.ActorHeader
			ReadFields(r,
				&actorHeader.TypePath, &actorHeader.Root, &actorHeader.Name,
				ConditionalFields(version >= 51, &actorHeader.Flags), &actorHeader.NeedTransform,
				&actorHeader.RotationX, &actorHeader.RotationY, &actorHeader.RotationZ, &actorHeader.RotationW,
				&actorHeader.PositionX, &actorHeader.PositionY, &actorHeader.PositionZ,
				&actorHeader.ScaleX, &actorHeader.ScaleY, &actorHeader.ScaleZ,
				&actorHeader.WasPlaced,
			)
			levelData.ActorHeaders = append(levelData.ActorHeaders, actorHeader)

		} else {
			panic("Unknown header type: " + fmt.Sprint(headerType))
		}
	}

	endPos := r.(*countingreader.CountingReader).Position()
	bytesRead := uint64(endPos - startPos)
	diff := levelData.Size - bytesRead
	if diff > 4 {
		ReadFields(r, &levelData.CollectableCount)
		if levelData.CollectableCount > 0 && isPersistentLevel {
			var name string
			ReadFields(r, &name, &levelData.CollectableCount)
		}
		for range levelData.CollectableCount {
			var collectable saveformat.ObjectReference
			ReadFields(r, &collectable.LevelName, &collectable.PathName)
		}
	}

	ReadFields(r, &levelData.ObjectSize, &levelData.ObjectCount)

	//go through objects
	for i := range levelData.ObjectCount {
		startPos := r.(*countingreader.CountingReader).Position()
		var objectSize uint32
		if headerTypes[i] == 0 {
			var component saveformat.ComponentObject
			ReadFields(r, &component.SaveVersion, &component.Flag, &component.Size)
			objectSize = component.Size
			ReadAllProperties(r, &component.Properties)

			//Zero field not always present, part of trailing bytes?

			if !component.IsValid() {
				panic("Invalid component object at index" + fmt.Sprint(i))
			}
			levelData.ComponentObjects = append(levelData.ComponentObjects, component)
		} else if headerTypes[i] == 1 {
			var actor saveformat.ActorObject
			ReadFields(r, &actor.SaveVersion, &actor.Flag, &actor.Size,
				&actor.ParentReference, &actor.ComponentCount)
			objectSize = actor.Size
			for range actor.ComponentCount {
				var component saveformat.ObjectReference
				ReadFields(r, &component.LevelName, &component.PathName)
				actor.Components = append(actor.Components, component)
			}
			ReadAllProperties(r, &actor.Properties)
			if !actor.IsValid() {
				panic("Invalid actor object at index" + fmt.Sprint(i))
			}
			levelData.ActorObjects = append(levelData.ActorObjects, actor)
		}
		endPos := r.(*countingreader.CountingReader).Position()
		bytesRead := uint32(endPos - startPos)
		objectSize += 12 // 12 bytes for SaveVersion, Flag, Size
		if bytesRead < objectSize {
			trailingBytes := objectSize - bytesRead
			if trailingBytes != 0 {
				if _, err := io.CopyN(io.Discard, r, int64(trailingBytes)); err != nil {
					panic(err)
				}
			}
		} else if bytesRead > objectSize {
			panic("Read more bytes than expected object at index" + fmt.Sprint(i))
		}
	}

	if !isPersistentLevel && version >= 51 {
		var v uint32
		ReadFields(r, &v)
	}

	ReadFields(r, &levelData.SecondCollectableCount)
	for range levelData.SecondCollectableCount {
		var collectable saveformat.ObjectReference
		ReadFields(r, &collectable.LevelName, &collectable.PathName)
	}

	return &levelData
}

// TODO return ref list here? do we need this?
func readReferenceList(zr io.Reader, body *saveformat.SaveFileBody) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error reading ref list, skipping: ", r)
		}
	}()

	ReadFields(zr, &body.ReferenceListCount)
	for range body.ReferenceListCount {
		var reference saveformat.ObjectReference
		ReadFields(zr, &reference.LevelName, &reference.PathName)
		body.References = append(body.References, reference)
	}
}
