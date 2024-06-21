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
	if isFixedType(typeTag) {
		content, err = encodeFixedPrimitiveContent(value)
		if err != nil {
			return err
		}
	} else {
		switch typeTag {
		case TypeString:
			content = []byte(value.(string))
		case TypeStruct:
			contentBuffer := bytes.NewBuffer(nil)
			t := reflect.ValueOf(value).Type()
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				fieldValue := reflect.ValueOf(value).Field(i).Interface()

				tmpBuf := bytes.NewBuffer(nil)

				err := encodeArgument(tmpBuf, fieldValue, field.Name)
				if err != nil {
					return err
				}

				_, err = contentBuffer.Write(tmpBuf.Bytes())
				if err != nil {
					return err
				}
			}
			content = contentBuffer.Bytes()
		case TypeSlice:
			contentBuffer := bytes.NewBuffer(nil)
			for i := 0; i < reflect.ValueOf(value).Len(); i++ {
				tmpBuf := bytes.NewBuffer(nil)

				element := reflect.ValueOf(value).Index(i).Interface()
				err := encodeArgument(tmpBuf, element, "")
				if err != nil {
					return err
				}

				_, err = contentBuffer.Write(tmpBuf.Bytes())
				if err != nil {
					return err
				}
			}
			content = contentBuffer.Bytes()
		default:
			return fmt.Errorf("Encoding for type %v not yet implemented", reflect.TypeOf(value))
		}
	}

	contentSizeData := len(content)
	contentSize := bytes.NewBuffer(nil)
	switch {
	case contentSizeData >= math.MinInt8 && contentSizeData <= math.MaxInt8:
		err = binary.Write(contentSize, binary.BigEndian, int8(contentSizeData))
	case contentSizeData >= math.MinInt16 && contentSizeData <= math.MaxInt16:
		err = binary.Write(contentSize, binary.BigEndian, int16(contentSizeData))
	case contentSizeData >= math.MinInt32 && contentSizeData <= math.MaxInt32:
		err = binary.Write(contentSize, binary.BigEndian, int32(contentSizeData))
	default:
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
	checksumBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(checksumBuffer, checksum)
	buffer.Write(checksumBuffer)
	return nil
}

func encodeFixedPrimitiveContent(value any) ([]byte, error) {
	resultWriter := bytes.NewBuffer(nil)
	binary.Write(resultWriter, binary.BigEndian, value)
	return resultWriter.Bytes(), nil
}

func decodeArgument(data []byte) (name string, value any, typ byte, err error) {
	withoutChecksum := data[:len(data)-4]
	checksum := data[len(data)-4:]
	ok := verifyChecksum(withoutChecksum, checksum)
	if !ok {
		err = fmt.Errorf("Checksum verification failed")
		return
	}

	buffer := bytes.NewBuffer(withoutChecksum)

	typ, err = buffer.ReadByte()
	if err != nil {
		return
	}

	name, err = readIdentifier(buffer)
	if err != nil {
		return
	}

	contentSizeDescriptor, err := buffer.ReadByte()
	if err != nil {
		return
	}
	sizeBytes := buffer.Next(int(contentSizeDescriptor))
	var size int
	switch len(sizeBytes) {
	case 1:
		size = int(int8(sizeBytes[0]))
	case 2:
		size = int(int16(binary.BigEndian.Uint16(sizeBytes)))
	case 4:
		size = int(int32(binary.BigEndian.Uint32(sizeBytes)))
	case 8:
		size = int(int64(binary.BigEndian.Uint64(sizeBytes)))
	default:
		err = fmt.Errorf("Invalid size descriptor: %v", sizeBytes)
	}

	content := buffer.Next(size)

	if isFixedType(typ) {
		value, err = decodeFixedPrimitiveContent(typ, content)
		if err != nil {
			return
		}
	} else {
		switch typ {
		case TypeString:
			value = string(content)
		case TypeStruct:
			var structField []reflect.StructField
			splitData, err := splitArgumentListData(content)
			if err != nil {
				return "", nil, 0, err
			}
			dataForSetting := make(map[string]any)
			for _, fieldData := range splitData {
				name, fieldValue, _, err := decodeArgument(fieldData)
				if err != nil {
					return "", nil, 0, err
				}
				dataForSetting[name] = fieldValue
				structField = append(structField, reflect.StructField{Name: name, Type: reflect.TypeOf(fieldValue)})
			}
			structType := reflect.StructOf(structField)
			instance := reflect.New(structType).Elem()
			for name, fieldValue := range dataForSetting {
				field := instance.FieldByName(name)
				if field.IsValid() && field.CanSet() {
					field.Set(reflect.ValueOf(fieldValue))
				}
			}
			value = instance.Interface()
		case TypeSlice:
			splitData, err := splitArgumentListData(content)
			if err != nil {
				return "", nil, 0, err
			}
			tmp := make([]any, 0)
			typesInSlice := make([]reflect.Type, 0)
			for _, fieldData := range splitData {
				_, fieldValue, _, err := decodeArgument(fieldData)
				if err != nil {
					return "", nil, 0, err
				}
				typesInSlice = append(typesInSlice, reflect.TypeOf(fieldValue))
				tmp = append(tmp, fieldValue)
			}
			if len(tmp) == 0 {
				value = tmp
				break
			}
			isAllSameType := true
			sameType := typesInSlice[0]
			for _, t := range typesInSlice {
				if t != sameType {
					isAllSameType = false
					break
				}
			}
			if isAllSameType {
				slice := reflect.MakeSlice(reflect.SliceOf(sameType), len(tmp), len(tmp))
				for i, v := range tmp {
					slice.Index(i).Set(reflect.ValueOf(v))
				}
				value = slice.Interface()
			} else {
				value = tmp
			}
		default:
			err = fmt.Errorf("Decoding for type '%s' not yet implemented", TypeToString[typ])
		}

	}

	return
}

