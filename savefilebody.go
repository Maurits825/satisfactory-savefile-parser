package main

import (
	"errors"
	"fmt"
	"io"
)

type SaveFileBody struct {
	UncompressedSize   uint64
	Value6             uint32
	NoneString1        string
	Value0             uint32
	Unknown1           uint32
	Value1             uint32
	NoneString2        string
	Unknown2           uint32
	LevelGroupingGrids []LevelGroupingGrid
	SubLevelCount      uint32
	Levels             []LevelData
	Zero               uint32
	ReferenceListCount uint32
	References         []ObjectReference
}

type LevelGroupingGrid struct {
	GridName   string
	Unknown1   uint32
	Unknown2   uint32
	LevelCount uint32
	LevelInfos []LevelInfo
}

type LevelInfo struct {
	StringValue string
	IntValue    uint32
}

type LevelData struct {
	Name                   string
	Size                   uint64
	HeaderCount            uint32
	ActorHeaders           []ActorHeader
	ComponentHeaders       []ComponentHeader
	CollectableCount       uint32
	Collectables           []ObjectReference
	ObjectSize             uint64
	ObjectCount            uint32
	ActorObjects           []ActorObject
	ComponentObjects       []ComponentObject
	SecondCollectableCount uint32
	SecondCollectables     []ObjectReference
}

func (l *LevelData) PrintSummary() {
	fmt.Printf("Level data name: %s, count: %d\n",
		l.Name, l.HeaderCount)

	if len(l.ActorHeaders) > 0 {
		fmt.Println("Actors: ", len(l.ActorHeaders))
		for i := range l.ActorHeaders {
			fmt.Printf("type: %s, size: %d, components: %d, properties: %d\n",
				l.ActorHeaders[i].TypePath, l.ActorObjects[i].Size, l.ActorObjects[i].ComponentCount, len(l.ActorObjects[i].Properties))
			if len(l.ActorObjects[i].Properties) > 0 {
				fmt.Printf("Properties: %+v\n", l.ActorObjects[i].Properties)
			}
		}
	}
	if len(l.ComponentHeaders) > 0 {
		fmt.Println("Components: ", len(l.ComponentHeaders))
		for i := range l.ComponentHeaders {
			fmt.Printf("type: %s, size: %d, properties: %d\n",
				l.ComponentHeaders[i].TypePath, l.ComponentObjects[i].Size, len(l.ComponentObjects[i].Properties))
			if len(l.ComponentObjects[i].Properties) > 0 {
				fmt.Printf("Properties: %+v\n", l.ComponentObjects[i].Properties)
			}
		}
	}
}

type ActorHeader struct {
	TypePath      string
	Root          string
	Name          string
	Flags         uint32
	NeedTransform uint32
	RotationX     float32
	RotationY     float32
	RotationZ     float32
	RotationW     float32
	PositionX     float32
	PositionY     float32
	PositionZ     float32
	ScaleX        float32
	ScaleY        float32
	ScaleZ        float32
	WasPlaced     uint32
}

type ActorObject struct {
	SaveVersion     uint32
	Flag            uint32
	Size            uint32
	ParentReference ObjectReference
	ComponentCount  uint32
	Components      []ObjectReference
	Properties      []Property
}

func (a *ActorObject) isValid() bool {
	lastProp := a.Properties[len(a.Properties)-1]
	return (a.Flag == 0 || a.Flag == 1) &&
		a.Size > 0 &&
		lastProp.Name == "None" && lastProp.Type == ""
}

type ComponentHeader struct {
	TypePath        string
	Root            string
	Name            string
	Flags           uint32
	ParentActorName string
}

type ComponentObject struct {
	SaveVersion uint32
	Flag        uint32
	Size        uint32
	Properties  []Property
	Zero        uint32
}

func (c *ComponentObject) isValid() bool {
	lastProp := c.Properties[len(c.Properties)-1]
	return (c.Flag == 0 || c.Flag == 1) &&
		c.Size > 0 && c.Zero == 0 &&
		lastProp.Name == "None" && lastProp.Type == ""
}

type ObjectReference struct {
	LevelName string
	PathName  string
}

func (body *SaveFileBody) isValid() bool {
	return body.Value6 == 6 &&
		body.Value0 == 0 &&
		body.Value1 == 1 &&
		body.NoneString1 == "None" &&
		body.NoneString2 == "None"
}

