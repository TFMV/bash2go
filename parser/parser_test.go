package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mvdan.cc/sh/v3/syntax"
)

// TestParseBashScript tests the ParseBashScript function
func TestParseBashScript(t *testing.T) {
	// Create a temporary test script
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "test.sh")

	script := `#!/bin/bash
# Test script
echo "Hello, World!"
NAME="Test"
echo "Hello, $NAME!"
`

	err := os.WriteFile(scriptPath, []byte(script), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Parse the script
	result, err := ParseBashScript(scriptPath)
	if err != nil {
		t.Fatalf("ParseBashScript failed: %v", err)
	}

	// Verify the result
	if result == nil {
		t.Fatal("ParseBashScript returned nil result")
	}

	if result.File == nil {
		t.Fatal("ParseBashScript returned nil File")
	}
}

// TestParseBashString tests the ParseBashString function
func TestParseBashString(t *testing.T) {
	script := `#!/bin/bash
echo "Hello, World!"
NAME="Test"
echo "Hello, $NAME!"
`

	// Parse the script
	result, err := ParseBashString(script)
	if err != nil {
		t.Fatalf("ParseBashString failed: %v", err)
	}

	// Verify the result
	if result == nil {
		t.Fatal("ParseBashString returned nil result")
	}

	if result.File == nil {
		t.Fatal("ParseBashString returned nil File")
	}
}

// TestBuildIR tests the BuildIR function
func TestBuildIR(t *testing.T) {
	script := `#!/bin/bash
echo "Hello, World!"
NAME="Test"
echo "Hello, $NAME!"
`

	// Parse the script
	result, err := ParseBashString(script)
	if err != nil {
		t.Fatalf("ParseBashString failed: %v", err)
	}

	// Build the IR
	ir, err := BuildIR(result)
	if err != nil {
		t.Fatalf("BuildIR failed: %v", err)
	}

	// Verify the IR
	if ir == nil {
		t.Fatal("BuildIR returned nil IR")
	}

	// Check that we have the expected statements
	if len(ir.MainStatements) == 0 {
		t.Fatal("BuildIR returned IR with no statements")
	}

	// Check that we have the required packages
	if _, ok := ir.RequiredPackages["fmt"]; !ok {
		t.Fatal("BuildIR returned IR without fmt package")
	}

	if _, ok := ir.RequiredPackages["os"]; !ok {
		t.Fatal("BuildIR returned IR without os package")
	}
}

// TestProcessCallExpr tests the processCallExpr function
func TestProcessCallExpr(t *testing.T) {
	script := `echo "Hello, World!"`

	// Parse the script
	result, err := ParseBashString(script)
	if err != nil {
		t.Fatalf("ParseBashString failed: %v", err)
	}

	// Get the CallExpr from the parsed script
	var callExpr *syntax.CallExpr
	for _, stmt := range result.File.Stmts {
		if stmt.Cmd != nil {
			if call, ok := stmt.Cmd.(*syntax.CallExpr); ok {
				callExpr = call
				break
			}
		}
	}

	if callExpr == nil {
		t.Fatal("Failed to find CallExpr in parsed script")
	}

	// Process the CallExpr
	cmd := processCallExpr(callExpr)

	// Verify the command
	if cmd.Name != "echo" {
		t.Fatalf("Expected command name 'echo', got '%s'", cmd.Name)
	}

	if len(cmd.Args) != 1 {
		t.Fatalf("Expected 1 argument, got %d", len(cmd.Args))
	}

	if !strings.Contains(cmd.Args[0], "Hello, World!") {
		t.Fatalf("Expected argument to contain 'Hello, World!', got '%s'", cmd.Args[0])
	}

	// Check builtin and gexe flags
	if !cmd.IsBuiltin {
		t.Fatal("Expected IsBuiltin to be true for 'echo' command")
	}

	if cmd.UseGexe {
		t.Fatal("Expected UseGexe to be false for 'echo' command")
	}
}

