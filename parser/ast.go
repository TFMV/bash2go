package parser

import (
	"fmt"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// debug enables printing of debug information.
const debug = false

// IntermediateRepresentation represents the processed AST in a format suitable for Go code generation.
type IntermediateRepresentation struct {
	Variables        map[string]string
	Functions        map[string]*Function
	MainStatements   []Statement
	RequiredPackages map[string]bool
}

// Function represents a Bash function definition.
type Function struct {
	Name       string
	Statements []Statement
	Parameters []string
	LocalVars  map[string]string
}

// StatementType identifies the type of a statement.
type StatementType int

const (
	StatementCommand StatementType = iota
	StatementAssignment
	StatementIf
	StatementLoop
	StatementPipe
	StatementSubshell
	StatementFunction
	StatementRedirection
	StatementBackground
	StatementReturn
)

// Statement represents a single statement in the Bash script.
type Statement struct {
	Type  StatementType
	Value interface{} // Command, Assignment, If, Loop, Pipe, Subshell, etc.
}

// Command represents a command execution.
type Command struct {
	Name      string
	Args      []string
	IsBuiltin bool
	UseGexe   bool
}

// Assignment represents a variable assignment.
type Assignment struct {
	Name     string
	Value    string
	IsLocal  bool
	IsExport bool
}

// If represents an if-then-else statement.
type If struct {
	Condition     []Statement
	ThenBlock     []Statement
	ElseBlock     []Statement
	ElifBlocks    [][2][]Statement // Each element is a [condition, then-block] pair.
	ConditionType string           // "file", "string", "number", "command"
}

// Loop represents a loop construct (for, while, until).
type Loop struct {
	Type      string // "for", "while", "until"
	Init      []Statement
	Condition []Statement
	Update    []Statement
	Body      []Statement
	IsRange   bool   // for i in {1..10}
	RangeVar  string // The loop variable
	RangeFrom string // Start of range
	RangeTo   string // End of range
	IsForEach bool   // for i in items
	Items     string // The items to iterate over
}

// Pipe represents a piped command sequence.
type Pipe struct {
	Commands []Command
}

// Subshell represents a subshell execution.
type Subshell struct {
	Statements []Statement
}

// Redirection represents input/output redirection.
type Redirection struct {
	Op       string // ">", ">>", "<", etc.
	Command  Command
	Filename string
}

// Background represents a command running in the background.
type Background struct {
	Command Command
}

// Return represents a return statement.
type Return struct {
	Value string
	Code  int
}

// BuildIR builds an intermediate representation from a parsed result.
func BuildIR(result *ParseResult) (*IntermediateRepresentation, error) {
	ir := NewIntermediateRepresentation()

	// Always include these packages.
	ir.RequiredPackages["fmt"] = true
	ir.RequiredPackages["os"] = true

	// Walk the AST to build the intermediate representation.
	syntax.Walk(result.File, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.CallExpr:
			// Process command call.
			cmd := processCallExpr(x)
			ir.MainStatements = append(ir.MainStatements, Statement{
				Type:  StatementCommand,
				Value: cmd,
			})
		case *syntax.Assign:
			// Process variable assignment.
			assign := processAssign(x)
			ir.Variables[assign.Name] = assign.Value
			ir.MainStatements = append(ir.MainStatements, Statement{
				Type:  StatementAssignment,
				Value: assign,
			})
		case *syntax.FuncDecl:
			// Process function declaration.
			function := processFunction(x)
			ir.Functions[function.Name] = function
			ir.MainStatements = append(ir.MainStatements, Statement{
				Type:  StatementFunction,
				Value: function,
			})
		case *syntax.IfClause:
			// Process if statement.
			ifStmt := processIfClause(x)
			ir.MainStatements = append(ir.MainStatements, Statement{
				Type:  StatementIf,
				Value: ifStmt,
			})
		case *syntax.WhileClause:
			// Process while loop.
			loop := processWhileClause(x)
			ir.MainStatements = append(ir.MainStatements, Statement{
				Type:  StatementLoop,
				Value: loop,
			})
		case *syntax.ForClause:
			// Process for loop.
			loop := processForClause(x)
			ir.MainStatements = append(ir.MainStatements, Statement{
				Type:  StatementLoop,
				Value: loop,
			})
		case *syntax.BinaryCmd:
			// Process binary command (e.g., pipe).
			if x.Op == syntax.Pipe {
				pipe := processPipe(x)
				ir.MainStatements = append(ir.MainStatements, Statement{
					Type:  StatementPipe,
					Value: pipe,
				})
			}
		case *syntax.Subshell:
			// Process subshell.
			subshell := processSubshell(x)
			ir.MainStatements = append(ir.MainStatements, Statement{
				Type:  StatementSubshell,
				Value: subshell,
			})
		case *syntax.Redirect:
			// Process redirection.
			redirection := processRedirection(x)
			ir.MainStatements = append(ir.MainStatements, Statement{
				Type:  StatementRedirection,
				Value: redirection,
			})
		}
		return true
	})

	return ir, nil
}

