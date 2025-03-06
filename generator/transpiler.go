package generator

import (
	"fmt"
	"strings"

	"github.com/TFMV/bash2go/parser"
)

// Import the template.go file
//go:generate go run -mod=mod github.com/TFMV/bash2go/generator/template.go

// GoCodeGenerator generates Go code from an intermediate representation
type GoCodeGenerator struct {
	IR              *parser.IntermediateRepresentation
	RequiredImports map[string]bool
	Generator       *CodeGenerator
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
	RequiresGexe     bool
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
		Generator:       NewCodeGenerator("main"),
	}
}

// Generate generates Go code from the intermediate representation
func (g *GoCodeGenerator) Generate() (string, error) {
	// Initialize the code generator
	g.Generator = NewCodeGenerator("main")
	g.RequiredImports = make(map[string]bool)

	// Check if we need special imports
	for _, stmt := range g.IR.MainStatements {
		if stmt.Type == parser.StatementCommand {
			cmd := stmt.Value.(parser.Command)
			if cmd.UseGexe {
				g.RequiredImports["github.com/vladimirvivien/gexe"] = true
			} else if !cmd.IsBuiltin {
				g.RequiredImports["os/exec"] = true
			}

			// Add fmt for echo commands
			if cmd.Name == "echo" {
				g.RequiredImports["fmt"] = true
			}
		} else if stmt.Type == parser.StatementPipe {
			g.RequiredImports["github.com/vladimirvivien/gexe"] = true
		}
	}

	// Add imports to the generator
	for imp := range g.RequiredImports {
		g.Generator.AddImport(imp)
	}

	// Add variables
	for name, value := range g.IR.Variables {
		g.Generator.AddGlobal(fmt.Sprintf("var %s = %s", name, value))
	}

	// Add functions
	for name, function := range g.IR.Functions {
		funcBody, err := g.generateStatements(function.Statements)
		if err != nil {
			return "", err
		}

		// Split the function body into lines
		bodyLines := strings.Split(funcBody, "\n")

		// Create a new function
		fn := Function{
			Name: name,
			Body: bodyLines,
			Comments: []string{
				fmt.Sprintf("Function %s from the original Bash script", name),
			},
		}

		g.Generator.AddFunction(fn)
	}

	// Create main function
	mainBody, err := g.generateStatements(g.IR.MainStatements)
	if err != nil {
		return "", err
	}

	// Split the main body into lines
	mainLines := strings.Split(mainBody, "\n")

	// Create the main function
	mainFn := Function{
		Name: "main",
		Body: mainLines,
		Comments: []string{
			"Main function generated from Bash script",
		},
	}

	g.Generator.AddFunction(mainFn)

	// Build the code
	return g.Generator.Build()
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
		return g.generateAssignment(assignment)
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
	case parser.StatementRedirection:
		redirection := stmt.Value.(parser.Redirection)
		return g.generateRedirection(redirection)
	case parser.StatementFunction:
		// Functions are handled separately in the Generate method
		return "// Function declaration (handled separately)", nil
	case parser.StatementBackground:
		background := stmt.Value.(parser.Background)
		// Generate the command code first
		cmdCode, err := g.generateCommand(background.Command)
		if err != nil {
			return "", err
		}
		g.RequiredImports["sync"] = true
		return fmt.Sprintf("go func() {\n\t%s\n}()", cmdCode), nil
	case parser.StatementReturn:
		returnStmt := stmt.Value.(parser.Return)
		if returnStmt.Value != "" {
			return fmt.Sprintf("return %s", returnStmt.Value), nil
		}
		return fmt.Sprintf("return %d", returnStmt.Code), nil
	default:
		return fmt.Sprintf("// Unsupported statement type: %v", stmt.Type), nil
	}
}

