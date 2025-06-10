package saveformat

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

type Property struct {
	Name  string
	Type  string
	Value any
}

type ObjectReference struct {
	LevelName string
	PathName  string
}

func (body *SaveFileBody) IsValid() bool {
	return body.Value6 == 6 &&
		body.Value0 == 0 &&
		body.Value1 == 1 &&
		body.NoneString1 == "None" &&
		body.NoneString2 == "None"
}

func (a *ActorObject) IsValid() bool {
	lastProp := a.Properties[len(a.Properties)-1]
	return (a.Flag == 0 || a.Flag == 1) &&
		a.Size > 0 &&
		lastProp.Name == "None" && lastProp.Type == ""
}
func (c *ComponentObject) IsValid() bool {
	lastProp := c.Properties[len(c.Properties)-1]
	return (c.Flag == 0 || c.Flag == 1) &&
		c.Size > 0 && c.Zero == 0 &&
		lastProp.Name == "None" && lastProp.Type == ""
}

type ArrayStructProperty struct {
	Name        string
	Type        string
	Size        uint32
	Padding     uint32
	ElementType string
	Padding1    uint32
	Padding2    uint32
	Padding3    uint32
	Padding4    uint32
	PaddingByte byte
	Value       []any
}

type InventoryItem struct {
	Reference         ObjectReference
	ItemHasProperties uint32
	ItemType          ObjectReference
	PropertySize      uint32
	Properties        []Property
}
