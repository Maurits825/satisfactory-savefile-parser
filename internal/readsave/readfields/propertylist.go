package readfields

import (
	"encoding/binary"
	"io"

	"github.com/Maurits825/satisfactory-savefile-parser/internal/countingreader"
	"github.com/Maurits825/satisfactory-savefile-parser/pkg/saveformat"
)

type PropertyHeader struct {
	Size    uint32
	Index   uint32
	Padding byte
}

type BoolProperty struct {
	Padding1 uint32
	Index    uint32
	Value    byte
	Padding2 byte
}

type ByteProperty struct {
	Size    uint32
	Index   uint32
	Type    string
	Padding byte
	Value   any
}

type EnumProperty struct {
	Size    uint32
	Index   uint32
	Type    string
	Padding byte
	Value   string
}

type GenericProperty[T any] struct {
	Size    uint32
	Index   uint32
	Padding byte
	Value   T
}

func readGenericProperty[T any](r io.Reader) T {
	var p GenericProperty[T]
	ReadFields(r, &p.Size, &p.Index, &p.Padding, &p.Value)
	return p.Value
}

var genericPropertyReaders = map[string]func(io.Reader) any{
	"IntProperty":    func(r io.Reader) any { return readGenericProperty[int32](r) },
	"FloatProperty":  func(r io.Reader) any { return readGenericProperty[float32](r) },
	"DoubleProperty": func(r io.Reader) any { return readGenericProperty[float64](r) },
	"Int8Property":   func(r io.Reader) any { return readGenericProperty[int8](r) },
	"Int64Property":  func(r io.Reader) any { return readGenericProperty[int64](r) },
	"UInt32Property": func(r io.Reader) any { return readGenericProperty[uint32](r) },
	"StrProperty":    func(r io.Reader) any { return readGenericProperty[string](r) },
	"NameProperty":   func(r io.Reader) any { return readGenericProperty[string](r) },
}

type ObjectProperty struct {
	PropertyHeader
	Value saveformat.ObjectReference
}

type SoftObjectProperty struct {
	PropertyHeader
	ObjectReferenceValue saveformat.ObjectReference
	Value                uint32
}

type TextProperty struct {
	PropertyHeader
	Flags              uint32
	HistoryType        int8
	IsCultureInvariant uint32
	Value              string
}

type MapProperty struct {
	Size        uint32
	Index       uint32
	KeyType     string
	ValueType   string
	Padding     byte
	Mode        uint32
	NumElements uint32
	// Elements    []MapElement //TODO
}

type SetProperty struct {
	Size     uint32
	Index    uint32
	Type     string
	Padding1 byte
	Padding2 uint32
	Length   uint32
	// Elements   []interface{} //TODO
}

type StructProperty struct {
	Size     uint32
	Index    uint32
	Type     string
	Padding1 int64
	Padding2 int64
	Padding3 byte
	Value    any
}

type ArrayProperty struct {
	Size    uint32
	Index   uint32
	Type    string
	Padding byte
	Length  uint32
	Value   []any
}

type ArraySoftObjectProperty struct {
	Reference saveformat.ObjectReference
	Value     uint32
}

type Box struct {
	MinX    float64
	MinY    float64
	MinZ    float64
	MaxX    float64
	MaxY    float64
	MaxZ    float64
	IsValid byte
}

type FluidBox struct {
	Value float32
}

type LinearColor struct {
	R float32
	G float32
	B float32
	A float32
}

type Quat struct {
	X float64
	Y float64
	Z float64
	W float64
}

type RailroadTrackPosition struct {
	ObjectRef saveformat.ObjectReference
	Offset    float32
	Forward   float32
}

type Vector struct {
	X float64
	Y float64
	Z float64
}

type DateTime struct {
	Timestamp int64
}

type ClientIdentityInfo struct {
	UUID          string
	IdentityCount uint32
	Identities    []Identity
}

type Identity struct {
	Type     byte
	DataSize uint32
	Data     []byte
}

func ReadAllProperties(cr *countingreader.CountingReader, props *[]saveformat.Property) {
	for {
		var p saveformat.Property
		ReadFields(cr, &p.Name)
		if p.Name == "None" {
			*props = append(*props, p)
			return
		} else if p.Name == "" {
			//there can be a buggy byte on InventoryItem...
			ReadFields(cr, &p.Name)
		}

		ReadFields(cr, &p.Type)
		p.Value = readPropertyData(cr, p.Type)
		*props = append(*props, p)
	}
}

func eatPropertyData(r io.Reader, pSize uint32) {
	if _, err := io.CopyN(io.Discard, r, int64(pSize)); err != nil {
		panic("Error reading property data: " + err.Error())
	}
}

