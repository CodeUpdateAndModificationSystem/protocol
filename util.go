package protocol

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

func bytesToHexString(b []byte) string {
	var buf bytes.Buffer
	for _, v := range b {
		buf.WriteString(fmt.Sprintf("%02X ", v))
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
