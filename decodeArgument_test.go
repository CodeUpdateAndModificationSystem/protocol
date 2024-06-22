package protocol

import (
	"bytes"
	"reflect"
	"testing"
)

func TestDecodeFixedArgument(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		expected  any
		expectErr bool
	}{
		{"bool", []byte{
			TypeBool, 'b', 'o', 'o', 'l', 0xFF, 0x01, 0x01, 0x01,
		}, true, false},
		{"byte", []byte{
			TypeUInt8, 'b', 'y', 't', 'e', 0xFF, 0x01, 0x01, 0xDE,
		}, byte(0xDE), false},
		{"uint", []byte{
			TypeUInt16, 'u', 'i', 'n', 't', 0xFF, 0x01, 0x02, 0x12, 0x34,
		}, uint16(0x1234), false},
		{"int", []byte{
			TypeInt32, 'i', 'n', 't', 0xFF, 0x01, 0x04, 0xED, 0xCB, 0xA9, 0x88,
		}, int32(-0x12345678), false},
		{"float", []byte{
			TypeFloat64, 'f', 'l', 'o', 'a', 't', 0xFF, 0x01, 0x08, 0x40, 0x09, 0x1E, 0xB8, 0x51, 0xEB, 0x85, 0x1F,
		}, float64(3.14), false},
		{"complex", []byte{
			TypeComplex128, 'c', 'o', 'm', 'p', 'l', 'e', 'x', 0xFF, 0x01, 0x10, 0x40, 0x51, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x7A, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00,
		}, complex128(69.0 + 420.0i), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.data)
			err := writeChecksum(buffer)
			if err != nil {
				t.Fatalf("error writing checksum: %v", err)
			}

			name, value, typ, err := decodeArgument(buffer.Bytes())

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			if name != tt.name {
				t.Fatalf("expected name: %v, got: %v", tt.name, name)
			}

			if typ != tt.data[0] {
				t.Fatalf("expected type: %v, got: %v", TypeToString[tt.data[0]], TypeToString[typ])
			}

			if value != tt.expected {
				t.Fatalf("expected: %v, got: %v", tt.expected, value)
			}
		})
	}
}

func TestDecodeStringArgument(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		expected  string
		expectErr bool
	}{
		{"regular", []byte{
			TypeString, 'r', 'e', 'g', 'u', 'l', 'a', 'r', 0xFF, 0x01, 0x0A, 'm', 'o', 'i', 'n', ' ', 'd', 'i', 'k', 'k', 'a',
		}, "moin dikka", false},
		{"empty", []byte{
			TypeString, 'e', 'm', 'p', 't', 'y', 0xFF, 0x01, 0x00,
		}, "", false},
		{"special", []byte{
			TypeString, 's', 'p', 'e', 'c', 'i', 'a', 'l', 0xFF, 0x01, 0x0C, 0xF0, 0x9F, 0xA4, 0xA1, 0xF0, 0x9F, 0xA4, 0xA1, 0xF0, 0x9F, 0xA4, 0xA1,
		}, "🤡🤡🤡", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.data)
			err := writeChecksum(buffer)
			if err != nil {
				t.Fatalf("error writing checksum: %v", err)
			}

			name, value, typ, err := decodeArgument(buffer.Bytes())

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			if name != tt.name {
				t.Fatalf("expected name: %v, got: %v", tt.name, name)
			}

			if typ != tt.data[0] {
				t.Fatalf("expected type: %v, got: %v", TypeToString[tt.data[0]], TypeToString[typ])
			}

			if value != tt.expected {
				t.Fatalf("expected: %v, got: %v", tt.expected, value)
			}
		})
	}
}

