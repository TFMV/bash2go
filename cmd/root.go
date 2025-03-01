package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TFMV/bash2go/compiler"
	"github.com/TFMV/bash2go/generator"
	"github.com/TFMV/bash2go/parser"
	"github.com/spf13/cobra"
)

var (
	outputFile string
	rootCmd    = &cobra.Command{
		Use:   "bash2go",
		Short: "bash2go is a tool that translates Bash scripts into Go programs",
		Long: `bash2go is a tool that translates Bash scripts into Go programs,
producing a compiled binary that replaces shell scripts. This provides better 
performance, portability, and maintainability than traditional shell scripts.`,
	}
)

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add convert command
	convertCmd := &cobra.Command{
		Use:   "convert [bash script]",
		Short: "Convert a Bash script to Go source code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return convertBashToGo(args[0], outputFile, false)
		},
	}
	convertCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output Go file (required)")
	convertCmd.MarkFlagRequired("output")
	rootCmd.AddCommand(convertCmd)

	// Add build command
	buildCmd := &cobra.Command{
		Use:   "build [bash script]",
		Short: "Convert a Bash script to Go and compile it to a binary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return convertBashToGo(args[0], outputFile, true)
		},
	}
	buildCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output binary name (required)")
	buildCmd.MarkFlagRequired("output")
	rootCmd.AddCommand(buildCmd)
}

// convertBashToGo converts a Bash script to Go code and optionally compiles it
func convertBashToGo(inputScript, outputFile string, shouldCompile bool) error {
	fmt.Printf("Converting %s to Go", inputScript)
	if shouldCompile {
		fmt.Printf(" and compiling to %s\n", outputFile)
	} else {
		fmt.Printf(" and saving to %s\n", outputFile)
	}

	// Parse the Bash script
	result, err := parser.ParseBashScript(inputScript)
	if err != nil {
		return fmt.Errorf("failed to parse Bash script: %v", err)
	}

	// Build intermediate representation
	ir, err := parser.BuildIR(result)
	if err != nil {
		return fmt.Errorf("failed to build intermediate representation: %v", err)
	}

	// Generate Go code
	generator := generator.NewGoCodeGenerator(ir)
	goCode, err := generator.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate Go code: %v", err)
	}

	// Determine output Go file
	var goFile string
	if shouldCompile {
		// Create a temporary file for the Go code
		goFile = filepath.Join(os.TempDir(), filepath.Base(inputScript)+".go")
	} else {
		goFile = outputFile
	}

	// Write the Go code to the file
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		return fmt.Errorf("failed to write Go code to file: %v", err)
	}

	fmt.Printf("Generated Go code saved to %s\n", goFile)

	// Compile if requested
	if shouldCompile {
		fmt.Printf("Compiling %s to %s\n", goFile, outputFile)

		// Build the Go program
		options := compiler.DefaultBuildOptions(outputFile, goFile)
		if err := compiler.BuildGoProgram(options); err != nil {
			return fmt.Errorf("failed to build Go program: %v", err)
		}

		fmt.Printf("Compiled binary saved to %s\n", outputFile)

		// Remove the temporary Go file
		os.Remove(goFile)
	}

	return nil
}
