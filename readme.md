# Protocol Specification

## Overview

This protocol facilitates communication between a server and a client by encoding and decoding function calls and their arguments. It ensures data integrity through CRC32 checksums and supports various data types, including primitives, arrays, slices, structs, and maps. There is also an option to use gzip compression on the argument content (gzip usage is usually only recommended for data above ~100 bytes). The encoding is binary, using big-endian format for consistency.

## Structure of Encoded Messages

1. **Header**:
    - **Magic Number/Signature (8 bytes)**: Fixed sequence of bytes to identify the protocol. `69DE DE69 F09F 90BB`
    - **Version (1 byte)**: Major version number, indicating breaking changes.
    - **Subversion (1 byte)**: Minor version number, indicating non-breaking changes.
    - **Compression Flag (1 byte)**: Indicates whether the message is compressed (0x01) or not (0x00).
    - **RESERVED (5 bytes)**: 5 bytes reserved for future use.
2. **Function Identifier**:
    - **Function Identifier (variable length, 0xFF-terminated)**: Null-terminated string representing the function name.
3. **Arguments**: Each argument is encoded with the following structure:
    - **Type Tag (1 byte)**: Indicates the argument type (e.g., integer, string, struct, array, map).
    - **Argument Name (variable length, 0xFF-terminated)**: Null-terminated string representing the argument name.
    - **Size Descriptor Length (1 byte)**: Number of bytes used to describe the size of the argument content.
    - **Size (variable length)**: Size of the argument content, encoded in the number of bytes specified by the Size Descriptor Length.
    - **Content (variable length)**: Actual data of the argument, recursively encoded for complex types.
    - **Checksum (4 bytes)**: CRC32 checksum of the entire argument (type tag, name, size descriptor, size, and content) for data integrity.
4. **Overall Message Checksum**:
    - **Overall Checksum (4 bytes)**: CRC32 checksum of the entire message, excluding the overall checksum itself.

## Argument Types

- **Primitive Types** (e.g., integers, floats, strings):
    - Type tags indicate the type.
    - Size descriptor specifies the length of the content for variable-length types like strings.
- **Structs**:
    - Each field of the struct is encoded as a nested argument within the argument content.
    - Argument content is a list of arguments representing the fields of the struct.
- **Arrays/Slices**:
    - Each element is encoded as an argument with an empty name.
    - Argument content is a list of arguments representing the elements of the array or slice.
- **Maps**:
    - Each key-value pair is encoded as separate arguments within the map content.
    - Argument content is a list of arguments representing the key-value pairs, interpreted in pairs (key followed by value).

## Example Encoding Format

```
| Version (1 byte) | Subversion (1 byte) | Function Identifier (variable length, 0xFF-terminated) |
| Argument Type (1 byte) | Argument Name (variable length, 0xFF-terminated) | Size Descriptor Length (1 byte) | Size (variable length) | Content (variable length) | Checksum (4 bytes) |
... 
| Overall Checksum (4 bytes) |
```

## Documentation

### Function Encoding

The `EncodeFunctionCall` function is used to encode a function call with its name and arguments into a binary format suitable for transmission. 

#### Parameters:
- `name` (string): The name of the function to be called.
- `options` (*options): Encoding options including version, subversion, and compression flag.
- `args` (map[string]any): A map of arguments where the key is the argument name and the value is the argument value.

#### Returns:
- `[]byte`: The encoded function call as a byte slice.
- `error`: An error object if encoding fails.

#### Example:
```go
options := &options{
    version: 1,
    subversion: 0,
    compression: false,
}
args := map[string]any{
    "arg1": 42,
    "arg2": "example",
}
encodedCall, err := EncodeFunctionCall("MyFunction", options, args)
if err != nil {
    log.Fatalf("Encoding failed: %v", err)
}
```

### Function Decoding

The `DecodeFunctionCall` function is used to decode a received binary message back into a function name and its arguments.

#### Parameters:

- `data` (\[\]byte): The binary data representing the encoded function call.
- `options` (\*options): Decoding options including expected version and subversion.

#### Returns:

- `string`: The name of the decoded function.
- `map[string]Argument`: A map of decoded arguments.
- `error`: An error object if decoding fails.

#### Example:

```go
options := &options{
    version: 1,
    subversion: 0,
}
functionName, args, err := DecodeFunctionCall(encodedData, options)
if err != nil {
    log.Fatalf("Decoding failed: %v", err)
}
fmt.Printf("Function: %s\n", functionName)
for name, arg := range args {
    fmt.Printf("Argument: %s, Value: %v, Type: %d\n", name, arg.Value, arg.Typ)
}
```
