package generator

import (
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

func TestD2Extension(t *testing.T) {
	// Create a markdown processor with D2 extension
	d2Ext, err := NewD2Extension()
	if err != nil {
		t.Fatalf("Failed to create D2 extension: %v", err)
	}

	md := goldmark.New(
		goldmark.WithExtensions(d2Ext),
	)

	tests := []struct {
		name     string
		markdown string
		contains []string
		notContains []string
	}{
		{
			name: "D2 code block conversion",
			markdown: `# Test
Here is a D2 diagram:

` + "```d2\nx -> y\n```",
			contains: []string{
				"<div class=\"d2-diagram\">",
				"<img src=\"data:image/svg+xml;base64,",
				"alt=\"D2 Diagram\"",
			},
			notContains: []string{
				"```d2",
				"x -> y",
			},
		},
		{
			name: "Non-D2 code block unchanged",
			markdown: `# Test
Here is Python code:

` + "```python\nprint('hello')\n```",
			contains: []string{
				"<pre>",
				"<code",
				"print",
			},
			notContains: []string{
				"d2-diagram",
			},
		},
		{
			name: "Multiple D2 blocks",
			markdown: `# Test
First diagram:

` + "```d2\na -> b\n```" + `

Second diagram:

` + "```d2\nc -> d\n```",
			contains: []string{
				"<div class=\"d2-diagram\">",
			},
		},
		{
			name: "Mixed content",
			markdown: `# Presentation

Some text here.

` + "```d2\nserver -> client: request\n```" + `

More text.

` + "```javascript\nconsole.log('test');\n```",
			contains: []string{
				"<div class=\"d2-diagram\">",
				"<pre>",
				"console.log",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output strings.Builder
			err := md.Convert([]byte(tt.markdown), &output)
			if err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			result := output.String()

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput: %s", want, result)
				}
			}

			for _, notWant := range tt.notContains {
				if strings.Contains(result, notWant) {
					t.Errorf("Expected output not to contain %q, but it did.\nOutput: %s", notWant, result)
				}
			}
		})
	}
}

func TestD2ExtensionErrorHandling(t *testing.T) {
	d2Ext, err := NewD2Extension()
	if err != nil {
		t.Fatalf("Failed to create D2 extension: %v", err)
	}

	md := goldmark.New(
		goldmark.WithExtensions(d2Ext),
	)

	// Test with invalid D2 syntax - should keep as code block
	markdown := `# Test

` + "```d2\ninvalid -> -> syntax\n```"

	var output strings.Builder
	err = md.Convert([]byte(markdown), &output)
	if err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	result := output.String()
	
	// The extension should gracefully handle errors and keep the original code block
	if !strings.Contains(result, "<pre>") || !strings.Contains(result, "<code") {
		t.Error("Expected invalid D2 syntax to be kept as code block")
	}
}