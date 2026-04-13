package shell

import "testing"

func TestEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "'hello'"},
		{"hello world", "'hello world'"},
		{"/path/to/dir", "'/path/to/dir'"},
		{"", "''"},

		// Single quote escaping: ' → '"'"'
		{"it's", `'it'"'"'s'`},
		{"foo'bar'baz", `'foo'"'"'bar'"'"'baz'`},

		// Special shell chars should be safely quoted
		{"$HOME", "'$HOME'"},
		{"$(rm -rf /)", "'$(rm -rf /)'"},
		{"`whoami`", "'`whoami`'"},
		{"hello;world", "'hello;world'"},
		{"a && b", "'a && b'"},
		{"a | b", "'a | b'"},
		{"a > b", "'a > b'"},
		{"hello\nworld", "'hello\nworld'"},

		// Unicode
		{"café", "'café'"},
		{"日本語", "'日本語'"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Escape(tt.input)
			if got != tt.want {
				t.Errorf("Escape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
