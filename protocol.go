package protocol

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"sort"
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

func EncodeFunctionCall(name string, options *options, args map[string]any) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	_, err := buf.Write(signature)
	if err != nil {
		return nil, err
	}

	err = buf.WriteByte(options.version)
	if err != nil {
		return nil, err
	}
	err = buf.WriteByte(options.subversion)
	if err != nil {
		return nil, err
	}
	if options.compression {
		err = buf.WriteByte(1)
	} else {
		err = buf.WriteByte(0)
	}
	if err != nil {
		return nil, err
	}

	_, err = buf.Write([]byte{0, 0, 0, 0, 0})
	if err != nil {
		return nil, err
	}

	err = writeIdentifier(buf, name)
	if err != nil {
		return nil, err
	}

	argKeys := make([]string, 0, len(args))
	for key := range args {
		argKeys = append(argKeys, key)
	}
	sort.Slice(argKeys, func(i, j int) bool {
		return argKeys[i] < argKeys[j]
	})

	argsBuffer := bytes.NewBuffer(nil)
	for _, key := range argKeys {
		arg := args[key]
		err := encodeArgument(argsBuffer, arg, key)
		if err != nil {
			return nil, err
		}

	}

	var content []byte
	if options.compression {
		content, err = compressBuffer(argsBuffer)
		if err != nil {
			return nil, err
		}
	} else {
		content = argsBuffer.Bytes()
	}

	_, err = buf.Write(content)
	if err != nil {
		return nil, err
	}

	err = writeChecksum(buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func compressBuffer(buffer *bytes.Buffer) ([]byte, error) {
	compressedBuffer := bytes.NewBuffer(nil)
	writer, err := gzip.NewWriterLevel(compressedBuffer, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(writer, buffer)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return compressedBuffer.Bytes(), nil
}

type Argument struct {
	Name  string
	Value any
	Typ   byte
}

func DecodeFunctionCall(data []byte, options *options) (string, map[string]Argument, error) {
	buf := bytes.NewBuffer(data)

	signature := buf.Next(8)
	if !bytes.Equal(signature, Signature()) {
		return "", nil, fmt.Errorf("invalid signature")
	}

	version, err := buf.ReadByte()
	if err != nil {
		return "", nil, err
	}
	subversion, err := buf.ReadByte()
	if err != nil {
		return "", nil, err
	}

	if version != options.version {
		return "", nil, fmt.Errorf("Invalid version")
	}

	invalidSubversion := false
	if subversion != options.subversion {
		invalidSubversion = true
	}

	compression, err := buf.ReadByte()
	if err != nil {
		return "", nil, err
	}
	useCompression := compression == 1

	buf.Next(5)

	name, err := readIdentifier(buf)
	if err != nil {
		return "", nil, err
	}

	argData := buf.Next(buf.Len() - 4)
	checksum := buf.Next(4)
	checkedData := data[:len(data)-4]
	if !verifyChecksum(checkedData, checksum) {
		return "", nil, fmt.Errorf("FunctionCalls checksum verification failed")
	}

	argBuffer := bytes.NewBuffer(nil)
	if useCompression {
		argBuffer, err = decompressBuffer(argData)
		if err != nil {
			return "", nil, err
		}
	} else {
		argBuffer.Write(argData)
	}

	args := make(map[string]Argument)
	splitData, err := splitArgumentListData(argBuffer.Bytes())
	if err != nil {
		return "", nil, err
	}
	for _, data := range splitData {
		name, value, typ, err := decodeArgument(data)
		if err != nil {
			return "", nil, err
		}
		args[name] = Argument{
			Name:  name,
			Value: value,
			Typ:   typ,
		}
	}

	if invalidSubversion {
		err = &NonMatchingSubversionError{
			Expected: options.subversion,
			Actual:   subversion,
		}
	} else {
		err = nil
	}

	return name, args, err
}

func decompressBuffer(buffer []byte) (*bytes.Buffer, error) {
	reader, err := gzip.NewReader(bytes.NewReader(buffer))
	if err != nil {
		return nil, err
	}
	decompressedBuffer := bytes.NewBuffer(nil)
	_, err = io.Copy(decompressedBuffer, reader)
	if err != nil {
		return nil, err
	}
	return decompressedBuffer, nil
}
