package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmcarbo/rhesis/internal/generator"
	"github.com/jmcarbo/rhesis/internal/script"
)

func TestEndToEndPresentationGeneration(t *testing.T) {
	scriptContent := `# Integration Test

Default time: 5

## Test Slide 1

Duration: 8

This is test content

---

This is the transcription for slide 1

## Test Slide 2

Duration: 6

More test content

---

This is the transcription for slide 2`

	tmpDir, err := os.MkdirTemp("", "rhesis_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	scriptFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to write script file: %v", err)
	}

	parsedScript, err := script.ParseScript(scriptFile)
	if err != nil {
		t.Fatalf("Failed to parse script: %v", err)
	}

	gen := generator.NewHTMLGenerator()
	outputFile := filepath.Join(tmpDir, "presentation.html")

	err = gen.GeneratePresentation(parsedScript, outputFile, "modern")
	if err != nil {
		t.Fatalf("Failed to generate presentation: %v", err)
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Expected output file to be created")
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	html := string(content)

	expectedContent := []string{
		"Integration Test",
		"Test Slide 1",
		"Test Slide 2",
		"This is test content",
		"More test content",
		"This is the transcription for slide 1",
		"This is the transcription for slide 2",
		"data-duration=\"8\"",
		"data-duration=\"6\"",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(html, expected) {
			t.Errorf("Expected %s in generated HTML", expected)
		}
	}

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Expected valid HTML document")
	}
}

