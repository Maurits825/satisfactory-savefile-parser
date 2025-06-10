package main

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Property struct {
	Name  string
	Type  string
	Value any
}

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
	readFields(r, &p.Size, &p.Index, &p.Padding, &p.Value)
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
	Value ObjectReference
}

type SoftObjectProperty struct {
	PropertyHeader
	ObjectReferenceValue ObjectReference
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
	Reference ObjectReference
	Value     uint32
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

type InventoryItem struct {
	Reference         ObjectReference
	ItemHasProperties uint32
	ItemType          ObjectReference
	PropertySize      uint32
	Properties        []Property
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
	ObjectRef ObjectReference
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

func readAllProperties(r io.Reader, props *[]Property) {
	for {
		var p Property
		readFields(r, &p.Name)
		if p.Name == "None" {
			*props = append(*props, p)
			return
		} else if p.Name == "" {
			//there can be a buggy byte on InventoryItem...
			readFields(r, &p.Name)
		}

		readFields(r, &p.Type)
		p.Value = readPropertyData(p.Type, r)
		*props = append(*props, p)
	}
}

func eatPropertyData(r io.Reader, pSize uint32) {
	if _, err := io.CopyN(io.Discard, r, int64(pSize)); err != nil {
		panic("Error reading property data: " + err.Error())
	}
}

func readPropertyData(propertyType string, r io.Reader) any {
	if genericReader, ok := genericPropertyReaders[propertyType]; ok {
		return genericReader(r)
	}

	//TODO return p? instead of p.value, so we get all the data?
	switch propertyType {
	case "BoolProperty":
		var p BoolProperty
		readFields(r, &p.Padding1, &p.Index, &p.Value, &p.Padding2)
		return p.Value
	case "ByteProperty":
		var p ByteProperty
		readFields(r, &p.Size, &p.Index, &p.Type, &p.Padding)
		if p.Type == "None" {
			var b byte
			readFields(r, &b)
			p.Value = b
		} else {
			var s string
			readFields(r, &s)
			p.Value = s
		}
		return p.Value
	case "ObjectProperty":
		var p ObjectProperty
		readFields(r, &p.Size, &p.Index, &p.Padding, &p.Value)
		return p.Value
	case "SoftObjectProperty":
		var p SoftObjectProperty
		readFields(r, &p.Size, &p.Index, &p.Padding, &p.ObjectReferenceValue, &p.Value)
		return p.Value
	case "SetProperty":
		var p SetProperty
		readFields(r, &p.Size, &p.Index, &p.Type, &p.Padding1)
		eatPropertyData(r, p.Size)
		return nil
	case "StructProperty":
		var p StructProperty
		readFields(r, &p.Size, &p.Index, &p.Type, &p.Padding1, &p.Padding2, &p.Padding3)
		p.Value = readTypedData(r, p.Type)
		return p.Value
	case "ArrayProperty":
		return readArrayProperty(r)
	case "EnumProperty":
		var p EnumProperty
		readFields(r, &p.Size, &p.Index, &p.Type, &p.Padding)
		eatPropertyData(r, p.Size)
		return nil
	case "MapProperty":
		var p MapProperty
		readFields(r, &p.Size, &p.Index, &p.KeyType, &p.ValueType, &p.Padding)
		eatPropertyData(r, p.Size)
		return nil
	case "TextProperty":
		var pSize uint32
		readFields(r, &pSize)
		eatPropertyData(r, pSize+5)
		return nil
	default:
		panic("not implemented property type: " + propertyType)
	}
}

func readArrayValues[T any](r io.Reader, length uint32) []T {
	values := make([]T, length)
	for i := range length {
		var value T
		readFields(r, &value)
		values[i] = value
	}
	return values
}

func readArrayStructProperty(r io.Reader, length uint32) ArrayStructProperty {
	var p ArrayStructProperty
	readFields(r, &p.Name, &p.Type, &p.Size, &p.Padding, &p.ElementType,
		&p.Padding1, &p.Padding2, &p.Padding3, &p.Padding4, &p.PaddingByte,
	)
	startPos := r.(*countingReader).Position()
	for range length {
		value := readTypedData(r, p.ElementType)
		p.Value = append(p.Value, value)
	}
	endPos := r.(*countingReader).Position()
	bytesRead := uint32(endPos - startPos)

	structSize := p.Size
	if bytesRead < structSize {
		trailingBytes := structSize - bytesRead
		if trailingBytes != 0 {
			if _, err := io.CopyN(io.Discard, r, int64(trailingBytes)); err != nil {
				panic(err)
			}
		}
	} else if bytesRead > structSize {
		fmt.Println("Warning: read more bytes than expected ArrayStructProp")
	}

	return p
}

func readTypedData(r io.Reader, elementType string) any {
	var value any
	switch elementType {
	case "Box":
		var v Box
		readFields(r, &v.MinX, &v.MinY, &v.MinZ, &v.MaxX, &v.MaxY, &v.MaxZ, &v.IsValid)
		value = v
	case "FluidBox":
		var v FluidBox
		readFields(r, &v.Value)
		value = v
	case "Vector":
		var v Vector
		readFields(r, &v.X, &v.Y, &v.Z)
		value = v
	case "DateTime":
		var v DateTime
		readFields(r, &v.Timestamp)
		value = v
	case "InventoryItem":
		var v InventoryItem
		readFields(r, &v.Reference, &v.ItemHasProperties)

		if v.ItemHasProperties != 0 {
			readFields(r, &v.ItemType, &v.PropertySize)
			readAllProperties(r, &v.Properties)
		}
		value = v
	case "LinearColor":
		var v LinearColor
		readFields(r, &v.R, &v.G, &v.B, &v.A)
		value = v
	case "Quat":
		var v Quat
		readFields(r, &v.X, &v.Y, &v.Z, &v.W)
		value = v
	case "RailroadTrackPosition":
		var v RailroadTrackPosition
		readFields(r, &v.ObjectRef, &v.Offset, &v.Forward)
		value = v
	case "Guid":
		//todo hex type? we will need the length
		var v1, v2 int64
		readFields(r, &v1, &v2)
		value = v1
	case "ClientIdentityInfo":
		var v ClientIdentityInfo
		readFields(r, &v.UUID, &v.IdentityCount)
		for range v.IdentityCount {
			var id Identity
			readFields(r, &id.Type, &id.DataSize)
			id.Data = make([]byte, id.DataSize)
			err := binary.Read(r, binary.LittleEndian, &id.Data)
			if err != nil {
				panic(err)
			}
			v.Identities = append(v.Identities, id)
		}
		value = v
	default:
		props := make([]Property, 0)
		readAllProperties(r, &props)
		value = props
	}
	return value
}

func readArrayProperty(r io.Reader) any {
	var p ArrayProperty
	readFields(r, &p.Size, &p.Index, &p.Type, &p.Padding, &p.Length)

	var values any
	switch p.Type {
	case "ByteProperty":
		values = readArrayValues[byte](r, p.Length)
	case "EnumProperty", "StrProperty":
		values = readArrayValues[string](r, p.Length)
	case "ObjectProperty", "InterfaceProperty":
		values = readArrayValues[ObjectReference](r, p.Length)
	case "IntProperty":
		values = readArrayValues[int32](r, p.Length)
	case "Int64Property":
		values = readArrayValues[int64](r, p.Length)
	case "FloatProperty":
		values = readArrayValues[float32](r, p.Length)
	case "SoftObjectProperty":
		values = readArrayValues[ArraySoftObjectProperty](r, p.Length)
	case "StructProperty":
		values = readArrayStructProperty(r, p.Length)

	default:
		eatPropertyData(r, p.Size-4)
	}

	return values
}
