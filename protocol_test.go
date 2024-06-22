package protocol

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func bytesToHexString(b []byte) string {
	var buf bytes.Buffer
	for _, v := range b {
		buf.WriteString(fmt.Sprintf("%02X ", v))
	}
	return buf.String()
}

func formatXXD(data []byte) string {
	var sb strings.Builder

	for i := 0; i < len(data); i += 16 {
		sb.WriteString(fmt.Sprintf("%08x: ", i))
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				sb.WriteString(fmt.Sprintf("%02x", data[i+j]))
				if j%2 == 1 {
					sb.WriteString(" ")
				}
			} else {
				sb.WriteString("   ")
				if j%2 == 1 {
					sb.WriteString(" ")
				}
			}
		}
		sb.WriteString(" ")
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				b := data[i+j]
				if b >= 32 && b <= 126 {
					sb.WriteString(fmt.Sprintf("%c", b))
				} else {
					sb.WriteString(".")
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func saveAsBin(data []byte, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func TestEncodeFixedArgument(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		result    []byte
		expectErr bool
	}{
		{"bool", true, []byte{
			TypeBool, 'b', 'o', 'o', 'l', 0xFF, 0x01, 0x01, 0x01,
		}, false},
		{"byte", byte(0xDE), []byte{
			TypeUInt8, 'b', 'y', 't', 'e', 0xFF, 0x01, 0x01, 0xDE,
		}, false},
		{"uint", uint16(0x1234), []byte{
			TypeUInt16, 'u', 'i', 'n', 't', 0xFF, 0x01, 0x02, 0x12, 0x34,
		}, false},
		{"int", int32(-0x12345678), []byte{
			TypeInt32, 'i', 'n', 't', 0xFF, 0x01, 0x04, 0xED, 0xCB, 0xA9, 0x88,
		}, false},
		{"float", float64(3.14), []byte{
			TypeFloat64, 'f', 'l', 'o', 'a', 't', 0xFF, 0x01, 0x08, 0x40, 0x09, 0x1E, 0xB8, 0x51, 0xEB, 0x85, 0x1F,
		}, false},
		{"complex", complex128(69.0 + 420.0i), []byte{
			TypeComplex128, 'c', 'o', 'm', 'p', 'l', 'e', 'x', 0xFF, 0x01, 0x10, 0x40, 0x51, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x7A, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00,
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := encodeArgument(buf, tt.value, tt.name)

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			resultBuf := bytes.NewBuffer(tt.result)
			err = writeChecksum(resultBuf)
			if err != nil {
				t.Fatalf("error writing checksum: %v", err)
			}

			if !bytes.Equal(buf.Bytes(), resultBuf.Bytes()) {
				t.Fatalf(`
expected: %s
got:      %s
				`, bytesToHexString(resultBuf.Bytes()), bytesToHexString(buf.Bytes()))
			}
		})
	}
}

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

func TestEncodeStringArgument(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		result    []byte
		expectErr bool
	}{
		{"regular", "moin dikka", []byte{
			TypeString, 'r', 'e', 'g', 'u', 'l', 'a', 'r', 0xFF, 0x01, 0x0A, 0x6D, 0x6F, 0x69, 0x6E, 0x20, 0x64, 0x69, 0x6B, 0x6B, 0x61,
		}, false},
		{"empty", "", []byte{
			TypeString, 'e', 'm', 'p', 't', 'y', 0xFF, 0x01, 0x00,
		}, false},
		{"special", "🤡🤡🤡", []byte{
			TypeString, 's', 'p', 'e', 'c', 'i', 'a', 'l', 0xFF, 0x01, 0x0C, 0xF0, 0x9F, 0xA4, 0xA1, 0xF0, 0x9F, 0xA4, 0xA1, 0xF0, 0x9F, 0xA4, 0xA1,
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := encodeArgument(buf, tt.value, tt.name)

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			resultBuf := bytes.NewBuffer(tt.result)
			err = writeChecksum(resultBuf)
			if err != nil {
				t.Fatalf("error writing checksum: %v", err)
			}

			if !bytes.Equal(buf.Bytes(), resultBuf.Bytes()) {
				t.Fatalf(`
expected: %s
got:      %s
				`, bytesToHexString(resultBuf.Bytes()), bytesToHexString(buf.Bytes()))
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

func TestEncodeStructArgument(t *testing.T) {
	tests := []struct {
		name         string
		value        any
		outerResult  []byte
		innerResults [][]byte
		expectErr    bool
	}{
		{"only primitives", struct {
			Bool bool
			Byte byte
		}{
			true, 0xDE,
		}, []byte{
			TypeStruct, 'o', 'n', 'l', 'y', ' ', 'p', 'r', 'i', 'm', 'i', 't', 'i', 'v', 'e', 's', 0xFF, 0x01, 0x1A,
		}, [][]byte{
			{
				TypeBool, 'B', 'o', 'o', 'l', 0xFF, 0x01, 0x01, 0x01,
			},
			{
				TypeUInt8, 'B', 'y', 't', 'e', 0xFF, 0x01, 0x01, 0xDE,
			},
		}, false},
		{"fixed and string", struct {
			Byte byte
			Str  string
		}{
			0xDE, "moin dikka",
		}, []byte{
			TypeStruct, 'f', 'i', 'x', 'e', 'd', ' ', 'a', 'n', 'd', ' ', 's', 't', 'r', 'i', 'n', 'g', 0xFF, 0x01, 0x22,
		}, [][]byte{
			{
				TypeUInt8, 'B', 'y', 't', 'e', 0xFF, 0x01, 0x01, 0xDE,
			},
			{
				TypeString, 'S', 't', 'r', 0xFF, 0x01, 0x0A, 0x6D, 0x6F, 0x69, 0x6E, 0x20, 0x64, 0x69, 0x6B, 0x6B, 0x61,
			},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := encodeArgument(buf, tt.value, tt.name)

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			innerBuf := bytes.NewBuffer(nil)
			for _, innerResult := range tt.innerResults {
				tmpBuf := bytes.NewBuffer(innerResult)
				err = writeChecksum(tmpBuf)
				if err != nil {
					t.Fatalf("error writing checksum: %v", err)
				}
				innerBuf.Write(tmpBuf.Bytes())
			}

			resultBuf := bytes.NewBuffer(tt.outerResult)
			resultBuf.Write(innerBuf.Bytes())

			err = writeChecksum(resultBuf)
			if err != nil {
				t.Fatalf("error writing checksum: %v", err)
			}

			if !bytes.Equal(buf.Bytes(), resultBuf.Bytes()) {
				t.Fatalf(`
expected:
%s
got:
%s
				`, formatXXD(resultBuf.Bytes()), formatXXD(buf.Bytes()))
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

func TestEncodeNestedStruct(t *testing.T) {
	type inner struct {
		Byte byte
	}
	type outer struct {
		Inner inner
	}

	value := outer{
		Inner: inner{
			Byte: 0xDE,
		},
	}

	innerContentExpected := bytes.NewBuffer([]byte{
		TypeUInt8, 'B', 'y', 't', 'e', 0xFF, 0x01, 0x01, 0xDE,
	})
	err := writeChecksum(innerContentExpected)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	innerExpected := bytes.NewBuffer([]byte{
		TypeStruct, 'I', 'n', 'n', 'e', 'r', 0xFF, 0x01, byte(innerContentExpected.Len()),
	})
	innerExpected.Write(innerContentExpected.Bytes())
	err = writeChecksum(innerExpected)
	if err != nil {
		return
	}

	outerExpected := bytes.NewBuffer([]byte{
		TypeStruct, 'n', 'e', 's', 't', 'e', 'd', 0xFF, 0x01, byte(innerExpected.Len()),
	})
	outerExpected.Write(innerExpected.Bytes())
	err = writeChecksum(outerExpected)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(nil)
	err = encodeArgument(buf, value, "nested")
	if err != nil {
		t.Fatalf("error encoding argument: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), outerExpected.Bytes()) {
		t.Fatalf(`
expected:
%s
got:
%s
		`, formatXXD(outerExpected.Bytes()), formatXXD(buf.Bytes()))
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

func TestEncodeSlice(t *testing.T) {
	tests := []struct {
		name         string
		value        any
		outerResult  []byte
		innerResults [][]byte
		expectErr    bool
	}{
		{"primitives", []byte{0xDE, 0x68, 0xAA}, []byte{
			TypeSlice, 'p', 'r', 'i', 'm', 'i', 't', 'i', 'v', 'e', 's', 0xFF, 0x01, 0x1B,
		}, [][]byte{
			{
				TypeUInt8, 0xFF, 0x01, 0x01, 0xDE,
			},
			{
				TypeUInt8, 0xFF, 0x01, 0x01, 0x68,
			},
			{
				TypeUInt8, 0xFF, 0x01, 0x01, 0xAA,
			},
		}, false},
		{"strings", []string{"moin", "dikka"}, []byte{
			TypeSlice, 's', 't', 'r', 'i', 'n', 'g', 's', 0xFF, 0x01, 0x19,
		}, [][]byte{
			{
				TypeString, 0xFF, 0x01, 0x04, 'm', 'o', 'i', 'n',
			},
			{
				TypeString, 0xFF, 0x01, 0x05, 'd', 'i', 'k', 'k', 'a',
			},
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := encodeArgument(buf, tt.value, tt.name)

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			innerBuf := bytes.NewBuffer(nil)
			for _, innerResult := range tt.innerResults {
				tmpBuf := bytes.NewBuffer(innerResult)
				err = writeChecksum(tmpBuf)
				if err != nil {
					t.Fatalf("error writing checksum: %v", err)
				}
				innerBuf.Write(tmpBuf.Bytes())
			}

			resultBuf := bytes.NewBuffer(tt.outerResult)
			resultBuf.Write(innerBuf.Bytes())

			err = writeChecksum(resultBuf)
			if err != nil {
				t.Fatalf("error writing checksum: %v", err)
			}

			if !bytes.Equal(buf.Bytes(), resultBuf.Bytes()) {
				t.Fatalf(`
expected:
%s
got:
%s
				`, formatXXD(resultBuf.Bytes()), formatXXD(buf.Bytes()))
			}
		})
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

func TestEncodeNestedSlice(t *testing.T) {
	inner1 := []byte{0xDE, 0x68}
	inner2 := []byte{0xAA}
	value := [][]byte{
		inner1,
		inner2,
	}

	inner1ContentExpectedA := bytes.NewBuffer([]byte{
		TypeUInt8, 0xFF, 0x01, 0x01, 0xDE,
	})
	err := writeChecksum(inner1ContentExpectedA)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}
	inner1ContentExpectedB := bytes.NewBuffer([]byte{
		TypeUInt8, 0xFF, 0x01, 0x01, 0x68,
	})
	err = writeChecksum(inner1ContentExpectedB)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}
	inner1ContentExpected := bytes.NewBuffer(nil)
	inner1ContentExpected.Write(inner1ContentExpectedA.Bytes())
	inner1ContentExpected.Write(inner1ContentExpectedB.Bytes())
	inner1Expected := bytes.NewBuffer([]byte{
		TypeSlice, 0xFF, 0x01, byte(inner1ContentExpected.Len()),
	})

	inner1Expected.Write(inner1ContentExpected.Bytes())
	err = writeChecksum(inner1Expected)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	inner2ContentExpected := bytes.NewBuffer([]byte{
		TypeUInt8, 0xFF, 0x01, 0x01, 0xAA,
	})
	err = writeChecksum(inner2ContentExpected)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}
	inner2Expected := bytes.NewBuffer([]byte{
		TypeSlice, 0xFF, 0x01, byte(inner2ContentExpected.Len()),
	})

	inner2Expected.Write(inner2ContentExpected.Bytes())
	err = writeChecksum(inner2Expected)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	totalInnerSize := inner1Expected.Len() + inner2Expected.Len()
	outerExpected := bytes.NewBuffer([]byte{
		TypeSlice, 'n', 'e', 's', 't', 'e', 'd', 0xFF, 0x01, byte(totalInnerSize),
	})
	outerExpected.Write(inner1Expected.Bytes())
	outerExpected.Write(inner2Expected.Bytes())
	err = writeChecksum(outerExpected)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	buf := bytes.NewBuffer(nil)
	err = encodeArgument(buf, value, "nested")
	if err != nil {
		t.Fatalf("error encoding argument: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), outerExpected.Bytes()) {
		t.Fatalf(`
expected:
%s
got:
%s
`, formatXXD(outerExpected.Bytes()), formatXXD(buf.Bytes()))
	}
}
