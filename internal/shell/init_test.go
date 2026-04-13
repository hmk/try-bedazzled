package shell

import (
	"strings"
	"testing"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		env  string
		want ShellType
	}{
		{"/bin/bash", Bash},
		{"/bin/zsh", Bash}, // Zsh uses same syntax as Bash
		{"/usr/bin/fish", Fish},
		{"/opt/homebrew/bin/fish", Fish},
		{"", Bash}, // default
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			got := DetectShell(tt.env)
			if got != tt.want {
				t.Errorf("DetectShell(%q) = %v, want %v", tt.env, got, tt.want)
			}
		})
	}
}

func TestGenerateInitBash(t *testing.T) {
	output := GenerateInit(Bash, "/usr/local/bin/try", "/home/user/src/tries")

	// Should define a function called try()
	if !strings.Contains(output, "try() {") {
		t.Error("missing try() function definition")
	}
	// Should use the binary path
	if !strings.Contains(output, "'/usr/local/bin/try' exec --path '/home/user/src/tries'") {
		t.Error("missing binary invocation with path")
	}
	// Should redirect stderr to /dev/tty
	if !strings.Contains(output, "2>/dev/tty") {
		t.Error("missing stderr redirect to /dev/tty")
	}
	// Should eval the output
	if !strings.Contains(output, `eval "$out"`) {
		t.Error("missing eval")
	}
}

func TestGenerateInitFish(t *testing.T) {
	output := GenerateInit(Fish, "/usr/local/bin/try", "/home/user/src/tries")

	// Should use fish function syntax
	if !strings.Contains(output, "function try") {
		t.Error("missing fish function definition")
	}
	if !strings.Contains(output, "set -l out") {
		t.Error("missing fish local variable")
	}
	if !strings.Contains(output, "$argv") {
		t.Error("missing fish $argv")
	}
	if !strings.Contains(output, "eval $out") {
		t.Error("missing fish eval")
	}
	if !strings.Contains(output, "end") {
		t.Error("missing fish end")
	}
}

func TestGenerateInitEscapesPaths(t *testing.T) {
	// Path with special characters
	output := GenerateInit(Bash, "/path/to/my try", "/home/user/my tries")
	if !strings.Contains(output, "'/path/to/my try'") {
		t.Error("binary path not escaped")
	}
	if !strings.Contains(output, "'/home/user/my tries'") {
		t.Error("tries path not escaped")
	}
}
