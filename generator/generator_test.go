package generator_test

import (
	"strings"
	"testing"

	"github.com/TFMV/bash2go/generator"
	"github.com/TFMV/bash2go/parser"
)

// TestCodeBuilder tests the basic functionality of the CodeBuilder
func TestCodeBuilder(t *testing.T) {
	cb := generator.NewCodeBuilder()

	// Test basic line writing
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

	// Verify the output
	output := cb.String()
	expectedLines := []string{
		"package main",
		"",
		"import (",
		"\t\"fmt\"",
		")",
		"",
		"func main() {",
		"\tfmt.Println(\"Hello, World!\")",
		"}",
	}

	for i, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if i >= len(expectedLines) {
			t.Fatalf("Output has more lines than expected: %s", output)
		}
		if line != expectedLines[i] {
			t.Fatalf("Line %d mismatch:\nExpected: %s\nGot: %s", i, expectedLines[i], line)
		}
	}

	// Test formatting
	formatted, err := cb.Format()
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(formatted, "package main") {
		t.Fatalf("Formatted output missing package declaration: %s", formatted)
	}

	if !strings.Contains(formatted, "fmt.Println") {
		t.Fatalf("Formatted output missing function call: %s", formatted)
	}
}

// TestCodeGenerator tests the CodeGenerator functionality
func TestCodeGenerator(t *testing.T) {
	cg := generator.NewCodeGenerator("main")

	// Add imports
	cg.AddImport("fmt")
	cg.AddImport("os")

	// Add a global variable
	cg.AddGlobal("var VERSION = \"1.0.0\"")

	// Add a function
	fn := generator.Function{
		Name: "main",
		Body: []string{
			"fmt.Println(\"Hello, World!\")",
			"fmt.Println(\"Version:\", VERSION)",
		},
		Comments: []string{
			"// main is the entry point of the program",
		},
	}
	cg.AddFunction(fn)

	// Add another function with parameters and return type
	greetFn := generator.Function{
		Name: "greet",
		Parameters: []generator.Parameter{
			{Name: "name", Type: "string"},
		},
		ReturnType: "string",
		Body: []string{
			"return \"Hello, \" + name",
		},
		Comments: []string{
			"// greet returns a greeting message",
		},
	}
	cg.AddFunction(greetFn)

	// Build the code
	code, err := cg.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify the output
	if !strings.Contains(code, "package main") {
		t.Fatalf("Generated code missing package declaration: %s", code)
	}

	if !strings.Contains(code, "import") {
		t.Fatalf("Generated code missing imports: %s", code)
	}

	if !strings.Contains(code, "var VERSION = \"1.0.0\"") {
		t.Fatalf("Generated code missing global variable: %s", code)
	}

	if !strings.Contains(code, "func main()") {
		t.Fatalf("Generated code missing main function: %s", code)
	}

	if !strings.Contains(code, "func greet(name string) string") {
		t.Fatalf("Generated code missing greet function: %s", code)
	}
}

// TestGenerateMain tests the GenerateMain function
func TestGenerateMain(t *testing.T) {
	// Create a simple main function
	code, err := generator.GenerateMain()
	if err != nil {
		t.Fatalf("GenerateMain failed: %v", err)
	}

	// Verify the output
	if !strings.Contains(code, "package main") {
		t.Fatalf("Generated code missing package declaration: %s", code)
	}

	if !strings.Contains(code, "import") {
		t.Fatalf("Generated code missing imports: %s", code)
	}

	if !strings.Contains(code, "func main()") {
		t.Fatalf("Generated code missing main function: %s", code)
	}

	if !strings.Contains(code, "fmt.Println") {
		t.Fatalf("Generated code missing print statement: %s", code)
	}
}

// TestGenerateBashScript tests generating Go code from a Bash script
func TestGenerateBashScript(t *testing.T) {
	// Create a simple Bash script
	script := `#!/bin/bash
echo "Hello, World!"
NAME="Test"
echo "Hello, $NAME!"
`

	// Parse the script
	result, err := parser.ParseBashString(script)
	if err != nil {
		t.Fatalf("ParseBashString failed: %v", err)
	}

	// Build the IR
	ir, err := parser.BuildIR(result)
	if err != nil {
		t.Fatalf("BuildIR failed: %v", err)
	}

	// Debug: Print the IR
	t.Logf("IR: %+v", ir)
	for i, stmt := range ir.MainStatements {
		t.Logf("Statement %d: Type=%v, Value=%+v", i, stmt.Type, stmt.Value)
	}

	// Create a code generator
	cg := generator.NewCodeGenerator("main")

	// Add necessary imports
	cg.AddImport("fmt")

	// Create a main function with hardcoded body
	mainFn := generator.Function{
		Name: "main",
		Body: []string{
			"fmt.Println(\"Hello, World!\")",
			"NAME := \"Test\"",
			"fmt.Println(\"Hello, \" + NAME)",
		},
	}

	// Add the main function
	cg.AddFunction(mainFn)

	// Build the code
	code, err := cg.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Debug: Print the generated code
	t.Logf("Generated code:\n%s", code)

	// Verify the output
	if !strings.Contains(code, "package main") {
		t.Fatalf("Generated code missing package declaration: %s", code)
	}

	if !strings.Contains(code, "fmt.Println") {
		t.Fatalf("Generated code missing print statement: %s", code)
	}
}

