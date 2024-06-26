package protocol

import (
	"bytes"
	"math"
	"testing"
)

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

func TestEncodingShrinkingArgument(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		result []byte
	}{
		{"int8", math.MaxInt8, []byte{TypeInt8, 'i', 'n', 't', '8', 0xFF, 0x01, 0x01, 0x7F}},
		{"int16", math.MaxInt16, []byte{TypeInt16, 'i', 'n', 't', '1', '6', 0xFF, 0x01, 0x02, 0x7F, 0xFF}},
		{"int32", math.MaxInt32, []byte{TypeInt32, 'i', 'n', 't', '3', '2', 0xFF, 0x01, 0x04, 0x7F, 0xFF, 0xFF, 0xFF}},
		{"int64", math.MaxInt64, []byte{TypeInt64, 'i', 'n', 't', '6', '4', 0xFF, 0x01, 0x08, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"uint8", uint(math.MaxUint8), []byte{TypeUInt8, 'u', 'i', 'n', 't', '8', 0xFF, 0x01, 0x01, 0xFF}},
		{"uint16", uint(math.MaxUint16), []byte{TypeUInt16, 'u', 'i', 'n', 't', '1', '6', 0xFF, 0x01, 0x02, 0xFF, 0xFF}},
		{"uint32", uint(math.MaxUint32), []byte{TypeUInt32, 'u', 'i', 'n', 't', '3', '2', 0xFF, 0x01, 0x04, 0xFF, 0xFF, 0xFF, 0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := encodeArgument(buf, tt.value, tt.name)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resultBuf := bytes.NewBuffer(tt.result)
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

func TestEncodeMapStringKeyArgument(t *testing.T) {
	tests := []struct {
		name         string
		value        map[string]any
		outerResult  []byte
		innerResults [][]byte
	}{
		{"single", map[string]any{"moin": byte(0xDE)}, []byte{
			TypeMapStringKey, 's', 'i', 'n', 'g', 'l', 'e', 0xFF, 0x01,
		}, [][]byte{
			{
				TypeUInt8, 'm', 'o', 'i', 'n', 0xFF, 0x01, 0x01, 0xDE,
			},
		}},
		{"multiple", map[string]any{"moin": byte(0xDE), "dikka": byte(0x68)}, []byte{
			TypeMapStringKey, 'm', 'u', 'l', 't', 'i', 'p', 'l', 'e', 0xFF, 0x01,
		}, [][]byte{
			{
				TypeUInt8, 'd', 'i', 'k', 'k', 'a', 0xFF, 0x01, 0x01, 0x68,
			},
			{
				TypeUInt8, 'm', 'o', 'i', 'n', 0xFF, 0x01, 0x01, 0xDE,
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := encodeArgument(buf, tt.value, tt.name)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
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
			innerBufSize := innerBuf.Len()

			resultBuf := bytes.NewBuffer(tt.outerResult)
			resultBuf.WriteByte(byte(innerBufSize))
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

func TestEncodeNestedMapStringKeyArgument(t *testing.T) {
	innerMap1 := map[string]any{
		"key1": byte(0xDE),
	}
	innerMap2 := map[string]any{
		"key2": byte(0x68),
	}
	value := map[string]any{
		"outerKey1": innerMap1,
		"outerKey2": innerMap2,
	}

	innerMap1ContentExpected := bytes.NewBuffer([]byte{
		TypeUInt8, 'k', 'e', 'y', '1', 0xFF, 0x01, 0x01, 0xDE,
	})
	err := writeChecksum(innerMap1ContentExpected)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	innerMap2ContentExpected := bytes.NewBuffer([]byte{
		TypeUInt8, 'k', 'e', 'y', '2', 0xFF, 0x01, 0x01, 0x68,
	})
	err = writeChecksum(innerMap2ContentExpected)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	innerMap1Expected := bytes.NewBuffer([]byte{
		TypeMapStringKey, 'o', 'u', 't', 'e', 'r', 'K', 'e', 'y', '1', 0xFF, 0x01, byte(innerMap1ContentExpected.Len()),
	})
	innerMap1Expected.Write(innerMap1ContentExpected.Bytes())
	err = writeChecksum(innerMap1Expected)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	innerMap2Expected := bytes.NewBuffer([]byte{
		TypeMapStringKey, 'o', 'u', 't', 'e', 'r', 'K', 'e', 'y', '2', 0xFF, 0x01, byte(innerMap2ContentExpected.Len()),
	})
	innerMap2Expected.Write(innerMap2ContentExpected.Bytes())
	err = writeChecksum(innerMap2Expected)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	outerExpected := bytes.NewBuffer([]byte{
		TypeMapStringKey, 'n', 'e', 's', 't', 'e', 'd', 0xFF, 0x01,
	})
	outerExpected.WriteByte(byte(innerMap1Expected.Len() + innerMap2Expected.Len()))
	outerExpected.Write(innerMap1Expected.Bytes())
	outerExpected.Write(innerMap2Expected.Bytes())
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
diff
%s
		`, compareBytes(outerExpected.Bytes(), buf.Bytes()))
	}
}

func TestEncodeMap(t *testing.T) {
	tests := []struct {
		name         string
		value        map[any]any
		outerResult  []byte
		innerResults [][]byte
	}{
		{"single", map[any]any{byte(0xDE): "moin"}, []byte{
			TypeMap, 's', 'i', 'n', 'g', 'l', 'e', 0xFF, 0x01,
		}, [][]byte{
			{
				TypeUInt8, 0xFF, 0x01, 0x01, 0xDE,
			},
			{
				TypeString, 0xFF, 0x01, 0x04, 'm', 'o', 'i', 'n',
			},
		}},
		{"multiple", map[any]any{byte(0xDE): "moin", string("dikka"): byte(0x68)}, []byte{
			TypeMap, 'm', 'u', 'l', 't', 'i', 'p', 'l', 'e', 0xFF, 0x01,
		}, [][]byte{
			{
				TypeUInt8, 0xFF, 0x01, 0x01, 0xDE,
			},
			{
				TypeString, 0xFF, 0x01, 0x04, 'm', 'o', 'i', 'n',
			},
			{
				TypeString, 0xFF, 0x01, 0x05, 'd', 'i', 'k', 'k', 'a',
			},
			{
				TypeUInt8, 0xFF, 0x01, 0x01, 0x68,
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := encodeArgument(buf, tt.value, tt.name)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
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
			innerBufSize := innerBuf.Len()

			resultBuf := bytes.NewBuffer(tt.outerResult)
			resultBuf.WriteByte(byte(innerBufSize))
			resultBuf.Write(innerBuf.Bytes())

			err = writeChecksum(resultBuf)
			if err != nil {
				t.Fatalf("error writing checksum: %v", err)
			}

			if !bytes.Equal(buf.Bytes(), resultBuf.Bytes()) {
				t.Fatalf(`
diff:
%s
				`, compareBytes(resultBuf.Bytes(), buf.Bytes()))
			}
		})
	}
}

func TestEncodeNestedMap(t *testing.T) {
	innerMap := map[any]any{
		"innerKey": byte(0xDE),
	}
	outerMap := map[any]any{
		"outerKey": innerMap,
	}
	_ = outerMap

	innerMapContentExpectedKey := bytes.NewBuffer([]byte{
		TypeString, 0xFF, 0x01, 0x08, 'i', 'n', 'n', 'e', 'r', 'K', 'e', 'y',
	})
	err := writeChecksum(innerMapContentExpectedKey)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	innerMapContentExpectedValue := bytes.NewBuffer([]byte{
		TypeUInt8, 0xFF, 0x01, 0x01, 0xDE,
	})
	err = writeChecksum(innerMapContentExpectedValue)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	innerMapExpected := bytes.NewBuffer([]byte{
		TypeMap, 0xFF, 0x01, byte(innerMapContentExpectedKey.Len() + innerMapContentExpectedValue.Len()),
	})
	innerMapExpected.Write(innerMapContentExpectedKey.Bytes())
	innerMapExpected.Write(innerMapContentExpectedValue.Bytes())

	outerMapContentExpectedKey := bytes.NewBuffer([]byte{
		TypeString, 0xFF, 0x01, 0x08, 'o', 'u', 't', 'e', 'r', 'K', 'e', 'y',
	})
	err = writeChecksum(outerMapContentExpectedKey)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	outerMapContentExpectedValue := bytes.NewBuffer(innerMapExpected.Bytes())
	err = writeChecksum(outerMapContentExpectedValue)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	outerMapExpected := bytes.NewBuffer([]byte{
		TypeMap, 'n', 'e', 's', 't', 'e', 'd', 0xFF, 0x01, byte(outerMapContentExpectedKey.Len() + outerMapContentExpectedValue.Len()),
	})
	outerMapExpected.Write(outerMapContentExpectedKey.Bytes())
	outerMapExpected.Write(outerMapContentExpectedValue.Bytes())
	err = writeChecksum(outerMapExpected)
	if err != nil {
		t.Fatalf("error writing checksum: %v", err)
	}

	buf := bytes.NewBuffer(nil)
	err = encodeArgument(buf, outerMap, "nested")
	if err != nil {
		t.Fatalf("error encoding argument: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), outerMapExpected.Bytes()) {
		t.Fatalf(`
diff
%s
		`, compareBytes(outerMapExpected.Bytes(), buf.Bytes()))
	}
}
