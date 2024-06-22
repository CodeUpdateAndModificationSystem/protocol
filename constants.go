package protocol

import (
	"fmt"
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

type DecodingError struct {
	err error
}

func (e *DecodingError) Error() string {
	return fmt.Sprintf("error while decoding: %v", e.err)
}

type NonMatchingSubversionError struct {
	Expected byte
	Actual   byte
}

func (e *NonMatchingSubversionError) Error() string {
	return fmt.Sprintf("subversion mismatch: expected %v, got %v", e.Expected, e.Actual)
}

var signature = []byte{0x69, 0xDE, 0xDE, 0x69, 0xF0, 0x9F, 0x90, 0xBB}

func Signature() []byte {
	return signature
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
	reflect.Array:      TypeSlice,
	reflect.Map:        TypeMap,
}
var TypeToString = map[byte]string{
	TypeBool:       "bool",
	TypeUInt8:      "uint8",
	TypeUInt16:     "uint16",
	TypeUInt32:     "uint32",
	TypeUInt64:     "uint64",
	TypeInt8:       "int8",
	TypeInt16:      "int16",
	TypeInt32:      "int32",
	TypeInt64:      "int64",
	TypeFloat32:    "float32",
	TypeFloat64:    "float64",
	TypeComplex64:  "complex64",
	TypeComplex128: "complex128",
	TypeString:     "string",
	TypeStruct:     "struct",
	TypeSlice:      "slice",
	TypeMap:        "map",
}

func isFixedType(typeTag byte) bool {
	return typeTag >= TypeBool && typeTag <= TypeComplex128
}
