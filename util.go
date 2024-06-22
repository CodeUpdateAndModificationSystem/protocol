package protocol

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strings"
)

func bytesToHexString(b []byte) string {
	var buf bytes.Buffer
	for _, v := range b {
		buf.WriteString(fmt.Sprintf("%#02X ", v))
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

func saveAsBin(data []byte, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func shrinkInt(input int) (any, byte) {
	switch {
	case input >= math.MinInt8 && input <= math.MaxInt8:
		return int8(input), TypeInt8
	case input >= math.MinInt16 && input <= math.MaxInt16:
		return int16(input), TypeInt16
	case input >= math.MinInt32 && input <= math.MaxInt32:
		return int32(input), TypeInt32
	default:
		return int64(input), TypeInt64
	}
}

func shrinkUint(input uint) (any, byte) {
	switch {
	case input <= math.MaxUint8:
		return uint8(input), TypeUInt8
	case input <= math.MaxUint16:
		return uint16(input), TypeUInt16
	case input <= math.MaxUint32:
		return uint32(input), TypeUInt32
	default:
		return uint64(input), TypeUInt64
	}
}