// TestExtractWordValue tests the extractWordValue function
func TestExtractWordValue(t *testing.T) {
	script := `echo "Hello, World!"`

	// Parse the script
	result, err := ParseBashString(script)
	if err != nil {
		t.Fatalf("ParseBashString failed: %v", err)
	}

	// Get the Word from the parsed script
	var word *syntax.Word
	for _, stmt := range result.File.Stmts {
		if stmt.Cmd != nil {
			if call, ok := stmt.Cmd.(*syntax.CallExpr); ok {
				if len(call.Args) > 1 {
					word = call.Args[1]
					break
				}
			}
		}
	}

	if word == nil {
		// Create a simple word for testing
		word = &syntax.Word{
			Parts: []syntax.WordPart{
				&syntax.Lit{Value: "Hello, World!"},
			},
		}
	}

	// Extract the word value
	value := extractWordValue(word)

	// Verify the value
	if value != "Hello, World!" {
		t.Fatalf("Expected 'Hello, World!', got '%s'", value)
	}
}

// TestProcessAssign tests the processAssign function
func TestProcessAssign(t *testing.T) {
	script := `NAME="Test"`

	// Parse the script
	result, err := ParseBashString(script)
	if err != nil {
		t.Fatalf("ParseBashString failed: %v", err)
	}

	// Get the Assign from the parsed script
	var assign *syntax.Assign
	for _, stmt := range result.File.Stmts {
		// In the mvdan.cc/sh/v3/syntax package, assignments are part of the Cmd field
		// We need to walk through the AST to find the Assign node
		syntax.Walk(stmt, func(node syntax.Node) bool {
			if a, ok := node.(*syntax.Assign); ok {
				assign = a
				return false // Stop walking
			}
			return true // Continue walking
		})
		if assign != nil {
			break
		}
	}

	if assign == nil {
		t.Fatal("Failed to find Assign in parsed script")
	}

	// Process the Assign
	assignment := processAssign(assign)

	// Verify the assignment
	if assignment.Name != "NAME" {
		t.Fatalf("Expected name 'NAME', got '%s'", assignment.Name)
	}

	if !strings.Contains(assignment.Value, "Test") {
		t.Fatalf("Expected value to contain 'Test', got '%s'", assignment.Value)
	}

	if assignment.IsLocal {
		t.Fatal("Expected IsLocal to be false")
	}

	if assignment.IsExport {
		t.Fatal("Expected IsExport to be false")
	}
}

// TestProcessIfClause tests the processIfClause function
func TestProcessIfClause(t *testing.T) {
	script := `if [ -f "file.txt" ]; then
    echo "File exists"
else
    echo "File does not exist"
fi`

	// Parse the script
	result, err := ParseBashString(script)
	if err != nil {
		t.Fatalf("ParseBashString failed: %v", err)
	}

	// Get the IfClause from the parsed script
	var ifClause *syntax.IfClause
	for _, stmt := range result.File.Stmts {
		if stmt.Cmd != nil {
			if ifStmt, ok := stmt.Cmd.(*syntax.IfClause); ok {
				ifClause = ifStmt
				break
			}
		}
	}

	if ifClause == nil {
		t.Fatal("Failed to find IfClause in parsed script")
	}

	// Process the IfClause
	ifStmt := processIfClause(ifClause)

	// Verify the if statement
	if len(ifStmt.Condition) == 0 {
		t.Fatal("Expected non-empty condition")
	}

	if len(ifStmt.ThenBlock) == 0 {
		t.Fatal("Expected non-empty then block")
	}

	if ifStmt.ConditionType != "file" {
		t.Fatalf("Expected condition type 'file', got '%s'", ifStmt.ConditionType)
	}
}

// TestProcessForClause tests the processForClause function
func TestProcessForClause(t *testing.T) {
	script := `for i in 1 2 3; do
    echo "Item $i"
done`

	// Parse the script
	result, err := ParseBashString(script)
	if err != nil {
		t.Fatalf("ParseBashString failed: %v", err)
	}

	// Get the ForClause from the parsed script
	var forClause *syntax.ForClause
	for _, stmt := range result.File.Stmts {
		if stmt.Cmd != nil {
			if forStmt, ok := stmt.Cmd.(*syntax.ForClause); ok {
				forClause = forStmt
				break
			}
		}
	}

	if forClause == nil {
		t.Fatal("Failed to find ForClause in parsed script")
	}

	// Process the ForClause
	loop := processForClause(forClause)

	// Verify the loop
	if !loop.IsForEach {
		t.Fatal("Expected IsForEach to be true")
	}

	if len(loop.Body) == 0 {
		t.Fatal("Expected non-empty body")
	}
}

