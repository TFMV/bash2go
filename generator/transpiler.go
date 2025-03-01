package generator

import (
	"fmt"
	"strings"

	"github.com/TFMV/bash2go/parser"
)

// GoCodeGenerator generates Go code from an intermediate representation
type GoCodeGenerator struct {
	IR              *parser.IntermediateRepresentation
	RequiredImports map[string]bool
}

// TemplateData holds data for main template
type TemplateData struct {
	Imports          []string
	Functions        []FunctionData
	Variables        []VariableData
	MainBody         string
	RequiresExec     bool
	RequiresFilepath bool
	RequiresStrings  bool
}

// FunctionData holds data for function template
type FunctionData struct {
	Name        string
	Body        string
	ReturnValue bool
}

// VariableData holds data for variable template
type VariableData struct {
	Name  string
	Value string
}

// NewGoCodeGenerator creates a new code generator
func NewGoCodeGenerator(ir *parser.IntermediateRepresentation) *GoCodeGenerator {
	return &GoCodeGenerator{
		IR:              ir,
		RequiredImports: make(map[string]bool),
	}
}

// Generate generates Go code from the intermediate representation
func (g *GoCodeGenerator) Generate() (string, error) {
	// Process all statements and build template data
	templateData := TemplateData{
		Imports:          []string{},
		Functions:        []FunctionData{},
		Variables:        []VariableData{},
		MainBody:         "",
		RequiresExec:     false,
		RequiresFilepath: false,
		RequiresStrings:  false,
	}

	// Add functions
	for name, function := range g.IR.Functions {
		funcBody, err := g.generateStatements(function.Statements)
		if err != nil {
			return "", err
		}

		templateData.Functions = append(templateData.Functions, FunctionData{
			Name:        name,
			Body:        funcBody,
			ReturnValue: false,
		})
	}

	// Add variables
	for name, value := range g.IR.Variables {
		templateData.Variables = append(templateData.Variables, VariableData{
			Name:  name,
			Value: value,
		})
	}

	// Generate main body
	mainBody, err := g.generateStatements(g.IR.MainStatements)
	if err != nil {
		return "", err
	}
	templateData.MainBody = mainBody

	// Add required imports
	for imp := range g.RequiredImports {
		templateData.Imports = append(templateData.Imports, imp)
	}

	// Check if we need special imports
	for _, stmt := range g.IR.MainStatements {
		if stmt.Type == parser.StatementCommand {
			cmd := stmt.Value.(parser.Command)
			if cmd.Name != "echo" && cmd.Name != "cd" && cmd.Name != "rm" && cmd.Name != "mkdir" && cmd.Name != "cp" {
				templateData.RequiresExec = true
			}
		} else if stmt.Type == parser.StatementPipe {
			templateData.RequiresExec = true
		}
	}

	// If we have any file operations, we need filepath
	if g.RequiredImports["os"] {
		templateData.RequiresFilepath = true
	}

	// We don't need strings for simple echo commands
	templateData.RequiresStrings = false

	// Execute main template
	return ExecuteTemplate(mainTemplate, templateData)
}

// generateStatements generates Go code for a slice of statements
func (g *GoCodeGenerator) generateStatements(statements []parser.Statement) (string, error) {
	var result strings.Builder

	for _, stmt := range statements {
		code, err := g.generateStatement(stmt)
		if err != nil {
			return "", err
		}
		result.WriteString(code)
		result.WriteString("\n")
	}

	return result.String(), nil
}

// generateStatement generates Go code for a single statement
func (g *GoCodeGenerator) generateStatement(stmt parser.Statement) (string, error) {
	switch stmt.Type {
	case parser.StatementCommand:
		cmd := stmt.Value.(parser.Command)
		return g.generateCommand(cmd)
	case parser.StatementAssignment:
		assignment := stmt.Value.(parser.Assignment)
		return fmt.Sprintf("%s := %s", assignment.Name, assignment.Value), nil
	case parser.StatementIf:
		ifStmt := stmt.Value.(parser.If)
		return g.generateIf(ifStmt)
	case parser.StatementLoop:
		loop := stmt.Value.(parser.Loop)
		return g.generateLoop(loop)
	case parser.StatementPipe:
		pipe := stmt.Value.(parser.Pipe)
		return g.generatePipe(pipe)
	case parser.StatementSubshell:
		subshell := stmt.Value.(parser.Subshell)
		return g.generateSubshell(subshell)
	default:
		return "", fmt.Errorf("unsupported statement type: %v", stmt.Type)
	}
}

