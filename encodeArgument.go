package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"sort"
)

func encodeArgument(writeBuf *bytes.Buffer, value any, name string) error {
	buf := bytes.NewBuffer(nil)

	typeTag, ok := AnyToTypeTag(value)
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
		contentBuffer := bytes.NewBuffer(nil)
		err = binary.Write(contentBuffer, binary.BigEndian, value)
		if err != nil {
			return err
		}
		content = contentBuffer.Bytes()
	} else {
		switch typeTag {
		case TypeInt:
			contentBuffer := bytes.NewBuffer(nil)
			as64 := int64(value.(int))
			err = binary.Write(contentBuffer, binary.BigEndian, as64)
			if err != nil {
				return err
			}
			content = contentBuffer.Bytes()
		case TypeUInt:
			contentBuffer := bytes.NewBuffer(nil)
			as64 := uint64(value.(uint))
			err = binary.Write(contentBuffer, binary.BigEndian, as64)
			if err != nil {
				return err
			}
			content = contentBuffer.Bytes()
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
		case TypeMapStringKey:
			contentBuffer := bytes.NewBuffer(nil)
			m := map[string]any(value.(map[string]any))
			keys := make([]string, 0, len(m))
			for key := range m {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				tmpBuf := bytes.NewBuffer(nil)

				err := encodeArgument(tmpBuf, m[key], key)
				if err != nil {
					return err
				}

				_, err = contentBuffer.Write(tmpBuf.Bytes())
				if err != nil {
					return err
				}
			}
			content = contentBuffer.Bytes()
		case TypeMap:
			contentBuffer := bytes.NewBuffer(nil)
			m := map[any]any(value.(map[any]any))
			keys := make([]any, 0, len(m))
			for key := range m {
				keys = append(keys, key)
			}
			sort.Slice(keys, func(i, j int) bool {
				return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
			})

			for _, key := range keys {
				tmpBuf := bytes.NewBuffer(nil)

				err := encodeArgument(tmpBuf, key, "")
				if err != nil {
					return err
				}

				_, err = contentBuffer.Write(tmpBuf.Bytes())
				if err != nil {
					return err
				}

				tmpBuf = bytes.NewBuffer(nil)

				err = encodeArgument(tmpBuf, m[key], "")
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
	_, shrunkenType := shrinkInt(contentSizeData)
	switch shrunkenType {
	case TypeInt8:
		err = binary.Write(contentSize, binary.BigEndian, int8(contentSizeData))
	case TypeInt16:
		err = binary.Write(contentSize, binary.BigEndian, int16(contentSizeData))
	case TypeInt32:
		err = binary.Write(contentSize, binary.BigEndian, int32(contentSizeData))
	case TypeInt64:
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

	_, err = writeBuf.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
