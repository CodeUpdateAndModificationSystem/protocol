package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math"
	"reflect"
)

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

func writeChecksum(buffer *bytes.Buffer) error {
	checksum := crc32.ChecksumIEEE(buffer.Bytes())
	checksumBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(checksumBuffer, checksum)
	buffer.Write(checksumBuffer)
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
		contentBuffer := bytes.NewBuffer(nil)
		err = binary.Write(contentBuffer, binary.BigEndian, value)
		if err != nil {
			return err
		}
		content = contentBuffer.Bytes()
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