func readPropertyData(cr *countingreader.CountingReader, propertyType string) any {
	if genericReader, ok := genericPropertyReaders[propertyType]; ok {
		return genericReader(cr)
	}

	//TODO return p? instead of p.value, so we get all the data?
	switch propertyType {
	case "BoolProperty":
		var p BoolProperty
		ReadFields(cr, &p.Padding1, &p.Index, &p.Value, &p.Padding2)
		return p.Value
	case "ByteProperty":
		var p ByteProperty
		ReadFields(cr, &p.Size, &p.Index, &p.Type, &p.Padding)
		if p.Type == "None" {
			var b byte
			ReadFields(cr, &b)
			p.Value = b
		} else {
			var s string
			ReadFields(cr, &s)
			p.Value = s
		}
		return p.Value
	case "ObjectProperty":
		var p ObjectProperty
		ReadFields(cr, &p.Size, &p.Index, &p.Padding, &p.Value)
		return p.Value
	case "SoftObjectProperty":
		var p SoftObjectProperty
		ReadFields(cr, &p.Size, &p.Index, &p.Padding, &p.ObjectReferenceValue, &p.Value)
		return p.Value
	case "SetProperty":
		var p SetProperty
		ReadFields(cr, &p.Size, &p.Index, &p.Type, &p.Padding1)
		eatPropertyData(cr, p.Size)
		return nil
	case "StructProperty":
		var p StructProperty
		ReadFields(cr, &p.Size, &p.Index, &p.Type, &p.Padding1, &p.Padding2, &p.Padding3)
		p.Value = readTypedData(cr, p.Type)
		return p.Value
	case "ArrayProperty":
		return readArrayProperty(cr)
	case "EnumProperty":
		var p EnumProperty
		ReadFields(cr, &p.Size, &p.Index, &p.Type, &p.Padding)
		eatPropertyData(cr, p.Size)
		return nil
	case "MapProperty":
		var p MapProperty
		ReadFields(cr, &p.Size, &p.Index, &p.KeyType, &p.ValueType, &p.Padding)
		eatPropertyData(cr, p.Size)
		return nil
	case "TextProperty":
		var pSize uint32
		ReadFields(cr, &pSize)
		eatPropertyData(cr, pSize+5)
		return nil
	default:
		panic("not implemented property type: " + propertyType)
	}
}

func readArrayValues[T any](r io.Reader, length uint32) []T {
	values := make([]T, length)
	for i := range length {
		var value T
		ReadFields(r, &value)
		values[i] = value
	}
	return values
}

func readArrayStructProperty(cr *countingreader.CountingReader, length uint32) saveformat.ArrayStructProperty {
	var p saveformat.ArrayStructProperty
	ReadFields(cr, &p.Name, &p.Type, &p.Size, &p.Padding, &p.ElementType,
		&p.Padding1, &p.Padding2, &p.Padding3, &p.Padding4, &p.PaddingByte,
	)

	read := func() uint32 {
		for range length {
			value := readTypedData(cr, p.ElementType)
			p.Value = append(p.Value, value)
		}
		return p.Size
	}
	countingreader.ReadAndYeet(cr, read)
	return p
}

func readTypedData(cr *countingreader.CountingReader, elementType string) any {
	var value any
	switch elementType {
	case "Box":
		var v Box
		ReadFields(cr, &v.MinX, &v.MinY, &v.MinZ, &v.MaxX, &v.MaxY, &v.MaxZ, &v.IsValid)
		value = v
	case "FluidBox":
		var v FluidBox
		ReadFields(cr, &v.Value)
		value = v
	case "Vector":
		var v Vector
		ReadFields(cr, &v.X, &v.Y, &v.Z)
		value = v
	case "DateTime":
		var v DateTime
		ReadFields(cr, &v.Timestamp)
		value = v
	case "InventoryItem":
		var v saveformat.InventoryItem
		ReadFields(cr, &v.Reference, &v.ItemHasProperties)

		if v.ItemHasProperties != 0 {
			ReadFields(cr, &v.ItemType, &v.PropertySize)
			ReadAllProperties(cr, &v.Properties)
		}
		value = v
	case "LinearColor":
		var v LinearColor
		ReadFields(cr, &v.R, &v.G, &v.B, &v.A)
		value = v
	case "Quat":
		var v Quat
		ReadFields(cr, &v.X, &v.Y, &v.Z, &v.W)
		value = v
	case "RailroadTrackPosition":
		var v RailroadTrackPosition
		ReadFields(cr, &v.ObjectRef, &v.Offset, &v.Forward)
		value = v
	case "Guid":
		//todo hex type? we will need the length
		var v1, v2 int64
		ReadFields(cr, &v1, &v2)
		value = v1
	case "ClientIdentityInfo":
		var v ClientIdentityInfo
		ReadFields(cr, &v.UUID, &v.IdentityCount)
		for range v.IdentityCount {
			var id Identity
			ReadFields(cr, &id.Type, &id.DataSize)
			id.Data = make([]byte, id.DataSize)
			err := binary.Read(cr, binary.LittleEndian, &id.Data)
			if err != nil {
				panic(err)
			}
			v.Identities = append(v.Identities, id)
		}
		value = v
	default:
		props := make([]saveformat.Property, 0)
		ReadAllProperties(cr, &props)
		value = props
	}
	return value
}

func readArrayProperty(cr *countingreader.CountingReader) any {
	var p ArrayProperty
	ReadFields(cr, &p.Size, &p.Index, &p.Type, &p.Padding, &p.Length)

	var values any
	switch p.Type {
	case "ByteProperty":
		values = readArrayValues[byte](cr, p.Length)
	case "EnumProperty", "StrProperty":
		values = readArrayValues[string](cr, p.Length)
	case "ObjectProperty", "InterfaceProperty":
		values = readArrayValues[saveformat.ObjectReference](cr, p.Length)
	case "IntProperty":
		values = readArrayValues[int32](cr, p.Length)
	case "Int64Property":
		values = readArrayValues[int64](cr, p.Length)
	case "FloatProperty":
		values = readArrayValues[float32](cr, p.Length)
	case "SoftObjectProperty":
		values = readArrayValues[ArraySoftObjectProperty](cr, p.Length)
	case "StructProperty":
		values = readArrayStructProperty(cr, p.Length)

	default:
		eatPropertyData(cr, p.Size-4)
	}

	return values
}
