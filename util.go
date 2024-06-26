package protocol

import (
	"bytes"
	"fmt"
	"math"
	"os"
)

func bytesToHexString(b []byte) string {
	var buf bytes.Buffer
	for _, v := range b {
		buf.WriteString(fmt.Sprintf("%#02X ", v))
	}
	return buf.String()
}

func formatXXD(data []byte) string {
	var result string
	for i := 0; i < len(data); i += 16 {
		result += fmt.Sprintf("%08x: ", i)
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				result += formatHex(data[i+j])
			} else {
				result += "   "
			}
		}
		result += " "
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				result += formatAsciiColored(data[i+j], false)
			} else {
				result += " "
			}
		}
		result += "\n"
	}
	return result
}

func compareBytes(expected, got []byte) string {
	var buffer bytes.Buffer
	chunkSize := 16

	// Compare data1
	for i := 0; i < len(expected); i += chunkSize {
		end := i + chunkSize
		if end > len(expected) {
			end = len(expected)
		}

		chunk := expected[i:end]
		hex := ""
		ascii := ""

		for j := 0; j < chunkSize; j++ {
			if j < len(chunk) {
				hex += formatHex(chunk[j])
				ascii += formatAsciiColored(chunk[j], false)
			} else {
				hex += "   "
				ascii += " "
			}
		}

		buffer.WriteString(fmt.Sprintf("%08x: %s %s\n", i, hex, ascii))
	}

	buffer.WriteString("\n")

	// Compare data2
	for i := 0; i < len(got); i += chunkSize {
		end := i + chunkSize
		if end > len(got) {
			end = len(got)
		}

		chunk := got[i:end]
		hex := ""
		ascii := ""

		for j := 0; j < chunkSize; j++ {
			if j < len(chunk) {
				if i+j >= len(expected) || expected[i+j] != chunk[j] {
					hex += fmt.Sprintf("\033[1;31m%02x\033[0m ", chunk[j])
					ascii += formatAsciiColored(chunk[j], true)
				} else {
					hex += formatHex(chunk[j])
					ascii += formatAsciiColored(chunk[j], false)
				}
			} else {
				hex += "   "
				ascii += " "
			}
		}

		buffer.WriteString(fmt.Sprintf("%08x: %s %s\n", i, hex, ascii))
	}

	return buffer.String()
}

func formatHex(b byte) string {
	switch b {
	case 0xff, 0x00:
		return fmt.Sprintf("\033[1;37m%02x\033[0m ", b) // White
	default:
		return fmt.Sprintf("\033[1;90m%02x\033[0m ", b) // Gray
	}
}

func formatAscii(b byte) string {
	if 32 <= b && b < 127 {
		return fmt.Sprintf("%c", b)
	}
	return "."
}

func formatAsciiColored(b byte, isDifferent bool) string {
	color := "\033[1;90m" // Gray
	if b == 0xff || b == 0x00 {
		color = "\033[1;37m" // White
	} else if isDifferent {
		color = "\033[1;31m" // Red
	}
	return fmt.Sprintf("%s%s\033[0m", color, formatAscii(b))
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
