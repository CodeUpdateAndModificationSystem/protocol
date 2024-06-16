# Protocol Specification

## Overview

The protocol is designed to facilitate communication between a server and a client by encoding and decoding function calls and their arguments. The protocol ensures data integrity through the use of CRC32 checksums and supports various data types, including primitives, arrays, slices, structs, and maps. The encoding is binary, using big-endian format for consistency and intuitiveness.

## Structure of Encoded Messages

1. **Versioning**:
    
    - **Version (1 byte)**: Major version number, used for breaking changes.
    - **Subversion (1 byte)**: Minor version number, used for non-breaking changes.
2. **Function Identifier**:
    
    - **Function Identifier (variable length, 0xFF-terminated)**: The name of the function being called.
3. **Arguments**: Each argument is encoded with the following structure:
    
    - **Type Tag (1 byte)**: Indicates the type of the argument (e.g., integer, string, struct, array, map).
    - **Argument Name (variable length, 0xFF-terminated)**: The name of the argument.
    - **Size Descriptor Length (1 byte)**: The number of bytes used to describe the size of the argument content.
    - **Size (variable length)**: The actual size of the argument content, encoded in the number of bytes specified by the Size Descriptor Length.
    - **Content (variable length)**: The actual data of the argument, which can be recursively encoded for complex types.
    - **Checksum (4 bytes)**: A CRC32 checksum of the entire argument (type tag, name, size descriptor, size, and content) for data integrity.
4. **Overall Message Checksum**:
    
    - **Overall Checksum (4 bytes)**: A CRC32 checksum of the entire message, excluding this checksum.

## Argument Types

- **Primitive Types**: (e.g., integers, floats, strings)
    
    - Type tags indicate the type.
    - Size descriptor specifies the length of the content for variable-length types like strings.
- **Structs**:
    
    - Each field of the struct is encoded as a nested argument within the argument content.
    - The argument content is a list of arguments representing the fields of the struct.
- **Arrays/Slices**:
    
    - Each element is encoded as an argument with an empty name.
    - The argument content is a list of arguments representing the elements of the array.
- **Maps**:
    
    - Each key-value pair is encoded as separate arguments within the map content.
    - The argument content is a list of arguments representing the key-value pairs, interpreted in pairs (key followed by value).

## Example Encoding Format

```
| Version (1 byte) | Subversion (1 byte) | Function Identifier (variable length, 0xFF-terminated) |
| Argument Type (1 byte) | Argument Name (variable length, 0xFF-terminated) | Size Descriptor Length (1 byte) | Size (variable length) | Content (variable length) | Checksum (4 bytes) |
... 
| Overall Checksum (4 bytes) |
```

# Implementation Outline

1. **Protocol Package Functions**:
    - `Encode(version byte, subversion byte, identifier string, args ...interface{}) ([]byte, error)`: Encodes the function call.
    - `Decode(data []byte) (version byte, subversion byte, identifier string, args []interface{}, err error)`: Decodes the function call.
    - `ExtractName(data []byte) (string, error)`: Extracts the function identifier.
    - `DecodeArguments(data []byte) (args []interface{}, err error)`: Decodes the arguments.
    - `ComputeChecksum(data []byte) ([]byte, error)`: Computes a checksum for the given data.
    - `VerifyChecksum(data, checksum []byte) (bool, error)`: Verifies the checksum for the given data.