func TestEndToEndWithImage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rhesis_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	imageFile := filepath.Join(tmpDir, "test.png")
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x01, 0x00, 0x00, 0x00, 0x00, 0x37, 0x6E, 0xF9,
		0x24, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9C, 0x62, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x01, 0xE2, 0x21, 0xBC, 0x33, 0x00, 0x00,
		0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42,
		0x60, 0x82,
	}
	if err := os.WriteFile(imageFile, pngData, 0644); err != nil {
		t.Fatalf("Failed to write image file: %v", err)
	}

	scriptContent := `# Test with Image

## Image Slide

Duration: 10

Image: ` + imageFile + `

---

This slide contains an image`

	scriptFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to write script file: %v", err)
	}

	parsedScript, err := script.ParseScript(scriptFile)
	if err != nil {
		t.Fatalf("Failed to parse script: %v", err)
	}

	gen := generator.NewHTMLGenerator()
	outputFile := filepath.Join(tmpDir, "presentation.html")

	err = gen.GeneratePresentation(parsedScript, outputFile, "modern")
	if err != nil {
		t.Fatalf("Failed to generate presentation: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	html := string(content)

	if !strings.Contains(html, "data:image/png;base64,") {
		t.Error("Expected base64 encoded image in HTML")
	}
}

func TestFullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full workflow test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "rhesis_workflow_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	scriptContent := `# Workflow Test

Default time: 3

## Quick Test

Duration: 2

Fast test content

---

This is a quick test`

	scriptFile := filepath.Join(tmpDir, "workflow.md")
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to write script file: %v", err)
	}

	parsedScript, err := script.ParseScript(scriptFile)
	if err != nil {
		t.Fatalf("Failed to parse script: %v", err)
	}

	gen := generator.NewHTMLGenerator()
	outputFile := filepath.Join(tmpDir, "workflow.html")

	err = gen.GeneratePresentation(parsedScript, outputFile, "modern")
	if err != nil {
		t.Fatalf("Failed to generate presentation: %v", err)
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Expected HTML file to be created")
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read HTML file: %v", err)
	}

	html := string(content)
	if !strings.Contains(html, "Workflow Test") {
		t.Error("Expected presentation title in HTML")
	}

	if !strings.Contains(html, "data-ready") {
		t.Error("Expected data-ready attribute for player integration")
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		scriptPath  string
		outputPath  string
		expectError bool
	}{
		{
			name:        "nonexistent script",
			scriptPath:  "nonexistent.md",
			outputPath:  "output.html",
			expectError: true,
		},
		{
			name:        "invalid output path",
			scriptPath:  "example.md",
			outputPath:  "/invalid/path/output.html",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.scriptPath != "nonexistent.md" {
				if _, err := os.Stat(tt.scriptPath); os.IsNotExist(err) {
					t.Skip("Skipping test - example.md not found")
				}
			}

			var parsedScript *script.Script
			var err error

			if tt.scriptPath == "nonexistent.md" {
				parsedScript, err = script.ParseScript(tt.scriptPath)
				if !tt.expectError && err != nil {
					t.Errorf("Unexpected error parsing script: %v", err)
					return
				}
				if tt.expectError && err == nil {
					t.Error("Expected error parsing script")
					return
				}
				if err != nil {
					return
				}
			} else {
				parsedScript, err = script.ParseScript(tt.scriptPath)
				if err != nil {
					t.Errorf("Failed to parse script: %v", err)
					return
				}
			}

			gen := generator.NewHTMLGenerator()
			err = gen.GeneratePresentation(parsedScript, tt.outputPath, "modern")

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestMarkdownFormatVariations(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		validate    func(t *testing.T, s *script.Script)
	}{
		{
			name: "code blocks in slides",
			content: `# Code Presentation

## JavaScript Example

Duration: 10

Here's a JavaScript function:

` + "```javascript" + `
function greet(name) {
    return "Hello, " + name + "!";
}
` + "```" + `

---

This example shows a simple greeting function in JavaScript.

## Python Example

` + "```python" + `
def greet(name):
    return f"Hello, {name}!"
` + "```" + `

---

The same function in Python uses f-strings.`,
			expectError: false,
			validate: func(t *testing.T, s *script.Script) {
				if len(s.Slides) != 2 {
					t.Errorf("Expected 2 slides, got %d", len(s.Slides))
				}
				if !strings.Contains(s.Slides[0].Content, "```javascript") {
					t.Error("Expected JavaScript code block")
				}
				if !strings.Contains(s.Slides[1].Content, "```python") {
					t.Error("Expected Python code block")
				}
			},
		},
		{
			name: "lists and nested content",
			content: `# List Examples

## Unordered Lists

- First item
- Second item
  - Nested item 1
  - Nested item 2
- Third item

---

This slide demonstrates unordered lists with nesting.

## Ordered Lists

1. First step
2. Second step
   a. Sub-step A
   b. Sub-step B
3. Third step

---

Ordered lists can also be nested.`,
			expectError: false,
			validate: func(t *testing.T, s *script.Script) {
				if !strings.Contains(s.Slides[0].Content, "- Nested item 1") {
					t.Error("Expected nested list items")
				}
				if !strings.Contains(s.Slides[1].Content, "1. First step") {
					t.Error("Expected ordered list")
				}
			},
		},
		{
			name: "markdown formatting",
			content: `# Formatted Text

## Text Styles

**Bold text** and *italic text* and ***bold italic***

> This is a blockquote
> spanning multiple lines

[Link to example](https://example.com)

---

Various text formatting options are supported.`,
			expectError: false,
			validate: func(t *testing.T, s *script.Script) {
				content := s.Slides[0].Content
				if !strings.Contains(content, "**Bold text**") {
					t.Error("Expected bold markdown")
				}
				if !strings.Contains(content, "*italic text*") {
					t.Error("Expected italic markdown")
				}
				if !strings.Contains(content, "> This is a blockquote") {
					t.Error("Expected blockquote")
				}
				if !strings.Contains(content, "[Link to example]") {
					t.Error("Expected markdown link")
				}
			},
		},
		{
			name: "presentation without metadata",
			content: `# Simple Presentation

## First Slide

Content without any duration or metadata

---

Just transcription

## Second Slide

More content`,
			expectError: false,
			validate: func(t *testing.T, s *script.Script) {
				if s.DefaultTime != 10 {
					t.Errorf("Expected default time 10, got %d", s.DefaultTime)
				}
				if s.Slides[0].Duration != 10 {
					t.Errorf("Expected slide duration 10, got %d", s.Slides[0].Duration)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rhesis_md_test_*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			scriptFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(scriptFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write script file: %v", err)
			}

			parsedScript, err := script.ParseScript(scriptFile)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, parsedScript)
			}

			// Also test HTML generation
			gen := generator.NewHTMLGenerator()
			outputFile := filepath.Join(tmpDir, "output.html")
			if err := gen.GeneratePresentation(parsedScript, outputFile, "modern"); err != nil {
				t.Errorf("Failed to generate HTML: %v", err)
			}
		})
	}
}