// generateCommand generates Go code for a command
func (g *GoCodeGenerator) generateCommand(cmd parser.Command) (string, error) {
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
				// Check for variable substitution inside the string
				if strings.Contains(arg, "$") {
					// Replace $VAR with " + VAR + "
					processedArg := arg
					// Remove the outer quotes
					processedArg = processedArg[1 : len(processedArg)-1]
					// Replace ${VAR} with " + VAR + "
					processedArg = strings.ReplaceAll(processedArg, "${", "\" + ")
					processedArg = strings.ReplaceAll(processedArg, "}", " + \"")
					// Replace $VAR with " + VAR + "
					for _, varName := range extractVariableNames(processedArg) {
						processedArg = strings.ReplaceAll(processedArg, "$"+varName, "\" + "+varName+" + \"")
					}
					// Add the outer quotes back
					args = append(args, "\""+processedArg+"\"")
				} else {
					args = append(args, arg)
				}
			} else if strings.HasPrefix(arg, "$") {
				// This is a variable reference
				varName := strings.TrimPrefix(arg, "$")
				// Handle ${VAR} format
				if strings.HasPrefix(varName, "{") && strings.HasSuffix(varName, "}") {
					varName = varName[1 : len(varName)-1]
				}
				args = append(args, varName)
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

		// Handle the argument
		arg := cmd.Args[0]
		if strings.HasPrefix(arg, "$") {
			// This is a variable reference
			varName := strings.TrimPrefix(arg, "$")
			// Handle ${VAR} format
			if strings.HasPrefix(varName, "{") && strings.HasSuffix(varName, "}") {
				varName = varName[1 : len(varName)-1]
			}
			return fmt.Sprintf("err = os.Chdir(%s)", varName), nil
		}

		return fmt.Sprintf("err = os.Chdir(\"%s\")", arg), nil
	case "pwd":
		// Use os.Getwd instead of exec.Command
		g.RequiredImports["os"] = true
		return `dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(dir)`, nil
	case "mkdir":
		// Use os.MkdirAll instead of exec.Command
		g.RequiredImports["os"] = true
		if len(cmd.Args) == 0 {
			return "// Warning: mkdir command with no arguments", nil
		}

		// Handle the argument
		arg := cmd.Args[0]
		if strings.HasPrefix(arg, "$") {
			// This is a variable reference
			varName := strings.TrimPrefix(arg, "$")
			return fmt.Sprintf("err = os.MkdirAll(%s, 0755)", varName), nil
		}

		return fmt.Sprintf("err = os.MkdirAll(\"%s\", 0755)", arg), nil
	case "rm":
		// Use os.Remove or os.RemoveAll instead of exec.Command
		g.RequiredImports["os"] = true
		if len(cmd.Args) == 0 {
			return "// Warning: rm command with no arguments", nil
		}

		// Check for -r or -rf flag
		isRecursive := false
		var target string

		for _, arg := range cmd.Args {
			if arg == "-r" || arg == "-rf" || arg == "-fr" {
				isRecursive = true
			} else if !strings.HasPrefix(arg, "-") {
				target = arg
			}
		}

		// Handle variable reference
		if strings.HasPrefix(target, "$") {
			varName := strings.TrimPrefix(target, "$")
			if isRecursive {
				return fmt.Sprintf("err = os.RemoveAll(%s)", varName), nil
			}
			return fmt.Sprintf("err = os.Remove(%s)", varName), nil
		}

		if isRecursive {
			return fmt.Sprintf("err = os.RemoveAll(\"%s\")", target), nil
		}
		return fmt.Sprintf("err = os.Remove(\"%s\")", target), nil
	case "cp":
		// Use io/ioutil or os for file copying
		g.RequiredImports["io/ioutil"] = true
		g.RequiredImports["os"] = true
		if len(cmd.Args) < 2 {
			return "// Warning: cp command with insufficient arguments", nil
		}

		src := cmd.Args[0]
		dst := cmd.Args[1]

		// Handle variable references
		if strings.HasPrefix(src, "$") {
			src = strings.TrimPrefix(src, "$")
		} else {
			src = fmt.Sprintf("\"%s\"", src)
		}

		if strings.HasPrefix(dst, "$") {
			dst = strings.TrimPrefix(dst, "$")
		} else {
			dst = fmt.Sprintf("\"%s\"", dst)
		}

		return fmt.Sprintf(`data, err := ioutil.ReadFile(%s)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(%s, data, 0644)`, src, dst), nil
	case "test", "[":
		// Use os.Stat and other Go functions for test conditions
		g.RequiredImports["os"] = true

		if len(cmd.Args) < 2 {
			return "// Warning: test command with insufficient arguments", nil
		}

		// Handle different test conditions
		switch cmd.Args[0] {
		case "-f":
			// Test if file exists
			arg := cmd.Args[1]
			if strings.HasPrefix(arg, "$") {
				// This is a variable reference
				varName := strings.TrimPrefix(arg, "$")
				return fmt.Sprintf("_, err := os.Stat(%s); err == nil", varName), nil
			}
			return fmt.Sprintf("_, err := os.Stat(\"%s\"); err == nil", arg), nil
		case "-d":
			// Test if directory exists
			arg := cmd.Args[1]
			if strings.HasPrefix(arg, "$") {
				// This is a variable reference
				varName := strings.TrimPrefix(arg, "$")
				return fmt.Sprintf(`info, err := os.Stat(%s)
	if err == nil && info.IsDir()`, varName), nil
			}
			return fmt.Sprintf(`info, err := os.Stat("%s")
	if err == nil && info.IsDir()`, arg), nil
		case "-z":
			// Test if string is empty
			arg := cmd.Args[1]
			if strings.HasPrefix(arg, "$") {
				// This is a variable reference
				varName := strings.TrimPrefix(arg, "$")
				return fmt.Sprintf("len(%s) == 0", varName), nil
			}
			return fmt.Sprintf("len(\"%s\") == 0", arg), nil
		case "-n":
			// Test if string is not empty
			arg := cmd.Args[1]
			if strings.HasPrefix(arg, "$") {
				// This is a variable reference
				varName := strings.TrimPrefix(arg, "$")
				return fmt.Sprintf("len(%s) > 0", varName), nil
			}
			return fmt.Sprintf("len(\"%s\") > 0", arg), nil
		default:
			// Use gexe for other test conditions
			g.RequiredImports["github.com/vladimirvivien/gexe"] = true
			return fmt.Sprintf("exe.Run(\"test %s\").Success()", strings.Join(cmd.Args, " ")), nil
		}
	case "exit":
		// Use os.Exit
		g.RequiredImports["os"] = true
		if len(cmd.Args) == 0 {
			return "os.Exit(0)", nil
		}

		// Handle the exit code
		code := cmd.Args[0]
		if strings.HasPrefix(code, "$") {
			// This is a variable reference
			varName := strings.TrimPrefix(code, "$")
			return fmt.Sprintf("os.Exit(%s)", varName), nil
		}

		return fmt.Sprintf("os.Exit(%s)", code), nil
	default:
		// For external commands, use gexe
		if cmd.UseGexe {
			g.RequiredImports["github.com/vladimirvivien/gexe"] = true

			// Build the command string
			var cmdStr strings.Builder
			cmdStr.WriteString(cmd.Name)

			for _, arg := range cmd.Args {
				cmdStr.WriteString(" ")

				// If the argument contains spaces, quote it
				if strings.Contains(arg, " ") && !strings.HasPrefix(arg, "\"") {
					cmdStr.WriteString("\"")
					cmdStr.WriteString(arg)
					cmdStr.WriteString("\"")
				} else {
					cmdStr.WriteString(arg)
				}
			}

			return fmt.Sprintf(`// Execute command: %s
	output := exe.Run("%s").Stdout()
	fmt.Print(output)`, cmdStr.String(), cmdStr.String()), nil
		}

		// For other commands, use exec.Command as a fallback
		g.RequiredImports["os/exec"] = true
		g.RequiredImports["fmt"] = true

		// Build the command arguments
		var args []string
		for _, arg := range cmd.Args {
			if strings.HasPrefix(arg, "$") {
				// This is a variable reference
				varName := strings.TrimPrefix(arg, "$")
				args = append(args, varName)
			} else {
				args = append(args, fmt.Sprintf("\"%s\"", arg))
			}
		}

		argsStr := ""
		if len(args) > 0 {
			argsStr = ", " + strings.Join(args, ", ")
		}

		return fmt.Sprintf(`cmd := exec.Command("%s"%s)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command: %%v", err)
	}
	fmt.Print(string(output))`, cmd.Name, argsStr), nil
	}
}

