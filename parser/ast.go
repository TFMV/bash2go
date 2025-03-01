package parser

import (
	"fmt"

	"mvdan.cc/sh/v3/syntax"
)

// IntermediateRepresentation represents the processed AST in a format suitable for Go code generation
type IntermediateRepresentation struct {
	Variables      map[string]string
	Functions      map[string]*Function
	MainStatements []Statement
}

// Function represents a Bash function definition
type Function struct {
	Name       string
	Statements []Statement
}

// StatementType identifies the type of a statement
type StatementType int

const (
	StatementCommand StatementType = iota
	StatementAssignment
	StatementIf
	StatementLoop
	StatementPipe
	StatementSubshell
	StatementFunction
)

// Statement represents a single statement in the Bash script
type Statement struct {
	Type  StatementType
	Value interface{} // Command, Assignment, If, Loop, Pipe, Subshell
}

// Command represents a command execution
type Command struct {
	Name string
	Args []string
}

// Assignment represents a variable assignment
type Assignment struct {
	Name  string
	Value string
}

// If represents an if-then-else statement
type If struct {
	Condition  []Statement
	ThenBlock  []Statement
	ElseBlock  []Statement
	ElifBlocks [][2][]Statement // Array of [condition, block] pairs
}

// Loop represents a loop construct (for, while, until)
type Loop struct {
	Type      string // "for", "while", "until"
	Init      []Statement
	Condition []Statement
	Body      []Statement
}

// Pipe represents a piped command sequence
type Pipe struct {
	Commands []Command
}

// Subshell represents a subshell execution
type Subshell struct {
	Statements []Statement
}

// VisitNode visits each node in the AST and builds the intermediate representation
func VisitNode(ir *IntermediateRepresentation, node syntax.Node) error {
	switch n := node.(type) {
	case *syntax.File:
		// Process file
		for _, stmt := range n.Stmts {
			if err := VisitNode(ir, stmt); err != nil {
				return err
			}
		}
	case *syntax.Stmt:
		// Process statement
		if n.Cmd != nil {
			stmt, err := processCommand(n.Cmd)
			if err != nil {
				return err
			}
			ir.MainStatements = append(ir.MainStatements, stmt)
		}
	default:
		// Handle other node types
	}

	return nil
}

// processCommand processes a command node
func processCommand(cmd syntax.Command) (Statement, error) {
	switch c := cmd.(type) {
	case *syntax.CallExpr:
		// Simple command
		command := Command{
			Name: "",
			Args: []string{},
		}

		if len(c.Args) > 0 {
			// Extract command name
			word := c.Args[0]
			if len(word.Parts) > 0 {
				if lit, ok := word.Parts[0].(*syntax.Lit); ok {
					command.Name = lit.Value
				}
			}

			// Extract arguments
			for i := 1; i < len(c.Args); i++ {
				word := c.Args[i]

				// Handle different types of arguments
				if len(word.Parts) > 0 {
					switch part := word.Parts[0].(type) {
					case *syntax.Lit:
						command.Args = append(command.Args, part.Value)
					case *syntax.ParamExp:
						command.Args = append(command.Args, "$"+part.Param.Value)
					case *syntax.DblQuoted:
						// Handle double-quoted strings
						var quotedValue string
						for _, qpart := range part.Parts {
							if lit, ok := qpart.(*syntax.Lit); ok {
								quotedValue += lit.Value
							}
						}
						command.Args = append(command.Args, quotedValue)
					default:
						command.Args = append(command.Args, fmt.Sprintf("%v", part))
					}
				}
			}
		}

		// Debug output
		fmt.Printf("Parsed command: %s with args: %v\n", command.Name, command.Args)

		return Statement{
			Type:  StatementCommand,
			Value: command,
		}, nil
	case *syntax.BinaryCmd:
		if c.Op == syntax.Pipe {
			// Handle pipe
			return Statement{
				Type:  StatementPipe,
				Value: Pipe{},
			}, nil
		}
		// For now, just create a dummy command for other binary commands
		return Statement{
			Type:  StatementCommand,
			Value: Command{Name: "echo", Args: []string{"Binary command not fully implemented"}},
		}, nil
	case *syntax.Subshell:
		// Handle subshell
		return Statement{
			Type:  StatementSubshell,
			Value: Subshell{Statements: []Statement{}},
		}, nil
	case *syntax.Block:
		// Handle blocks by creating a dummy command for now
		return Statement{
			Type:  StatementCommand,
			Value: Command{Name: "echo", Args: []string{"Block not fully implemented"}},
		}, nil
	case *syntax.IfClause:
		// Handle if statements by creating a dummy command for now
		return Statement{
			Type:  StatementCommand,
			Value: Command{Name: "echo", Args: []string{"If statement not fully implemented"}},
		}, nil
	case *syntax.WhileClause:
		// Handle while loops by creating a dummy command for now
		return Statement{
			Type:  StatementCommand,
			Value: Command{Name: "echo", Args: []string{"While loop not fully implemented"}},
		}, nil
	case *syntax.ForClause:
		// Handle for loops by creating a dummy command for now
		return Statement{
			Type:  StatementCommand,
			Value: Command{Name: "echo", Args: []string{"For loop not fully implemented"}},
		}, nil
	case *syntax.FuncDecl:
		// Handle function declarations by creating a dummy command for now
		return Statement{
			Type:  StatementCommand,
			Value: Command{Name: "echo", Args: []string{"Function declaration not fully implemented"}},
		}, nil
	default:
		return Statement{}, fmt.Errorf("unsupported command type: %T", cmd)
	}
}

// NewIntermediateRepresentation initializes a new IR
func NewIntermediateRepresentation() *IntermediateRepresentation {
	return &IntermediateRepresentation{
		Variables:      make(map[string]string),
		Functions:      make(map[string]*Function),
		MainStatements: []Statement{},
	}
}

// BuildIR builds an intermediate representation from a parsed result
func BuildIR(result *ParseResult) (*IntermediateRepresentation, error) {
	ir := NewIntermediateRepresentation()

	err := VisitNode(ir, result.File)
	if err != nil {
		return nil, err
	}

	return ir, nil
}
