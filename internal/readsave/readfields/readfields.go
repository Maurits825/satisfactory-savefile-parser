package readfields

import (
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/Maurits825/satisfactory-savefile-parser/pkg/saveformat"
)

func ReadFields(r io.Reader, fields ...any) {
	for _, field := range fields {
		switch field := field.(type) {
		case *string:
			s, err := readString(r)
			if err != nil {
				panic("Error reading string field: " + err.Error())
			}
			*field = s
		case *saveformat.ObjectReference:
			ReadFields(r, &field.LevelName, &field.PathName)
		case *ArraySoftObjectProperty:
			ReadFields(r, &field.Reference, &field.Value)
		case nil:
			continue
		case []any:
			ReadFields(r, field...)
		default:
			err := binary.Read(r, binary.LittleEndian, field)
			if err != nil {
				panic("Error reading field: " + err.Error())
			}
		}
	}
}

func ConditionalFields(useValue bool, fields ...any) any {
	if useValue {
		return fields
	}
	return nil
}

func readString(r io.Reader) (string, error) {
	var length int32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return "", fmt.Errorf("reading string length: %w", err)
	}

	if length == 0 {
		return "", nil
	}

	if length > 0 {
		// UTF-8 string
		data := make([]byte, length)
		if _, err := io.ReadFull(r, data); err != nil {
			return "", fmt.Errorf("reading UTF-8 string: %w, len: %v", err, length)
		}

		// Strip null terminator if present
		if data[length-1] == 0 {
			data = data[:length-1]
		}

		if !utf8.Valid(data) {
			return "", fmt.Errorf("invalid UTF-8 string len: %v", length)
		}

		return string(data), nil

	} else {
		// UTF-16 LE string
		charCount := -length

		byteCount := charCount * 2
		data := make([]byte, byteCount)

		if _, err := io.ReadFull(r, data); err != nil {
			return "", fmt.Errorf("reading UTF-16 string: %w", err)
		}

		utf16Data := make([]uint16, charCount)
		for i := 0; i < int(charCount); i++ {
			utf16Data[i] = binary.LittleEndian.Uint16(data[i*2 : i*2+2])
		}

		// Remove null terminator if present
		if utf16Data[len(utf16Data)-1] == 0 {
			utf16Data = utf16Data[:len(utf16Data)-1]
		}

		// Decode UTF-16 to string
		return string(utf16.Decode(utf16Data)), nil
	}
}