func splitArgumentListData(data []byte) ([][]byte, error) {
	result := make([][]byte, 0)
	buffer := bytes.NewBuffer(data)

	for {
		if buffer.Len() == 0 {
			break
		}

		tmpBuffer := bytes.NewBuffer(nil)

		untilFF, err := buffer.ReadBytes(0xFF)
		if err != nil {
			return nil, err
		}
		tmpBuffer.Write(untilFF)

		contentSizeDescriptor, err := buffer.ReadByte()
		if err != nil {
			return nil, err
		}
		tmpBuffer.WriteByte(contentSizeDescriptor)

		sizeBytes := buffer.Next(int(contentSizeDescriptor))
		if len(sizeBytes) != int(contentSizeDescriptor) {
			return nil, fmt.Errorf("not enough bytes for size descriptor")
		}
		tmpBuffer.Write(sizeBytes)

		var size int
		switch len(sizeBytes) {
		case 1:
			size = int(int8(sizeBytes[0]))
		case 2:
			size = int(int16(binary.BigEndian.Uint16(sizeBytes)))
		case 4:
			size = int(int32(binary.BigEndian.Uint32(sizeBytes)))
		case 8:
			size = int(int64(binary.BigEndian.Uint64(sizeBytes)))
		default:
			return nil, fmt.Errorf("invalid size descriptor: %v", sizeBytes)
		}

		content := buffer.Next(size)
		tmpBuffer.Write(content)

		crc32Bytes := buffer.Next(4)
		if len(crc32Bytes) != 4 {
			return nil, fmt.Errorf("not enough bytes for CRC32")
		}
		tmpBuffer.Write(crc32Bytes)

		result = append(result, tmpBuffer.Bytes())
	}
	return result, nil
}

func decodeFixedPrimitiveContent(typ byte, content []byte) (any, error) {
	reader := bytes.NewReader(content)
	switch typ {
	case TypeBool:
		var result bool
		err := binary.Read(reader, binary.BigEndian, &result)
		return result, err
	case TypeUInt8:
		var result uint8
		err := binary.Read(reader, binary.BigEndian, &result)
		return result, err
	case TypeUInt16:
		var result uint16
		err := binary.Read(reader, binary.BigEndian, &result)
		return result, err
	case TypeUInt32:
		var result uint32
		err := binary.Read(reader, binary.BigEndian, &result)
		return result, err
	case TypeUInt64:
		var result uint64
		err := binary.Read(reader, binary.BigEndian, &result)
		return result, err
	case TypeInt8:
		var result int8
		err := binary.Read(reader, binary.BigEndian, &result)
		return result, err
	case TypeInt16:
		var result int16
		err := binary.Read(reader, binary.BigEndian, &result)
		return result, err
	case TypeInt32:
		var result int32
		err := binary.Read(reader, binary.BigEndian, &result)
		return result, err
	case TypeInt64:
		var result int64
		err := binary.Read(reader, binary.BigEndian, &result)
		return result, err
	case TypeFloat32:
		var result float32
		err := binary.Read(reader, binary.BigEndian, &result)
		return result, err
	case TypeFloat64:
		var result float64
		err := binary.Read(reader, binary.BigEndian, &result)
		return result, err
	case TypeComplex64:
		var result complex64
		err := binary.Read(reader, binary.BigEndian, &result)
		return result, err
	case TypeComplex128:
		var result complex128
		err := binary.Read(reader, binary.BigEndian, &result)
		return result, err
	default:
		return nil, fmt.Errorf("Unsupported type: %v", typ)
	}
}

func readIdentifier(buffer *bytes.Buffer) (string, error) {
	identifierBytes, err := buffer.ReadBytes(0xFF)
	return string(identifierBytes[:len(identifierBytes)-1]), err
}

func verifyChecksum(data []byte, checksum []byte) bool {
	return crc32.ChecksumIEEE(data) == binary.BigEndian.Uint32(checksum)
}
