package protocol

import (
	"bytes"
	"fmt"
	"testing"
)

func bytesToHexString(b []byte) string {
	var buf bytes.Buffer
	for _, v := range b {
		buf.WriteString(fmt.Sprintf("%02X ", v))
	}
	return buf.String()
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
			writeChecksum(resultBuf)
			if err != nil {
				t.Fatalf("error writing checksum: %v", err)
			}

			if !bytes.Equal(buf.Bytes(), resultBuf.Bytes()) {
				t.Fatalf(`
expected: %s
got:      %s
				`, bytesToHexString(tt.result), bytesToHexString(buf.Bytes()))
			}
		})
	}
}