// generateAssignment generates Go code for a variable assignment
func (g *GoCodeGenerator) generateAssignment(assign parser.Assignment) (string, error) {
	// Handle local variables
	if assign.IsLocal {
		return fmt.Sprintf("var %s = %s", assign.Name, assign.Value), nil
	}

	// Handle export variables
	if assign.IsExport {
		g.RequiredImports["os"] = true
		return fmt.Sprintf("os.Setenv(\"%s\", %s)", assign.Name, assign.Value), nil
	}

	// Handle regular variables
	return fmt.Sprintf("%s := %s", assign.Name, assign.Value), nil
}

// generateIf generates Go code for an if statement
func (g *GoCodeGenerator) generateIf(ifStmt parser.If) (string, error) {
	// Generate condition
	condition, err := g.generateCondition(ifStmt.Condition, ifStmt.ConditionType)
	if err != nil {
		return "", err
	}

	// Generate then block
	thenBlock, err := g.generateStatements(ifStmt.ThenBlock)
	if err != nil {
		return "", err
	}

	// Generate else block
	elseBlock := ""
	if len(ifStmt.ElseBlock) > 0 {
		var err error
		elseBlock, err = g.generateStatements(ifStmt.ElseBlock)
		if err != nil {
			return "", err
		}
	}

	// Build the if statement
	var result strings.Builder
	result.WriteString(fmt.Sprintf("if %s {\n", condition))
	result.WriteString(thenBlock)
	if elseBlock != "" {
		result.WriteString("} else {\n")
		result.WriteString(elseBlock)
	}
	result.WriteString("}")

	return result.String(), nil
}