// TestProcessWhileClause tests the processWhileClause function
func TestProcessWhileClause(t *testing.T) {
	script := `while [ $i -lt 5 ]; do
    echo "Counter: $i"
    i=$((i+1))
done`

	// Parse the script
	result, err := ParseBashString(script)
	if err != nil {
		t.Fatalf("ParseBashString failed: %v", err)
	}

	// Get the WhileClause from the parsed script
	var whileClause *syntax.WhileClause
	for _, stmt := range result.File.Stmts {
		if stmt.Cmd != nil {
			if whileStmt, ok := stmt.Cmd.(*syntax.WhileClause); ok {
				whileClause = whileStmt
				break
			}
		}
	}

	if whileClause == nil {
		t.Fatal("Failed to find WhileClause in parsed script")
	}

	// Process the WhileClause
	loop := processWhileClause(whileClause)

	// Verify the loop
	if loop.Type != "while" {
		t.Fatalf("Expected type 'while', got '%s'", loop.Type)
	}

	if len(loop.Condition) == 0 {
		t.Fatal("Expected non-empty condition")
	}

	if len(loop.Body) == 0 {
		t.Fatal("Expected non-empty body")
	}
}

// TestProcessPipe tests the processPipe function
func TestProcessPipe(t *testing.T) {
	script := `ls -la | grep "file"`

	// Parse the script
	result, err := ParseBashString(script)
	if err != nil {
		t.Fatalf("ParseBashString failed: %v", err)
	}

	// Get the BinaryCmd from the parsed script
	var binaryCmd *syntax.BinaryCmd
	for _, stmt := range result.File.Stmts {
		if stmt.Cmd != nil {
			if binCmd, ok := stmt.Cmd.(*syntax.BinaryCmd); ok {
				if binCmd.Op == syntax.Pipe {
					binaryCmd = binCmd
					break
				}
			}
		}
	}

	if binaryCmd == nil {
		t.Fatal("Failed to find BinaryCmd with Pipe operator in parsed script")
	}

	// Process the BinaryCmd
	pipe := processPipe(binaryCmd)

	// Verify the pipe
	if len(pipe.Commands) == 0 {
		t.Fatal("Expected non-empty commands")
	}
}

// TestProcessSubshell tests the processSubshell function
func TestProcessSubshell(t *testing.T) {
	script := `(
    cd /tmp
    echo "In subshell"
)`

	// Parse the script
	result, err := ParseBashString(script)
	if err != nil {
		t.Fatalf("ParseBashString failed: %v", err)
	}

	// Get the Subshell from the parsed script
	var subshell *syntax.Subshell
	for _, stmt := range result.File.Stmts {
		if stmt.Cmd != nil {
			if subsh, ok := stmt.Cmd.(*syntax.Subshell); ok {
				subshell = subsh
				break
			}
		}
	}

	if subshell == nil {
		t.Fatal("Failed to find Subshell in parsed script")
	}

	// Process the Subshell
	subsh := processSubshell(subshell)

	// Verify the subshell
	if len(subsh.Statements) == 0 {
		t.Fatal("Expected non-empty statements")
	}
}

// TestProcessRedirection tests the processRedirection function
func TestProcessRedirection(t *testing.T) {
	script := `echo "Hello" > file.txt`

	// Parse the script
	result, err := ParseBashString(script)
	if err != nil {
		t.Fatalf("ParseBashString failed: %v", err)
	}

	// Get the Redirect from the parsed script
	var redirect *syntax.Redirect
	for _, stmt := range result.File.Stmts {
		// In the mvdan.cc/sh/v3/syntax package, redirections are part of the Stmt
		// We need to check if the Stmt has any redirections
		if len(stmt.Redirs) > 0 {
			redirect = stmt.Redirs[0]
			break
		}
	}

	if redirect == nil {
		t.Fatal("Failed to find Redirect in parsed script")
	}

	// Process the Redirect
	redirection := processRedirection(redirect)

	// Verify the redirection
	if redirection.Op != ">" {
		t.Fatalf("Expected operator '>', got '%s'", redirection.Op)
	}

	if redirection.Filename != "file.txt" {
		t.Fatalf("Expected filename 'file.txt', got '%s'", redirection.Filename)
	}
}
