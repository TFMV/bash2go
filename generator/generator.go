// generator.go
package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"sort"
	"strings"
)

// CodeBuilder helps in constructing code with proper indentation.
type CodeBuilder struct {
	buf         bytes.Buffer
	indentLevel int
	indentStr   string
}

// NewCodeBuilder initializes a new CodeBuilder using a tab for indentation.
func NewCodeBuilder() *CodeBuilder {
	return &CodeBuilder{indentStr: "\t"}
}

// WriteLine writes a single line of code with the current indentation.
func (cb *CodeBuilder) WriteLine(line string) {
	for i := 0; i < cb.indentLevel; i++ {
		cb.buf.WriteString(cb.indentStr)
	}
	cb.buf.WriteString(line)
	cb.buf.WriteString("\n")
}

// Write writes multiple lines of code.
func (cb *CodeBuilder) Write(lines ...string) {
	for _, line := range lines {
		cb.WriteLine(line)
	}
}

// Indent increases the indentation level.
func (cb *CodeBuilder) Indent() {
	cb.indentLevel++
}

// Outdent decreases the indentation level.
func (cb *CodeBuilder) Outdent() {
	if cb.indentLevel > 0 {
		cb.indentLevel--
	}
}

// String returns the accumulated source code.
func (cb *CodeBuilder) String() string {
	return cb.buf.String()
}

// Format returns the source code formatted according to Go standards.
func (cb *CodeBuilder) Format() (string, error) {
	src, err := format.Source(cb.buf.Bytes())
	if err != nil {
		// If formatting fails, return the unformatted source with error details.
		return cb.buf.String(), fmt.Errorf("failed to format generated code: %w", err)
	}
	return string(src), nil
}

// CodeGenerator is the enterprise-grade code generator.
type CodeGenerator struct {
	packageName string
	imports     map[string]bool
	globals     []string
	functions   []Function
	cb          *CodeBuilder
}

// NewCodeGenerator creates a new CodeGenerator for a given package name.
func NewCodeGenerator(packageName string) *CodeGenerator {
	return &CodeGenerator{
		packageName: packageName,
		imports:     make(map[string]bool),
		cb:          NewCodeBuilder(),
	}
}

// AddImport registers an import package, avoiding duplicates.
func (cg *CodeGenerator) AddImport(pkg string) {
	cg.imports[pkg] = true
}

// AddGlobal adds a global variable declaration or any global-level code.
func (cg *CodeGenerator) AddGlobal(global string) {
	cg.globals = append(cg.globals, global)
}

// Parameter represents a function parameter.
type Parameter struct {
	Name string // Parameter name.
	Type string // Parameter type.
}

// Function represents a Go function with its signature, body, and optional comments.
type Function struct {
	Name       string      // Function name.
	Parameters []Parameter // List of parameters.
	ReturnType string      // Return type (empty if none).
	Body       []string    // Lines of code in the function body.
	Comments   []string    // Optional comment lines to document the function.
}

// AddFunction registers a function to be included in the generated source.
func (cg *CodeGenerator) AddFunction(fn Function) {
	cg.functions = append(cg.functions, fn)
}

// Build constructs the complete Go source file and returns the formatted source code.
func (cg *CodeGenerator) Build() (string, error) {
	// Package declaration.
	cg.cb.WriteLine(fmt.Sprintf("package %s", cg.packageName))
	cg.cb.WriteLine("")

	// Imports block.
	if len(cg.imports) > 0 {
		cg.cb.WriteLine("import (")
		cg.cb.Indent()
		// Sort imports for consistency.
		importKeys := make([]string, 0, len(cg.imports))
		for imp := range cg.imports {
			importKeys = append(importKeys, imp)
		}
		sort.Strings(importKeys)
		for _, imp := range importKeys {
			cg.cb.WriteLine(fmt.Sprintf("\"%s\"", imp))
		}
		cg.cb.Outdent()
		cg.cb.WriteLine(")")
		cg.cb.WriteLine("")
	}

	// Global variables and declarations.
	for _, global := range cg.globals {
		cg.cb.WriteLine(global)
	}
	if len(cg.globals) > 0 {
		cg.cb.WriteLine("")
	}

	// Functions.
	for _, fn := range cg.functions {
		// Write any comments.
		for _, comment := range fn.Comments {
			cg.cb.WriteLine("// " + comment)
		}
		// Build parameter list.
		params := make([]string, len(fn.Parameters))
		for i, p := range fn.Parameters {
			params[i] = fmt.Sprintf("%s %s", p.Name, p.Type)
		}
		paramList := strings.Join(params, ", ")
		// Build function signature.
		signature := fmt.Sprintf("func %s(%s)", fn.Name, paramList)
		if fn.ReturnType != "" {
			signature += " " + fn.ReturnType
		}
		cg.cb.WriteLine(signature + " {")
		cg.cb.Indent()
		for _, line := range fn.Body {
			cg.cb.WriteLine(line)
		}
		cg.cb.Outdent()
		cg.cb.WriteLine("}")
		cg.cb.WriteLine("")
	}

	// Return the formatted source code.
	return cg.cb.Format()
}

// GenerateMain generates a simple main function
func GenerateMain() (string, error) {
	cb := NewCodeBuilder()

	cb.WriteLine("package main")
	cb.WriteLine("")
	cb.WriteLine("import (")
	cb.Indent()
	cb.WriteLine("\"fmt\"")
	cb.Outdent()
	cb.WriteLine(")")
	cb.WriteLine("")
	cb.WriteLine("func main() {")
	cb.Indent()
	cb.WriteLine("fmt.Println(\"Hello, World!\")")
	cb.Outdent()
	cb.WriteLine("}")

	return cb.Format()
}
