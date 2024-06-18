package protocol

import (
	"bytes"
	"fmt"
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
