# bash2go

A Bash-to-Go Transpiler that converts Bash scripts to idiomatic Go code.

## Overview

bash2go is a tool that translates Bash scripts into Go programs, producing a compiled binary that replaces shell scripts. This provides better performance, portability, and maintainability than traditional shell scripts.

## Design Goals

- **Accurate Bash Translation**: Parse Bash syntax and convert it into idiomatic Go
- **Go-Idiomatic Execution**: Use proper Go packages instead of raw `exec.Command` when possible
- **Performance Optimized**: Compiled Go binaries execute significantly faster than interpreted shell scripts
- **Portable & Secure**: No dependencies on Unix utilities; works across Linux, macOS, and Windows
- **Extensible**: Support multiple levels of Bash features, from basic commands to complex control flow
- **CLI-Driven**: Provide an intuitive CLI to convert scripts and optionally compile them

## Features

bash2go supports a range of Bash features including:

- Basic command execution
- Variable assignment and substitution
- Function definitions and calls
- Conditional statements (if, case)
- Loops (for, while, until)
- Pipes between commands
- Subshells with isolated environment
- Command substitution

## Installation

```bash
# Clone the repository
git clone https://github.com/TFMV/bash2go.git
cd bash2go

# Build the binary
go build -o bash2go
```

## Usage

### Converting a Bash script to Go source code

```bash
bash2go convert script.sh -o output.go
```

### Converting and compiling to a binary

```bash
bash2go build script.sh -o mybinary
```

### Running the compiled binary

```bash
./mybinary
```

## Examples

The `examples` directory contains sample Bash scripts that demonstrate various features:

- `simple.sh`: Basic Bash script with variables, functions, loops, and conditionals
- `advanced.sh`: More complex script with pipes, subshells, and advanced features

## Project Structure

```
bash2go/
│── cmd/                # CLI entry point
│   ├── root.go         # Handles CLI commands
│── parser/             # Handles Bash parsing & AST conversion
│   ├── parse.go        # Parses Bash using mvdan/sh
│   ├── ast.go          # Traverses AST & extracts components
│── generator/          # Handles Go code generation
│   ├── template.go     # Defines Go code templates
│   ├── transpiler.go   # Converts AST to Go code
│── compiler/           # Handles Go code compilation
│   ├── build.go        # Runs go build
│── examples/           # Sample Bash scripts & their Go output
│── main.go             # CLI entrypoint
│── go.mod              # Go module definition
```

## Dependencies

- [mvdan.cc/sh/v3](https://github.com/mvdan/sh): Bash parser
- [github.com/spf13/cobra](https://github.com/spf13/cobra): CLI framework

## License

This project is licensed under the MIT License - see the LICENSE file for details.
