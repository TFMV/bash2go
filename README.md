# bash2go

A Bash-to-Go Transpiler that converts Bash scripts to idiomatic Go code.

## Overview

bash2go is a tool that translates Bash scripts into Go programs, producing compiled binaries that replace shell scripts. This provides better performance, portability, and maintainability than traditional shell scripts.

## Features

- Converts Bash scripts to idiomatic Go code
- Handles common Bash constructs:
  - Variable assignments and substitutions
  - Command execution
  - Control flow (if, for, while, until, case)
  - Functions
  - Pipes and redirections
  - Subshells and background tasks
- Generates standalone Go executables

## Installation

```bash
go install github.com/TFMV/bash2go@latest
```

## Usage

### Converting a Bash script to Go

```bash
bash2go convert script.sh -o script.go
```

### Building a Bash script directly to a binary

```bash
bash2go build script.sh -o script
```

## Examples

### Simple Hello World

**hello.sh**:

```bash
#!/bin/bash
echo "Hello, World!"
```

**Generated Go code**:

```go
package main

import (
    "fmt"
)

// Main function generated from Bash script
func main() {
    fmt.Println("Hello, World!")
}
```

### More Complex Example

**example.sh**:

```bash
#!/bin/bash
NAME="World"
echo "Hello, $NAME!"

if [ -f "file.txt" ]; then
    echo "File exists"
else
    echo "File does not exist"
fi

for i in 1 2 3; do
    echo "Item $i"
done
```

## Project Structure

- `cmd/`: Command-line interface
- `parser/`: Bash script parsing and AST building
- `generator/`: Go code generation
- `compiler/`: Go code compilation
- `examples/`: Example Bash scripts and their Go equivalents

## Development

### Building from source

```bash
git clone https://github.com/TFMV/bash2go.git
cd bash2go
go build
```

### Running tests

```bash
go test ./...
```

## Limitations

- Not all Bash features are supported yet
- Complex shell expansions may not translate perfectly
- External command execution relies on the `gexe` library

## License

[MIT License](LICENSE)
