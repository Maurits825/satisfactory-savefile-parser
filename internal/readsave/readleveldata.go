package readsave

import (
	"fmt"
	"io"

	"github.com/Maurits825/satisfactory-savefile-parser/internal/countingreader"
	. "github.com/Maurits825/satisfactory-savefile-parser/internal/readsave/readfields"
	"github.com/Maurits825/satisfactory-savefile-parser/pkg/saveformat"
)

func readLevelData(cr *countingreader.CountingReader, version uint32, isPersistentLevel bool) *saveformat.LevelData {
	var levelData saveformat.LevelData

	if !isPersistentLevel {
		ReadFields(cr, &levelData.Name)
	}

	ReadFields(cr, &levelData.Size, &levelData.HeaderCount)

	startPos := cr.Position()
	headerTypes := readLevelHeader(cr, &levelData, version)
	endPos := cr.Position()

	bytesRead := uint64(endPos - startPos)
	diff := levelData.Size - bytesRead
	if diff > 4 {
		ReadFields(cr, &levelData.CollectableCount)
		if levelData.CollectableCount > 0 && isPersistentLevel {
			var name string
			ReadFields(cr, &name, &levelData.CollectableCount)
		}
		for range levelData.CollectableCount {
			var collectable saveformat.ObjectReference
			ReadFields(cr, &collectable.LevelName, &collectable.PathName)
		}
	}

	ReadFields(cr, &levelData.ObjectSize, &levelData.ObjectCount)

	for i := range levelData.ObjectCount {
		read := func() uint32 {
			objectSize := readLevelObject(cr, &levelData, headerTypes[i])
			objectSize += 12 // 12 bytes for SaveVersion, Flag, Size
			return objectSize
		}
		countingreader.ReadAndYeet(cr, read)
	}

	if !isPersistentLevel && version >= 51 {
		var v uint32
		ReadFields(cr, &v)
	}

	ReadFields(cr, &levelData.SecondCollectableCount)
	for range levelData.SecondCollectableCount {
		var collectable saveformat.ObjectReference
		ReadFields(cr, &collectable.LevelName, &collectable.PathName)
	}

	return &levelData
}

func readLevelHeader(r io.Reader, levelData *saveformat.LevelData, version uint32) []uint32 {
	headerTypes := make([]uint32, 0, levelData.HeaderCount)

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

	return headerTypes
}

func readLevelObject(cr *countingreader.CountingReader, levelData *saveformat.LevelData, headerType uint32) uint32 {
	var objectSize uint32
	if headerType == 0 {
		var component saveformat.ComponentObject
		ReadFields(cr, &component.SaveVersion, &component.Flag, &component.Size)
		objectSize = component.Size
		ReadAllProperties(cr, &component.Properties)

		if !component.IsValid() {
			panic("Invalid component object")
		}

		levelData.ComponentObjects = append(levelData.ComponentObjects, component)
	} else if headerType == 1 {
		var actor saveformat.ActorObject
		ReadFields(cr, &actor.SaveVersion, &actor.Flag, &actor.Size,
			&actor.ParentReference, &actor.ComponentCount)
		objectSize = actor.Size
		for range actor.ComponentCount {
			var component saveformat.ObjectReference
			ReadFields(cr, &component.LevelName, &component.PathName)
			actor.Components = append(actor.Components, component)
		}
		ReadAllProperties(cr, &actor.Properties)

		if !actor.IsValid() {
			panic("Invalid actor object")
		}

		levelData.ActorObjects = append(levelData.ActorObjects, actor)
	}

	return objectSize
}
