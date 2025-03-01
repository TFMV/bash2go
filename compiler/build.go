package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// BuildOptions contains options for building the Go code
type BuildOptions struct {
	OutputFile    string // Name of the output binary
	TempDir       string // Temporary directory for intermediate files
	KeepTempFiles bool   // Whether to keep temporary files
	GoFile        string // Path to the generated Go file
}

// DefaultBuildOptions returns default build options
func DefaultBuildOptions(outputFile, goFile string) BuildOptions {
	return BuildOptions{
		OutputFile:    outputFile,
		TempDir:       "", // Will be set to a generated temp dir if empty
		KeepTempFiles: false,
		GoFile:        goFile,
	}
}

// BuildGoProgram compiles a Go source file into a binary
func BuildGoProgram(options BuildOptions) error {
	// Create a temporary directory if not specified
	if options.TempDir == "" {
		tempDir, err := os.MkdirTemp("", "bash2go-")
		if err != nil {
			return fmt.Errorf("failed to create temporary directory: %v", err)
		}
		options.TempDir = tempDir

		// Clean up temporary directory if not keeping temp files
		if !options.KeepTempFiles {
			defer os.RemoveAll(tempDir)
		}
	}

	// Copy or move the Go file to the temp directory
	goFileName := filepath.Base(options.GoFile)
	tempGoFile := filepath.Join(options.TempDir, goFileName)

	if options.GoFile != tempGoFile {
		data, err := os.ReadFile(options.GoFile)
		if err != nil {
			return fmt.Errorf("failed to read Go file: %v", err)
		}

		if err := os.WriteFile(tempGoFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write Go file to temp directory: %v", err)
		}
	}

	// Initialize a Go module
	cmd := exec.Command("go", "mod", "init", "bash2go_output")
	cmd.Dir = options.TempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to initialize Go module: %v\n%s", err, output)
	}

	// Download dependencies
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = options.TempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to tidy Go module: %v\n%s", err, output)
	}

	// Build the binary
	cmd = exec.Command("go", "build", "-o", options.OutputFile, goFileName)
	cmd.Dir = options.TempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to build Go program: %v\n%s", err, output)
	}

	// Move the binary to the current directory if it's not already there
	if filepath.Dir(options.OutputFile) == "." {
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %v", err)
		}

		outputPath := filepath.Join(options.TempDir, options.OutputFile)
		finalPath := filepath.Join(currentDir, options.OutputFile)

		data, err := os.ReadFile(outputPath)
		if err != nil {
			return fmt.Errorf("failed to read output binary: %v", err)
		}

		if err := os.WriteFile(finalPath, data, 0755); err != nil {
			return fmt.Errorf("failed to write output binary: %v", err)
		}
	}

	return nil
}