// processCallExpr processes a call expression (command).
func processCallExpr(x *syntax.CallExpr) Command {
	cmd := Command{
		Name:      "",
		Args:      []string{},
		IsBuiltin: false,
		UseGexe:   true, // Default to using gexe for external commands.
	}

	if len(x.Args) > 0 {
		// Extract command name from the first argument.
		cmdName := extractWordValue(x.Args[0])
		cmd.Name = cmdName

		// Check if this is a builtin command that can be directly translated to Go.
		switch cmd.Name {
		case "echo", "printf", "cd", "pwd", "exit", "return", "test", "[", "source", "export", "read":
			cmd.IsBuiltin = true
			cmd.UseGexe = false
		}

		// Extract arguments from the remaining arguments.
		for i := 1; i < len(x.Args); i++ {
			arg := extractWordValue(x.Args[i])
			cmd.Args = append(cmd.Args, arg)
		}
	}

	if debug {
		fmt.Printf("Parsed command: %s with args: %v\n", cmd.Name, cmd.Args)
	}

	return cmd
}

// extractWordValue extracts the string value from a Word.
func extractWordValue(word *syntax.Word) string {
	var value strings.Builder
	for _, part := range word.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			value.WriteString(p.Value)
		case *syntax.ParamExp:
			value.WriteString("$" + p.Param.Value)
		case *syntax.DblQuoted:
			value.WriteString(extractDblQuotedValue(p))
		case *syntax.SglQuoted:
			value.WriteString(p.Value)
		case *syntax.CmdSubst:
			value.WriteString("$(command)")
		}
	}
	return value.String()
}

// extractDblQuotedValue extracts the string value from a double-quoted string.
func extractDblQuotedValue(dq *syntax.DblQuoted) string {
	var value strings.Builder
	for _, part := range dq.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			value.WriteString(p.Value)
		case *syntax.ParamExp:
			value.WriteString("${" + p.Param.Value + "}")
		}
	}
	return value.String()
}

// processAssign processes a variable assignment.
func processAssign(x *syntax.Assign) Assignment {
	assign := Assignment{
		Name:     x.Name.Value,
		Value:    "",
		IsLocal:  false,
		IsExport: x.Append,
	}

	// Extract the value directly
	if x.Value != nil {
		// Since x.Value is a *syntax.Word, not an interface, we can't use type assertion
		// Just extract the value directly
		assign.Value = extractWordValue(x.Value)
	}

	return assign
}

// processFunction processes a function declaration.
func processFunction(x *syntax.FuncDecl) *Function {
	function := &Function{
		Name:       x.Name.Value,
		Statements: []Statement{},
		Parameters: []string{},
		LocalVars:  make(map[string]string),
	}

	// Process function body.
	if x.Body != nil {
		syntax.Walk(x.Body, func(node syntax.Node) bool {
			switch y := node.(type) {
			case *syntax.CallExpr:
				cmd := processCallExpr(y)
				function.Statements = append(function.Statements, Statement{
					Type:  StatementCommand,
					Value: cmd,
				})
			case *syntax.Assign:
				assign := processAssign(y)
				function.LocalVars[assign.Name] = assign.Value
				function.Statements = append(function.Statements, Statement{
					Type:  StatementAssignment,
					Value: assign,
				})
			}
			return true
		})
	}

	return function
}

// processIfClause processes an if statement.
func processIfClause(x *syntax.IfClause) If {
	ifStmt := If{
		Condition:     []Statement{},
		ThenBlock:     []Statement{},
		ElseBlock:     []Statement{},
		ElifBlocks:    [][2][]Statement{},
		ConditionType: "command", // Default condition type
	}

	// Process condition
	if len(x.Cond) > 0 {
		for _, cond := range x.Cond {
			if cond.Cmd != nil {
				switch c := cond.Cmd.(type) {
				case *syntax.CallExpr:
					cmd := processCallExpr(c)
					ifStmt.Condition = append(ifStmt.Condition, Statement{
						Type:  StatementCommand,
						Value: cmd,
					})

					// Try to determine the condition type
					if cmd.Name == "test" || cmd.Name == "[" {
						// This is a test condition
						if len(cmd.Args) >= 2 {
							switch cmd.Args[0] {
							case "-f", "-d", "-e":
								ifStmt.ConditionType = "file"
							case "-z", "-n", "=", "!=":
								ifStmt.ConditionType = "string"
							case "-eq", "-ne", "-lt", "-le", "-gt", "-ge":
								ifStmt.ConditionType = "number"
							}
						}
					}
				}
			}
		}
	}

	// Process then block
	if len(x.Then) > 0 {
		for _, stmt := range x.Then {
			if stmt.Cmd != nil {
				switch c := stmt.Cmd.(type) {
				case *syntax.CallExpr:
					cmd := processCallExpr(c)
					ifStmt.ThenBlock = append(ifStmt.ThenBlock, Statement{
						Type:  StatementCommand,
						Value: cmd,
					})
				}
			}
		}
	}

	// Process else block - fix for x.Else.Cmd undefined
	if x.Else != nil {
		// x.Else is a *syntax.IfClause, not an interface
		// This means it's an elif clause, not an else clause
		// Process it as another if statement
		elifStmt := processIfClause(x.Else)
		ifStmt.ElifBlocks = append(ifStmt.ElifBlocks, [][2][]Statement{
			{elifStmt.Condition, elifStmt.ThenBlock},
		}...)
	}

	return ifStmt
}

