package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math"
	"reflect"
)

type UnsupportedTypeError struct {
	Kind reflect.Kind
}

func (e *UnsupportedTypeError) Error() string {
	return fmt.Sprintf("unsupported type: %v", e.Kind)
}

type EncodingError struct {
	err error
}

func (e *EncodingError) Error() string {
	return fmt.Sprintf("error while encoding: %v", e.err)
}

const (
	TypeBool byte = iota + 1
	TypeUInt8
	TypeUInt16
	TypeUInt32
	TypeUInt64
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt64
	TypeFloat32
	TypeFloat64
	TypeComplex64
	TypeComplex128

	TypeString

	TypeStruct
	TypeSlice
	TypeMap
)

var TypeToTag = map[reflect.Kind]byte{
	reflect.Bool:       TypeBool,
	reflect.Uint8:      TypeUInt8,
	reflect.Uint16:     TypeUInt16,
	reflect.Uint32:     TypeUInt32,
	reflect.Uint64:     TypeUInt64,
	reflect.Int8:       TypeInt8,
	reflect.Int16:      TypeInt16,
	reflect.Int32:      TypeInt32,
	reflect.Int64:      TypeInt64,
	reflect.Float32:    TypeFloat32,
	reflect.Float64:    TypeFloat64,
	reflect.Complex64:  TypeComplex64,
	reflect.Complex128: TypeComplex128,
	reflect.String:     TypeString,
	reflect.Struct:     TypeStruct,
	reflect.Slice:      TypeSlice,
	reflect.Map:        TypeMap,
}

func isFixedType(typeTag byte) bool {
	return typeTag >= TypeBool && typeTag <= TypeComplex128
}

func writeIdentifier(buf *bytes.Buffer, name string) error {
	_, err := buf.Write([]byte(name))
	if err != nil {
		return err
	}
	err = buf.WriteByte(0xFF)
	if err != nil {
		return err
	}
	return nil
}

func encodeArgument(buf *bytes.Buffer, value any, name string) error {
	typeTag, ok := TypeToTag[reflect.TypeOf(value).Kind()]
	if !ok {
		return &UnsupportedTypeError{Kind: reflect.TypeOf(value).Kind()}
	}

	err := buf.WriteByte(typeTag)
	if err != nil {
		return err
	}

	err = writeIdentifier(buf, name)
	if err != nil {
		return err
	}

	var content []byte
	switch {
	case isFixedType(typeTag):
		content, err = encodeFixedPrimitiveContent(value)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("type handling not implemented")
	}

	contentSizeData := len(content)
	contentSize := bytes.NewBuffer(nil)
	switch { // shrinking the int to a fixed type
	case contentSizeData >= math.MinInt8 && contentSizeData <= math.MaxInt8:
		err = binary.Write(contentSize, binary.BigEndian, int8(contentSizeData))
	case contentSizeData >= math.MinInt16 && contentSizeData <= math.MaxInt16:
		err = binary.Write(contentSize, binary.BigEndian, int16(contentSizeData))
	case contentSizeData >= math.MinInt32 && contentSizeData <= math.MaxInt32:
		err = binary.Write(contentSize, binary.BigEndian, int32(contentSizeData))
	case contentSizeData >= math.MinInt64 && contentSizeData <= math.MaxInt64:
		err = binary.Write(contentSize, binary.BigEndian, int64(contentSizeData))
	}
	contentSizeDescriptor := byte(contentSize.Len())
	if err != nil {
		return err
	}

	err = buf.WriteByte(contentSizeDescriptor)
	if err != nil {
		return err
	}

	_, err = buf.Write(contentSize.Bytes())
	if err != nil {
		return err
	}

	_, err = buf.Write(content)
	if err != nil {
		return err
	}

	err = writeChecksum(buf)
	if err != nil {
		return err
	}

	return nil
}

func writeChecksum(buffer *bytes.Buffer) error {
	checksum := crc32.ChecksumIEEE(buffer.Bytes())
	return binary.Write(buffer, binary.BigEndian, checksum)
}

func encodeFixedPrimitiveContent(value any) ([]byte, error) {
	resultWriter := bytes.NewBuffer(nil)
	binary.Write(resultWriter, binary.BigEndian, value)
	return resultWriter.Bytes(), nil
}