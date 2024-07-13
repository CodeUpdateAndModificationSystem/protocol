package protocol

import (
	"bytes"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestEncodeFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		options  *options
		args     map[string]any
		expected []byte
	}{
		{
			name: "single primitive",
			options: &options{
				version:     1,
				subversion:  0,
				compression: false,
			},
			args: map[string]any{"int": 0xDE},
			expected: []byte{
				0x69, 0xDE, 0xDE, 0x69, 0xF0, 0x9F, 0x90, 0xBB,
				1, 0,
				0x00,
				0x00, 0x00, 0x00, 0x00, 0x00,
				's', 'i', 'n', 'g', 'l', 'e', ' ', 'p', 'r', 'i', 'm', 'i', 't', 'i', 'v', 'e', 0xFF,
			},
		},
		{
			name: "multiple primitives",
			options: &options{
				version:     1,
				subversion:  0,
				compression: false,
			},
			args: map[string]any{
				"int":  0xDE,
				"bool": true,
				"str":  "moin",
			},
			expected: []byte{
				0x69, 0xDE, 0xDE, 0x69, 0xF0, 0x9F, 0x90, 0xBB,
				1, 0,
				0x00,
				0x00, 0x00, 0x00, 0x00, 0x00,
				'm', 'u', 'l', 't', 'i', 'p', 'l', 'e', ' ', 'p', 'r', 'i', 'm', 'i', 't', 'i', 'v', 'e', 's', 0xFF,
			},
		},
		{
			name: "mixed with structs",
			options: &options{
				version:     1,
				subversion:  0,
				compression: false,
			},
			args: map[string]any{
				"string": "moin",
				"struct": struct {
					Something bool
				}{
					true,
				},
			},
			expected: []byte{
				0x69, 0xDE, 0xDE, 0x69, 0xF0, 0x9F, 0x90, 0xBB,
				1, 0,
				0x00,
				0x00, 0x00, 0x00, 0x00, 0x00,
				'm', 'i', 'x', 'e', 'd', ' ', 'w', 'i', 't', 'h', ' ', 's', 't', 'r', 'u', 'c', 't', 's', 0xFF,
			},
		},
		{
			name: "simple with compression",
			options: &options{
				version:     1,
				subversion:  0,
				compression: true,
			},
			args: map[string]any{
				"string": "moin",
				"int":    0xDE,
			},
			expected: []byte{
				0x69, 0xDE, 0xDE, 0x69, 0xF0, 0x9F, 0x90, 0xBB,
				1, 0,
				0x01,
				0x00, 0x00, 0x00, 0x00, 0x00,
				's', 'i', 'm', 'p', 'l', 'e', ' ', 'w', 'i', 't', 'h', ' ', 'c', 'o', 'm', 'p', 'r', 'e', 's', 's', 'i', 'o', 'n', 0xFF,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expectedContentBuffer := bytes.NewBuffer(nil)

			argKeys := make([]string, 0, len(test.args))
			for key := range test.args {
				argKeys = append(argKeys, key)
			}
			sort.Slice(argKeys, func(i, j int) bool {
				return argKeys[i] < argKeys[j]
			})

			for _, key := range argKeys {
				arg := test.args[key]
				err := encodeArgument(expectedContentBuffer, arg, key)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			var expectedContent []byte
			var err error
			if test.options.compression {
				expectedContent, err = compressBuffer(expectedContentBuffer)
				if err != nil {
					return
				}
			} else {
				expectedContent = expectedContentBuffer.Bytes()
			}

			expectedBuffer := bytes.NewBuffer(test.expected)
			expectedBuffer.Write(expectedContent)
			err = writeChecksum(expectedBuffer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result, err := EncodeFunctionCall(test.name, test.options, test.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Equal(result, expectedBuffer.Bytes()) {
				t.Fatalf(`
expected:
%s
got:
%s
				`, formatXXD(expectedBuffer.Bytes()), formatXXD(result))
			}

		})
	}
}

type decodeExpected struct {
	name string
	args map[string]Argument
}

func TestEncodeDecodeFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]any
		compress bool
		expected decodeExpected
	}{
		{
			name:     "single primitive",
			args:     map[string]any{"int": 0xDE},
			compress: false,
			expected: decodeExpected{
				name: "single primitive",
				args: map[string]Argument{"int": {Name: "int", Value: 0xDE, Typ: TypeInt}},
			},
		},
		{
			name:     "multiple",
			args:     map[string]any{"int": 0xDE, "str": "moin"},
			compress: false,
			expected: decodeExpected{
				name: "multiple",
				args: map[string]Argument{
					"int": {Name: "int", Value: 0xDE, Typ: TypeInt},
					"str": {Name: "str", Value: "moin", Typ: TypeString},
				},
			},
		},
		{
			name:     "mixed with structs",
			args:     map[string]any{"string": "moin", "struct": struct{ Something bool }{true}},
			compress: false,
			expected: decodeExpected{
				name: "mixed with structs",
				args: map[string]Argument{
					"string": {Name: "string", Value: "moin", Typ: TypeString},
					"struct": {Name: "struct", Value: struct{ Something bool }{true}, Typ: TypeStruct},
				},
			},
		},
		{
			name:     "simple with compression",
			args:     map[string]any{"string": "moin", "int": 0xDE},
			compress: true,
			expected: decodeExpected{
				name: "simple with compression",
				args: map[string]Argument{
					"string": {Name: "string", Value: "moin", Typ: TypeString},
					"int":    {Name: "int", Value: 0xDE, Typ: TypeInt},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := &options{
				version:     1,
				subversion:  0,
				compression: test.compress,
			}

			data, err := EncodeFunctionCall(test.name, options, test.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			name, args, err := DecodeFunctionCall(data, options)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if name != test.expected.name {
				t.Fatalf("expected name %q, got %q", test.expected.name, name)
			}

			if len(args) != len(test.expected.args) {
				t.Fatalf("expected %d arguments, got %d", len(test.expected.args), len(args))
			}

			for key, expectedArg := range test.expected.args {
				arg, ok := args[key]
				if !ok {
					t.Fatalf("expected argument %q not found", key)
				}

				if arg.Name != expectedArg.Name {
					t.Fatalf("expected argument name %q, got %q", expectedArg.Name, arg.Name)
				}

				if arg.Typ != expectedArg.Typ {
					t.Fatalf("expected argument type %q, got %q", TypeToString[expectedArg.Typ], TypeToString[arg.Typ])
				}

				if !reflect.DeepEqual(arg.Value, expectedArg.Value) {
					t.Fatalf("expected argument value %x (%T), got %x (%T)", expectedArg.Value, expectedArg.Value, arg.Value, arg.Value)
				}
			}
		})
	}
}

func TestCompressDecompressBuffer(t *testing.T) {
	data := bytes.NewBuffer(nil)
	for i := 0; i < 32; i++ {
		data.WriteByte(byte(rand.Intn(255)))
	}

	dataCopy := bytes.NewBuffer(data.Bytes())
	compressed, err := compressBuffer(data)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(compressed) == 0 {
		t.Fatalf("expected compressed data, got nothing")
	}

	decompressed, err := decompressBuffer(compressed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(decompressed.Bytes(), dataCopy.Bytes()) {
		t.Fatalf(`
expected:
%s
got:
%s
		`, formatXXD(dataCopy.Bytes()), formatXXD(decompressed.Bytes()))
	}

}