func TestDecodeStructArgument(t *testing.T) {
	tests := []struct {
		name      string
		dataRaw   any
		expectErr bool
	}{
		{"only primitives", struct {
			Bool bool
			Byte byte
		}{
			true, 0xDE,
		}, false},
		{"fixed and string", struct {
			Byte byte
			Str  string
		}{
			0xDE, "moin dikka",
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := bytes.NewBuffer(nil)
			err := encodeArgument(data, tt.dataRaw, tt.name)
			if err != nil {
				t.Fatalf("error encoding argument: %v", err)
			}

			buffer := data.Bytes()
			name, value, typ, err := decodeArgument(buffer)

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			if name != tt.name {
				t.Fatalf("expected name: %v, got: %v", tt.name, name)
			}

			if typ != TypeStruct {
				t.Fatalf("expected type: %v, got: %v", TypeToString[TypeStruct], TypeToString[typ])
			}

			if !reflect.DeepEqual(value, tt.dataRaw) {
				t.Fatalf("expected: %v, got: %v", tt.dataRaw, value)
			}
		})

	}
}

func TestSplitArgumentListData(t *testing.T) {
	first := bytes.NewBuffer([]byte{
		TypeUInt8, 'B', 'y', 't', 'e', 0xFF, 0x01, 0x01, 0xDE,
	})
	err := writeChecksum(first)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}
	second := bytes.NewBuffer([]byte{
		TypeString, 'S', 't', 'r', 0xFF, 0x01, 0x0A, 0x6D, 0x6F, 0x69, 0x6E, 0x20, 0x64, 0x69, 0x6B, 0x6B, 0x61,
	})
	err = writeChecksum(second)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	data := bytes.NewBuffer(nil)
	data.Write(first.Bytes())
	data.Write(second.Bytes())

	splitData, err := splitArgumentListData(data.Bytes())
	if err != nil {
		t.Fatalf("error splitting data: %v", err)
	}

	if len(splitData) != 2 {
		t.Fatalf("expected 2 elements, got: %v", len(splitData))
	}

	if !bytes.Equal(splitData[0], first.Bytes()) {
		t.Fatalf(`
expected:
%s
got:
%s
`, formatXXD(first.Bytes()), formatXXD(splitData[0]))
	}

	if !bytes.Equal(splitData[1], second.Bytes()) {
		t.Fatalf(`
expected:
%s
got:
%s
`, formatXXD(second.Bytes()), formatXXD(splitData[1]))
	}

}

func TestDecodeNestedStruct(t *testing.T) {
	type inner struct {
		Byte byte
	}
	type outer struct {
		Inner inner
	}

	dataRaw := outer{
		Inner: inner{
			Byte: 0xDE,
		},
	}

	data := bytes.NewBuffer(nil)
	err := encodeArgument(data, dataRaw, "nested")
	if err != nil {
		t.Fatalf("error encoding argument: %v", err)
	}

	buffer := data.Bytes()
	name, value, typ, err := decodeArgument(buffer)

	if err != nil {
		t.Fatalf("error decoding argument: %v", err)
	}

	if name != "nested" {
		t.Fatalf("expected name: %v, got: %v", "nested", name)
	}

	if typ != TypeStruct {
		t.Fatalf("expected type: %v, got: %v", TypeToString[TypeStruct], TypeToString[typ])
	}

	if !reflect.DeepEqual(value, value) {
		t.Fatalf("expected: %v, got: %v", value, value)
	}
}

func TestDecodeSlice(t *testing.T) {
	tests := []struct {
		name      string
		dataRaw   any
		expectErr bool
	}{
		{"primitives", []byte{0xDE, 0x68, 0xAA}, false},
		{"strings", []string{"moin", "dikka"}, false},
		{"empty", []any{}, false},
		{"mixed", []any{byte(0xDE), "moin"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := bytes.NewBuffer(nil)
			err := encodeArgument(data, tt.dataRaw, tt.name)
			if err != nil {
				t.Fatalf("error encoding argument: %v", err)
			}

			buffer := data.Bytes()
			name, value, typ, err := decodeArgument(buffer)

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			if name != tt.name {
				t.Fatalf("expected name: %v, got: %v", tt.name, name)
			}

			if typ != TypeSlice {
				t.Fatalf("expected type: %v, got: %v", TypeToString[TypeSlice], TypeToString[typ])
			}

			if !reflect.DeepEqual(value, tt.dataRaw) {
				t.Fatalf("expected: %#v, got: %#v", tt.dataRaw, value)
			}
		})

	}
}