func readSaveFileBody(zr io.Reader, version uint32) (*SaveFileBody, error) {
	var body SaveFileBody

	readFields(zr,
		&body.UncompressedSize, &body.Value6, &body.NoneString1, &body.Value0,
		&body.Unknown1, &body.Value1, &body.NoneString2, &body.Unknown2,
	)

	if !body.isValid() {
		return nil, errors.New("invalid save file body header")
	}

	for range 5 {
		grid := readLevelGroupingGrid(zr)
		body.LevelGroupingGrids = append(body.LevelGroupingGrids, *grid)
	}

	readFields(zr, &body.SubLevelCount)

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

func readLevelGroupingGrid(r io.Reader) *LevelGroupingGrid {
	var grid LevelGroupingGrid

	readFields(r,
		&grid.GridName, &grid.Unknown1, &grid.Unknown2, &grid.LevelCount,
	)

	for range grid.LevelCount {
		var levelInfo LevelInfo
		readFields(r, &levelInfo.StringValue, &levelInfo.IntValue)
		grid.LevelInfos = append(grid.LevelInfos, levelInfo)
	}

	return &grid
}

// todo split into read header/object funcs?
func readLevelData(r io.Reader, version uint32, isPersistentLevel bool) *LevelData {
	var levelData LevelData

	if !isPersistentLevel {
		readFields(r, &levelData.Name)
	}

	readFields(r, &levelData.Size, &levelData.HeaderCount)

	headerTypes := make([]uint32, 0, levelData.HeaderCount)
	startPos := r.(*countingReader).Position()
	for range levelData.HeaderCount {
		var headerType uint32
		readFields(r, &headerType)
		headerTypes = append(headerTypes, headerType)
		if headerType == 0 {
			var componentHeader ComponentHeader
			readFields(r,
				&componentHeader.TypePath, &componentHeader.Root,
				&componentHeader.Name, conditionalFields(version >= 51, &componentHeader.Flags),
				&componentHeader.ParentActorName,
			)
			levelData.ComponentHeaders = append(levelData.ComponentHeaders, componentHeader)

		} else if headerType == 1 {
			var actorHeader ActorHeader
			readFields(r,
				&actorHeader.TypePath, &actorHeader.Root, &actorHeader.Name,
				conditionalFields(version >= 51, &actorHeader.Flags), &actorHeader.NeedTransform,
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

	endPos := r.(*countingReader).Position()
	bytesRead := uint64(endPos - startPos)
	diff := levelData.Size - bytesRead
	if diff > 4 {
		readFields(r, &levelData.CollectableCount)
		if levelData.CollectableCount > 0 && isPersistentLevel {
			var name string
			readFields(r, &name, &levelData.CollectableCount)
		}
		for range levelData.CollectableCount {
			var collectable ObjectReference
			readFields(r, &collectable.LevelName, &collectable.PathName)
		}
	}

	readFields(r, &levelData.ObjectSize, &levelData.ObjectCount)

	//go through objects
	for i := range levelData.ObjectCount {
		startPos := r.(*countingReader).Position()
		var objectSize uint32
		if headerTypes[i] == 0 {
			var component ComponentObject
			readFields(r, &component.SaveVersion, &component.Flag, &component.Size)
			objectSize = component.Size
			readAllProperties(r, &component.Properties)

			//Zero field not always present, part of trailing bytes?

			if !component.isValid() {
				panic("Invalid component object at index" + fmt.Sprint(i))
			}
			levelData.ComponentObjects = append(levelData.ComponentObjects, component)
		} else if headerTypes[i] == 1 {
			var actor ActorObject
			readFields(r, &actor.SaveVersion, &actor.Flag, &actor.Size,
				&actor.ParentReference, &actor.ComponentCount)
			objectSize = actor.Size
			for range actor.ComponentCount {
				var component ObjectReference
				readFields(r, &component.LevelName, &component.PathName)
				actor.Components = append(actor.Components, component)
			}
			readAllProperties(r, &actor.Properties)
			if !actor.isValid() {
				panic("Invalid actor object at index" + fmt.Sprint(i))
			}
			levelData.ActorObjects = append(levelData.ActorObjects, actor)
		}
		endPos := r.(*countingReader).Position()
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
		readFields(r, &v)
	}

	readFields(r, &levelData.SecondCollectableCount)
	for range levelData.SecondCollectableCount {
		var collectable ObjectReference
		readFields(r, &collectable.LevelName, &collectable.PathName)
	}

	return &levelData
}

// TODO return ref list here? do we need this?
func readReferenceList(zr io.Reader, body *SaveFileBody) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error reading ref list, skipping: ", r)
		}
	}()

	readFields(zr, &body.ReferenceListCount)
	for range body.ReferenceListCount {
		var reference ObjectReference
		readFields(zr, &reference.LevelName, &reference.PathName)
		body.References = append(body.References, reference)
	}
}