// generateCondition generates Go code for a condition
func (g *GoCodeGenerator) generateCondition(conditions []parser.Statement, conditionType string) (string, error) {
	if len(conditions) == 0 {
		return "true", nil
	}

	// For now, just use the first condition
	stmt := conditions[0]
	if stmt.Type == parser.StatementCommand {
		cmd := stmt.Value.(parser.Command)

		// Handle test conditions
		if cmd.Name == "test" || cmd.Name == "[" {
			if len(cmd.Args) >= 2 {
				switch cmd.Args[0] {
				case "-f":
					// Test if file exists
					g.RequiredImports["os"] = true
					return fmt.Sprintf("_, err := os.Stat(\"%s\"); err == nil", cmd.Args[1]), nil
				case "-d":
					// Test if directory exists
					g.RequiredImports["os"] = true
					return fmt.Sprintf("info, err := os.Stat(\"%s\"); err == nil && info.IsDir()", cmd.Args[1]), nil
				case "-z":
					// Test if string is empty
					return fmt.Sprintf("len(\"%s\") == 0", cmd.Args[1]), nil
				case "-n":
					// Test if string is not empty
					return fmt.Sprintf("len(\"%s\") > 0", cmd.Args[1]), nil
				case "=":
					// Test if strings are equal
					if len(cmd.Args) >= 3 {
						return fmt.Sprintf("\"%s\" == \"%s\"", cmd.Args[1], cmd.Args[2]), nil
					}
				case "!=":
					// Test if strings are not equal
					if len(cmd.Args) >= 3 {
						return fmt.Sprintf("\"%s\" != \"%s\"", cmd.Args[1], cmd.Args[2]), nil
					}
				case "-eq":
					// Test if numbers are equal
					if len(cmd.Args) >= 3 {
						return fmt.Sprintf("%s == %s", cmd.Args[1], cmd.Args[2]), nil
					}
				case "-ne":
					// Test if numbers are not equal
					if len(cmd.Args) >= 3 {
						return fmt.Sprintf("%s != %s", cmd.Args[1], cmd.Args[2]), nil
					}
				case "-lt":
					// Test if number is less than
					if len(cmd.Args) >= 3 {
						return fmt.Sprintf("%s < %s", cmd.Args[1], cmd.Args[2]), nil
					}
				case "-le":
					// Test if number is less than or equal
					if len(cmd.Args) >= 3 {
						return fmt.Sprintf("%s <= %s", cmd.Args[1], cmd.Args[2]), nil
					}
				case "-gt":
					// Test if number is greater than
					if len(cmd.Args) >= 3 {
						return fmt.Sprintf("%s > %s", cmd.Args[1], cmd.Args[2]), nil
					}
				case "-ge":
					// Test if number is greater than or equal
					if len(cmd.Args) >= 3 {
						return fmt.Sprintf("%s >= %s", cmd.Args[1], cmd.Args[2]), nil
					}
				}
			}
		}

		// For other commands, use gexe
		g.RequiredImports["github.com/vladimirvivien/gexe"] = true

		// Build the command string
		var cmdStr strings.Builder
		cmdStr.WriteString(cmd.Name)

		for _, arg := range cmd.Args {
			cmdStr.WriteString(" ")
			cmdStr.WriteString(arg)
		}

		return fmt.Sprintf("exe.Run(\"%s\").Success()", cmdStr.String()), nil
	}

	return "true", nil
}