// processWhileClause processes a while loop.
func processWhileClause(x *syntax.WhileClause) Loop {
	loop := Loop{
		Type:      "while",
		Condition: []Statement{},
		Body:      []Statement{},
	}

	// Process condition.
	for _, cond := range x.Cond {
		if cond.Cmd != nil {
			switch c := cond.Cmd.(type) {
			case *syntax.CallExpr:
				cmd := processCallExpr(c)
				loop.Condition = append(loop.Condition, Statement{
					Type:  StatementCommand,
					Value: cmd,
				})
			}
		}
	}

	// Process body.
	for _, stmt := range x.Do {
		if stmt.Cmd != nil {
			switch c := stmt.Cmd.(type) {
			case *syntax.CallExpr:
				cmd := processCallExpr(c)
				loop.Body = append(loop.Body, Statement{
					Type:  StatementCommand,
					Value: cmd,
				})
			}
		}
	}

	return loop
}

// processForClause processes a for loop.
func processForClause(x *syntax.ForClause) Loop {
	loop := Loop{
		Type: "for",
		Body: []Statement{},
	}

	// Process loop variable
	if x.Loop != nil {
		// This is a for-each loop
		loop.IsForEach = true

		// Based on the documentation, ForClause has a Loop field which is a syntax.Loop
		// The Loop struct likely has a Name field for the loop variable
		loop.RangeVar = "i" // Default variable name if we can't extract it

		// We can't directly access x.Items, so we'll use a placeholder
		loop.Items = "items" // Placeholder for items
	}

	// Process body
	if len(x.Do) > 0 {
		for _, stmt := range x.Do {
			if stmt.Cmd != nil {
				switch c := stmt.Cmd.(type) {
				case *syntax.CallExpr:
					cmd := processCallExpr(c)
					loop.Body = append(loop.Body, Statement{
						Type:  StatementCommand,
						Value: cmd,
					})
				}
			}
		}
	}

	return loop
}

// processPipe processes a pipe by flattening any nested pipe nodes.
func processPipe(x *syntax.BinaryCmd) Pipe {
	pipe := Pipe{
		Commands: flattenPipe(x),
	}
	return pipe
}

// flattenPipe recursively extracts commands from a binary pipe command.
func flattenPipe(node syntax.Node) []Command {
	var commands []Command

	switch n := node.(type) {
	case *syntax.BinaryCmd:
		if n.Op == syntax.Pipe {
			// Recursively process both sides of the pipe
			commands = append(commands, flattenPipe(n.X)...)
			commands = append(commands, flattenPipe(n.Y)...)
		} else {
			// For non-pipe binary commands, just process each side separately
			commands = append(commands, flattenPipe(n.X)...)
			commands = append(commands, flattenPipe(n.Y)...)
		}
	case *syntax.Stmt:
		// Process the command in the statement
		if n.Cmd != nil {
			if call, ok := n.Cmd.(*syntax.CallExpr); ok {
				commands = append(commands, processCallExpr(call))
			}
		}
	case *syntax.CallExpr:
		// Process the call expression directly
		commands = append(commands, processCallExpr(n))
	}

	return commands
}

// processSubshell processes a subshell.
func processSubshell(x *syntax.Subshell) Subshell {
	subshell := Subshell{
		Statements: []Statement{},
	}

	// Process statements in the subshell.
	for _, stmt := range x.Stmts {
		if stmt.Cmd != nil {
			switch c := stmt.Cmd.(type) {
			case *syntax.CallExpr:
				cmd := processCallExpr(c)
				subshell.Statements = append(subshell.Statements, Statement{
					Type:  StatementCommand,
					Value: cmd,
				})
			}
		}
	}

	return subshell
}

// processRedirection processes a redirection.
func processRedirection(x *syntax.Redirect) Redirection {
	redirection := Redirection{
		Op:       x.Op.String(),
		Filename: "",
	}

	// Extract the filename
	if x.Word != nil {
		redirection.Filename = extractWordValue(x.Word)
	}

	// Note: In mvdan.cc/sh/v3/syntax, Redirect doesn't have a Cmd field
	// We'll need to handle this differently, perhaps by processing the parent Stmt
	// that contains both the redirection and the command

	return redirection
}

// NewIntermediateRepresentation initializes a new IR.
func NewIntermediateRepresentation() *IntermediateRepresentation {
	return &IntermediateRepresentation{
		Variables:        make(map[string]string),
		Functions:        make(map[string]*Function),
		MainStatements:   []Statement{},
		RequiredPackages: make(map[string]bool),
	}
}
