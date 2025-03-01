package generator

import (
	"bytes"
	"text/template"
)

// Templates for Go code generation
var (
	// Main program template
	mainTemplate = `package main

import (
	"fmt"
	"os"
	{{- if .RequiresExec }}
	"os/exec"
	{{- end }}
	{{- if .RequiresFilepath }}
	"path/filepath"
	{{- end }}
	{{- if .RequiresStrings }}
	"strings"
	{{- end }}
	{{- range .Imports }}
	{{- if and (ne . "fmt") (ne . "os") (ne . "os/exec") (ne . "path/filepath") (ne . "strings") }}
	"{{ . }}"
	{{- end }}
	{{- end }}
)

{{- range .Functions }}

// {{ .Name }} is a translated Bash function
func {{ .Name }}() {{ if .ReturnValue }}string{{ else }}error{{ end }} {
	{{ .Body }}
}
{{- end }}

func main() {
	var err error
	
	{{- range .Variables }}
	{{ .Name }} := {{ .Value }}
	{{- end }}
	
	{{ .MainBody }}
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
`

	// Command template
	commandTemplate = `cmd := exec.Command("{{ .Name }}"{{ range .Args }}, "{{ . }}"{{ end }})
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}
	fmt.Print(string(output))
`

	// Pipe template
	pipeTemplate = `{{ range $i, $cmd := .Commands }}
	{{ if eq $i 0 }}
	cmd{{ $i }} := exec.Command("{{ $cmd.Name }}"{{ range $cmd.Args }}, "{{ . }}"{{ end }})
	{{ else }}
	cmd{{ $i }} := exec.Command("{{ $cmd.Name }}"{{ range $cmd.Args }}, "{{ . }}"{{ end }})
	{{ end }}
	{{ end }}
	
	{{ range $i, $cmd := .Commands }}
	{{ if ne $i 0 }}
	// Connect previous command's stdout to this command's stdin
	pipe, err := cmd{{ subtract $i 1 }}.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %v", err)
	}
	cmd{{ $i }}.Stdin = pipe
	{{ end }}
	{{ end }}
	
	{{ range $i, $cmd := .Commands }}
	{{ if eq $i (subtract (len $.Commands) 1) }}
	// Set the final command's stdout to current stdout
	cmd{{ $i }}.Stdout = os.Stdout
	{{ end }}
	{{ end }}
	
	// Start all commands
	{{ range $i, $cmd := .Commands }}
	if err := cmd{{ $i }}.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}
	{{ end }}
	
	// Wait for all commands to complete
	{{ range $i, $cmd := .Commands }}
	if err := cmd{{ $i }}.Wait(); err != nil {
		return fmt.Errorf("command failed: %v", err)
	}
	{{ end }}
`

	// Subshell template
	subshellTemplate = `func() error {
		{{ .Statements }}
		return nil
	}()`

	// If statement template
	ifTemplate = `if {{ .Condition }} {
		{{ .ThenBlock }}
	}{{ if .ElseBlock }} else {
		{{ .ElseBlock }}
	}{{ end }}`

	// Loop template
	loopTemplate = `{{ if eq .Type "for" }}
	for {{ .Init }}; {{ .Condition }}; {{ .Update }} {
		{{ .Body }}
	}
	{{ else if eq .Type "while" }}
	for {{ .Condition }} {
		{{ .Body }}
	}
	{{ else if eq .Type "until" }}
	for !{{ .Condition }} {
		{{ .Body }}
	}
	{{ end }}`
)

// ExecuteTemplate executes a template with the given data
func ExecuteTemplate(tmpl string, data interface{}) (string, error) {
	funcMap := template.FuncMap{
		"subtract": func(a, b int) int { return a - b },
	}

	t, err := template.New("template").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