// generateCommand generates Go code for a command
func (g *GoCodeGenerator) generateCommand(cmd parser.Command) (string, error) {
	// Debug output
	fmt.Printf("Generating code for command: %s with args: %v\n", cmd.Name, cmd.Args)

	// Handle built-in commands with Go equivalents
	switch cmd.Name {
	case "echo":
		// Use fmt.Println instead of exec.Command
		g.RequiredImports["fmt"] = true
		if len(cmd.Args) == 0 {
			return "fmt.Println()", nil
		}

		// Handle quoted arguments
		var args []string
		for _, arg := range cmd.Args {
			// If the argument is already quoted, use it as is
			if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") {
				args = append(args, arg)
			} else {
				// Otherwise, quote it
				args = append(args, fmt.Sprintf("\"%s\"", arg))
			}
		}

		return fmt.Sprintf("fmt.Println(%s)", strings.Join(args, ", ")), nil
	case "cd":
		// Use os.Chdir instead of exec.Command
		g.RequiredImports["os"] = true
		if len(cmd.Args) == 0 {
			return "err = os.Chdir(os.Getenv(\"HOME\"))", nil
		}
		return fmt.Sprintf("err = os.Chdir(%s)", cmd.Args[0]), nil
	case "rm":
		// Use os.Remove or os.RemoveAll instead of exec.Command
		g.RequiredImports["os"] = true
		if len(cmd.Args) == 0 {
			return "// Warning: rm command with no arguments", nil
		}
		if contains(cmd.Args, "-r") || contains(cmd.Args, "-rf") {
			return fmt.Sprintf("err = os.RemoveAll(%s)", cmd.Args[len(cmd.Args)-1]), nil
		}
		return fmt.Sprintf("err = os.Remove(%s)", cmd.Args[len(cmd.Args)-1]), nil
	case "mkdir":
		// Use os.MkdirAll instead of exec.Command
		g.RequiredImports["os"] = true
		if len(cmd.Args) == 0 {
			return "// Warning: mkdir command with no arguments", nil
		}
		return fmt.Sprintf("err = os.MkdirAll(%s, 0755)", cmd.Args[len(cmd.Args)-1]), nil
	case "cp":
		// Use io/ioutil or os for file copying
		g.RequiredImports["io/ioutil"] = true
		if len(cmd.Args) < 2 {
			return "// Warning: cp command with insufficient arguments", nil
		}
		return fmt.Sprintf(`data, err := ioutil.ReadFile(%s)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(%s, data, 0644)`, cmd.Args[0], cmd.Args[1]), nil
	case "curl":
		// Use net/http instead of exec.Command
		g.RequiredImports["net/http"] = true
		g.RequiredImports["io/ioutil"] = true
		g.RequiredImports["fmt"] = true
		// This is a simplification; a real implementation would need to parse curl arguments
		return `resp, err := http.Get("url")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))`, nil
	default:
		// Special handling for placeholder commands
		if strings.Contains(cmd.Name, "not fully implemented") {
			g.RequiredImports["fmt"] = true
			return fmt.Sprintf("fmt.Println(\"%s\")", strings.ReplaceAll(cmd.Name, "\"", "\\\"")), nil
		}

		// For other commands, use exec.Command
		if cmd.Name == "" && len(cmd.Args) == 0 {
			return "// Empty command", nil
		}
		return ExecuteTemplate(commandTemplate, cmd)
	}
}

// generateIf generates Go code for an if statement
func (g *GoCodeGenerator) generateIf(ifStmt parser.If) (string, error) {
	// TODO: Implement if statement generation
	g.RequiredImports["fmt"] = true
	return "fmt.Println(\"If statement not fully implemented\")", nil
}

// generateLoop generates Go code for a loop
func (g *GoCodeGenerator) generateLoop(loop parser.Loop) (string, error) {
	// TODO: Implement loop generation
	g.RequiredImports["fmt"] = true
	return "fmt.Println(\"Loop not fully implemented\")", nil
}

// generatePipe generates Go code for a pipe
func (g *GoCodeGenerator) generatePipe(pipe parser.Pipe) (string, error) {
	if len(pipe.Commands) == 0 {
		return "// Empty pipe", nil
	}
	return ExecuteTemplate(pipeTemplate, pipe)
}

// generateSubshell generates Go code for a subshell
func (g *GoCodeGenerator) generateSubshell(subshell parser.Subshell) (string, error) {
	stmts, err := g.generateStatements(subshell.Statements)
	if err != nil {
		return "", err
	}

	data := struct {
		Statements string
	}{
		Statements: stmts,
	}

	return ExecuteTemplate(subshellTemplate, data)
}

// Helper function to check if a slice contains a string
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