// generateLoop generates Go code for a loop
func (g *GoCodeGenerator) generateLoop(loop parser.Loop) (string, error) {
	// Generate loop body
	body, err := g.generateStatements(loop.Body)
	if err != nil {
		return "", err
	}

	// Handle different loop types
	switch loop.Type {
	case "for":
		if loop.IsForEach {
			// This is a for-each loop
			g.RequiredImports["strings"] = true

			// Split the items by space
			return fmt.Sprintf(`items := strings.Fields("%s")
	for _, %s := range items {
		%s
	}`, loop.Items, loop.RangeVar, body), nil
		} else if loop.IsRange {
			// This is a range loop
			return fmt.Sprintf(`for %s := %s; %s <= %s; %s++ {
		%s
	}`, loop.RangeVar, loop.RangeFrom, loop.RangeVar, loop.RangeTo, loop.RangeVar, body), nil
		}

		// Default for loop
		return fmt.Sprintf(`for {
		%s
	}`, body), nil
	case "while":
		// Generate condition
		condition, err := g.generateCondition(loop.Condition, "command")
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(`for %s {
		%s
	}`, condition, body), nil
	case "until":
		// Generate condition
		condition, err := g.generateCondition(loop.Condition, "command")
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(`for !(%s) {
		%s
	}`, condition, body), nil
	default:
		return fmt.Sprintf(`// Unsupported loop type: %s
	for {
		%s
	}`, loop.Type, body), nil
	}
}

// generatePipe generates Go code for a pipe
func (g *GoCodeGenerator) generatePipe(pipe parser.Pipe) (string, error) {
	if len(pipe.Commands) == 0 {
		return "// Empty pipe", nil
	}

	// Use gexe for pipes
	g.RequiredImports["github.com/vladimirvivien/gexe"] = true

	// Build the piped command string
	var cmdStr strings.Builder

	for i, cmd := range pipe.Commands {
		if i > 0 {
			cmdStr.WriteString(" | ")
		}

		cmdStr.WriteString(cmd.Name)

		for _, arg := range cmd.Args {
			cmdStr.WriteString(" ")

			// If the argument contains spaces, quote it
			if strings.Contains(arg, " ") && !strings.HasPrefix(arg, "\"") {
				cmdStr.WriteString("\"")
				cmdStr.WriteString(arg)
				cmdStr.WriteString("\"")
			} else {
				cmdStr.WriteString(arg)
			}
		}
	}

	return fmt.Sprintf(`// Execute piped command: %s
	output := exe.Run("%s").Stdout()
	fmt.Print(output)`, cmdStr.String(), cmdStr.String()), nil
}

// generateSubshell generates Go code for a subshell
func (g *GoCodeGenerator) generateSubshell(subshell parser.Subshell) (string, error) {
	// Generate subshell statements
	stmts, err := g.generateStatements(subshell.Statements)
	if err != nil {
		return "", err
	}

	// Wrap in a function to create a new scope
	return fmt.Sprintf(`// Execute subshell
	func() {
		%s
	}()`, stmts), nil
}

// generateRedirection generates Go code for a redirection
func (g *GoCodeGenerator) generateRedirection(redirection parser.Redirection) (string, error) {
	// Use os package for redirections
	g.RequiredImports["os"] = true

	switch redirection.Op {
	case ">":
		// Output redirection (overwrite)
		return fmt.Sprintf(`// Redirect output to %s
	file, err := os.Create("%s")
	if err != nil {
		return err
	}
	defer file.Close()
	
	// TODO: Execute command and write output to file`, redirection.Filename, redirection.Filename), nil
	case ">>":
		// Output redirection (append)
		return fmt.Sprintf(`// Redirect output to %s (append)
	file, err := os.OpenFile("%s", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// TODO: Execute command and write output to file`, redirection.Filename, redirection.Filename), nil
	case "<":
		// Input redirection
		return fmt.Sprintf(`// Redirect input from %s
	file, err := os.Open("%s")
	if err != nil {
		return err
	}
	defer file.Close()
	
	// TODO: Execute command with input from file`, redirection.Filename, redirection.Filename), nil
	default:
		return fmt.Sprintf("// Unsupported redirection: %s", redirection.Op), nil
	}
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

// extractVariableNames extracts variable names from a string
func extractVariableNames(s string) []string {
	var result []string
	var inVar bool
	var varName strings.Builder

	for i := 0; i < len(s); i++ {
		if s[i] == '$' && i+1 < len(s) && isValidVarNameStart(s[i+1]) {
			inVar = true
			varName.Reset()
			continue
		}

		if inVar {
			if isValidVarNameChar(s[i]) {
				varName.WriteByte(s[i])
			} else {
				result = append(result, varName.String())
				inVar = false
			}
		}
	}

	if inVar {
		result = append(result, varName.String())
	}

	return result
}

// isValidVarNameStart checks if a character is valid as the first character of a variable name
func isValidVarNameStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// isValidVarNameChar checks if a character is valid in a variable name
func isValidVarNameChar(c byte) bool {
	return isValidVarNameStart(c) || (c >= '0' && c <= '9')
}
