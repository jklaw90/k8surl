package cmd

import (
	"testing"
)

func TestSplitURLs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single https URL",
			input: "https://example.com/path",
			want:  []string{"https://example.com/path"},
		},
		{
			name:  "single http URL",
			input: "http://example.com/path",
			want:  []string{"http://example.com/path"},
		},
		{
			name:  "two concatenated https URLs",
			input: "https://example.com/onehttps://example.com/two",
			want:  []string{"https://example.com/one", "https://example.com/two"},
		},
		{
			name:  "mixed http and https",
			input: "https://example.com/securehttp://example.com/plain",
			want:  []string{"https://example.com/secure", "http://example.com/plain"},
		},
		{
			name:  "three URLs",
			input: "https://a.comhttps://b.comhttp://c.com",
			want:  []string{"https://a.com", "https://b.com", "http://c.com"},
		},
		{
			name:  "no URLs",
			input: "just some text",
			want:  nil,
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitURLs(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("splitURLs(%q) returned %d URLs, want %d: got %v", tt.input, len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitURLs(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}