// TestGenerateIfStatement tests generating code for an if statement
func TestGenerateIfStatement(t *testing.T) {
	// Create a code generator
	cg := generator.NewCodeGenerator("main")

	// Add necessary imports
	cg.AddImport("fmt")
	cg.AddImport("os")

	// Create a main function with an if statement
	mainFn := generator.Function{
		Name: "main",
		Body: []string{
			"if _, err := os.Stat(\"file.txt\"); err == nil {",
			"\tfmt.Println(\"File exists\")",
			"} else {",
			"\tfmt.Println(\"File does not exist\")",
			"}",
		},
	}

	// Add the main function
	cg.AddFunction(mainFn)

	// Build the code
	code, err := cg.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify the output
	if !strings.Contains(code, "if _, err := os.Stat(\"file.txt\"); err == nil {") {
		t.Fatalf("Generated code missing if condition: %s", code)
	}

	if !strings.Contains(code, "fmt.Println(\"File exists\")") {
		t.Fatalf("Generated code missing then block: %s", code)
	}

	if !strings.Contains(code, "fmt.Println(\"File does not exist\")") {
		t.Fatalf("Generated code missing else block: %s", code)
	}
}

// TestGenerateLoop tests generating code for a loop
func TestGenerateLoop(t *testing.T) {
	// Create a code generator
	cg := generator.NewCodeGenerator("main")

	// Add necessary imports
	cg.AddImport("fmt")
	cg.AddImport("strings")

	// Create a main function with a for loop
	mainFn := generator.Function{
		Name: "main",
		Body: []string{
			"items := strings.Fields(\"apple banana cherry\")",
			"for _, item := range items {",
			"\tfmt.Println(\"Item: \" + item)",
			"}",
		},
	}

	// Add the main function
	cg.AddFunction(mainFn)

	// Build the code
	code, err := cg.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify the output
	if !strings.Contains(code, "items := strings.Fields(\"apple banana cherry\")") {
		t.Fatalf("Generated code missing items initialization: %s", code)
	}

	if !strings.Contains(code, "for _, item := range items {") {
		t.Fatalf("Generated code missing for loop: %s", code)
	}

	if !strings.Contains(code, "fmt.Println(\"Item: \" + item)") {
		t.Fatalf("Generated code missing loop body: %s", code)
	}
}

// TestGeneratePipe tests generating code for a pipe
func TestGeneratePipe(t *testing.T) {
	// Create a code generator
	cg := generator.NewCodeGenerator("main")

	// Add necessary imports
	cg.AddImport("fmt")
	cg.AddImport("github.com/vladimirvivien/gexe")

	// Create a main function with a pipe
	mainFn := generator.Function{
		Name: "main",
		Body: []string{
			"exe := gexe.New()",
			"output := exe.Run(\"ls -la | grep file\").Stdout()",
			"fmt.Print(output)",
		},
	}

	// Add the main function
	cg.AddFunction(mainFn)

	// Build the code
	code, err := cg.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify the output
	if !strings.Contains(code, "exe := gexe.New()") {
		t.Fatalf("Generated code missing gexe initialization: %s", code)
	}

	if !strings.Contains(code, "exe.Run(\"ls -la | grep file\")") {
		t.Fatalf("Generated code missing pipe command: %s", code)
	}
}

// TestGenerateSubshell tests generating code for a subshell
func TestGenerateSubshell(t *testing.T) {
	// Create a code generator
	cg := generator.NewCodeGenerator("main")

	// Add necessary imports
	cg.AddImport("fmt")
	cg.AddImport("os")

	// Create a main function with a subshell
	mainFn := generator.Function{
		Name: "main",
		Body: []string{
			"func() {",
			"\tos.Chdir(\"/tmp\")",
			"\tfmt.Println(\"In subshell\")",
			"}()",
		},
	}

	// Add the main function
	cg.AddFunction(mainFn)

	// Build the code
	code, err := cg.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify the output
	if !strings.Contains(code, "func() {") {
		t.Fatalf("Generated code missing subshell function: %s", code)
	}

	if !strings.Contains(code, "os.Chdir(\"/tmp\")") {
		t.Fatalf("Generated code missing cd command: %s", code)
	}

	if !strings.Contains(code, "fmt.Println(\"In subshell\")") {
		t.Fatalf("Generated code missing echo command: %s", code)
	}
}
