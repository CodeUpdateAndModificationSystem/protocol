package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"reflect"
)

func decodeArgument(data []byte) (name string, value any, typ byte, err error) {
	withoutChecksum := data[:len(data)-4]
	checksum := data[len(data)-4:]
	ok := verifyChecksum(withoutChecksum, checksum)
	if !ok {
		err = fmt.Errorf("Arguments checksum verification failed")
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
		case TypeMapStringKey:
			splitData, err := splitArgumentListData(content)
			if err != nil {
				return "", nil, 0, err
			}
			tmp := make(map[string]any)
			for _, fieldData := range splitData {
				name, fieldValue, _, err := decodeArgument(fieldData)
				if err != nil {
					return "", nil, 0, err
				}
				tmp[name] = fieldValue
			}
			value = tmp
		case TypeMap:
			splitData, err := splitArgumentListData(content)
			if err != nil {
				return "", nil, 0, err
			}
			tmp := make(map[any]any)
			for i := 0; i < len(splitData); i += 2 {
				keyFieldData := splitData[i]
				valueFieldData := splitData[i+1]

				_, keyFieldValue, _, err := decodeArgument(keyFieldData)
				if err != nil {
					return "", nil, 0, err
				}

				_, valueFieldValue, _, err := decodeArgument(valueFieldData)
				if err != nil {
					return "", nil, 0, err
				}

				tmp[keyFieldValue] = valueFieldValue
			}
			value = tmp
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
