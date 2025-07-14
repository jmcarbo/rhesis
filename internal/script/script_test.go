package script

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseScript(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		expected    *Script
	}{
		{
			name: "valid script",
			content: `# Test Presentation

Duration: 60
Default time: 5

## Slide 1

Duration: 10

Content 1

---

Transcription 1

## Slide 2

Content 2

---

Transcription 2`,
			expectError: false,
			expected: &Script{
				Title:       "Test Presentation",
				Duration:    60,
				DefaultTime: 5,
				Slides: []Slide{
					{Title: "Slide 1", Content: "Content 1", Transcription: "Transcription 1", Duration: 10},
					{Title: "Slide 2", Content: "Content 2", Transcription: "Transcription 2", Duration: 5},
				},
			},
		},
		{
			name: "script with default time",
			content: `# Test

## Slide 1

Content 1

---

Transcription 1`,
			expectError: false,
			expected: &Script{
				Title:       "Test",
				DefaultTime: 10,
				Slides: []Slide{
					{Title: "Slide 1", Content: "Content 1", Transcription: "Transcription 1", Duration: 10},
				},
			},
		},
		{
			name:        "invalid markdown - no title",
			content:     `## Slide without presentation title`,
			expectError: true,
		},
		{
			name: "multiline content",
			content: `# Test

## Slide 1

Line 1
Line 2
Line 3

---

Test transcription`,
			expectError: false,
			expected: &Script{
				Title:       "Test",
				DefaultTime: 10,
				Slides: []Slide{
					{Title: "Slide 1", Content: "Line 1\nLine 2\nLine 3", Transcription: "Test transcription", Duration: 10},
				},
			},
		},
		{
			name: "code blocks in content",
			content: `# Code Example

## Code Slide

Here's some code:

` + "```go" + `
func main() {
    fmt.Println("Hello")
}
` + "```" + `

---

This shows Go code`,
			expectError: false,
			expected: &Script{
				Title:       "Code Example",
				DefaultTime: 10,
				Slides: []Slide{
					{
						Title: "Code Slide",
						Content: "Here's some code:\n\n```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
						Transcription: "This shows Go code",
						Duration: 10,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test*.md")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			result, err := ParseScript(tmpFile.Name())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.Title != tt.expected.Title {
				t.Errorf("Expected title %s, got %s", tt.expected.Title, result.Title)
			}

			if result.DefaultTime != tt.expected.DefaultTime {
				t.Errorf("Expected default time %d, got %d", tt.expected.DefaultTime, result.DefaultTime)
			}

			if len(result.Slides) != len(tt.expected.Slides) {
				t.Errorf("Expected %d slides, got %d", len(tt.expected.Slides), len(result.Slides))
				return
			}

			for i, slide := range result.Slides {
				expected := tt.expected.Slides[i]
				if slide.Title != expected.Title {
					t.Errorf("Slide %d: expected title %s, got %s", i, expected.Title, slide.Title)
				}
				if slide.Content != expected.Content {
					t.Errorf("Slide %d: expected content %s, got %s", i, expected.Content, slide.Content)
				}
				if slide.Transcription != expected.Transcription {
					t.Errorf("Slide %d: expected transcription %s, got %s", i, expected.Transcription, slide.Transcription)
				}
				if slide.Duration != expected.Duration {
					t.Errorf("Slide %d: expected duration %d, got %d", i, expected.Duration, slide.Duration)
				}
			}
		})
	}
}

func TestScriptGetTotalDuration(t *testing.T) {
	script := &Script{
		Slides: []Slide{
			{Duration: 10},
			{Duration: 15},
			{Duration: 5},
		},
	}

	expected := 30 * time.Second
	result := script.GetTotalDuration()

	if result != expected {
		t.Errorf("Expected total duration %v, got %v", expected, result)
	}
}

func TestParseScriptWithImage(t *testing.T) {
	content := `# Test

## Slide 1

Image: ./test/image.png

---

Test transcription`

	tmpFile, err := os.CreateTemp("", "test*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	result, err := ParseScript(tmpFile.Name())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	expected := filepath.Clean("./test/image.png")
	if result.Slides[0].Image != expected {
		t.Errorf("Expected image path %s, got %s", expected, result.Slides[0].Image)
	}
}

func TestParseScriptNonexistentFile(t *testing.T) {
	_, err := ParseScript("nonexistent.md")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParseScriptEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		validation  func(t *testing.T, s *Script)
	}{
		{
			name: "empty slides",
			content: `# Empty Presentation

Duration: 60`,
			expectError: false,
			validation: func(t *testing.T, s *Script) {
				if len(s.Slides) != 0 {
					t.Errorf("Expected 0 slides, got %d", len(s.Slides))
				}
			},
		},
		{
			name: "very long title",
			content: `# ` + strings.Repeat("Very Long Title ", 100) + `

## Slide 1

Content`,
			expectError: false,
			validation: func(t *testing.T, s *Script) {
				if len(s.Title) < 1000 {
					t.Error("Expected very long title")
				}
			},
		},
		{
			name: "slide with only title",
			content: `# Minimal

## Only Title`,
			expectError: false,
			validation: func(t *testing.T, s *Script) {
				if s.Slides[0].Title != "Only Title" {
					t.Error("Expected title to be set")
				}
				if s.Slides[0].Content != "" {
					t.Error("Expected empty content")
				}
				if s.Slides[0].Transcription != "" {
					t.Error("Expected empty transcription")
				}
				if s.Slides[0].Duration != 10 {
					t.Errorf("Expected default duration 10, got %d", s.Slides[0].Duration)
				}
			},
		},
		{
			name: "mixed duration values",
			content: `# Mixed Durations

Default time: 15

## Slide 1

Duration: 5

Content

## Slide 2

Content

## Slide 3

Duration: 20

Content`,
			expectError: false,
			validation: func(t *testing.T, s *Script) {
				if s.Slides[0].Duration != 5 {
					t.Errorf("Slide 1: expected duration 5, got %d", s.Slides[0].Duration)
				}
				if s.Slides[1].Duration != 15 {
					t.Errorf("Slide 2: expected duration 15, got %d", s.Slides[1].Duration)
				}
				if s.Slides[2].Duration != 20 {
					t.Errorf("Slide 3: expected duration 20, got %d", s.Slides[2].Duration)
				}
			},
		},
		{
			name: "bullet points and formatting",
			content: `# Formatted

## Slide with Bullets

• Point 1
• Point 2
• Point 3

**Bold text** and *italic*

---

Transcription with formatting`,
			expectError: false,
			validation: func(t *testing.T, s *Script) {
				if !strings.Contains(s.Slides[0].Content, "• Point 1") {
					t.Error("Expected bullet points in content")
				}
				if !strings.Contains(s.Slides[0].Content, "**Bold text**") {
					t.Error("Expected markdown formatting in content")
				}
			},
		},
		{
			name: "no transcription",
			content: `# No Transcription

## Slide 1

Just content, no transcription`,
			expectError: false,
			validation: func(t *testing.T, s *Script) {
				if s.Slides[0].Content != "Just content, no transcription" {
					t.Errorf("Expected content, got %s", s.Slides[0].Content)
				}
				if s.Slides[0].Transcription != "" {
					t.Errorf("Expected empty transcription, got %s", s.Slides[0].Transcription)
				}
			},
		},
		{
			name: "multiple horizontal rules",
			content: `# Test

## Slide 1

Content

---

Transcription

---

This should be ignored

## Slide 2

Content 2`,
			expectError: false,
			validation: func(t *testing.T, s *Script) {
				if len(s.Slides) != 2 {
					t.Errorf("Expected 2 slides, got %d", len(s.Slides))
				}
				// First slide should have transcription until the next slide
				// The parser skips the second --- line, so it's not included
				expected := "Transcription\n\n\nThis should be ignored"
				if s.Slides[0].Transcription != expected {
					t.Errorf("Expected transcription %q, got %q", expected, s.Slides[0].Transcription)
				}
			},
		},
		{
			name: "special characters",
			content: `# Test © 2024

## Slide™ with "quotes"

Content with 'quotes' and ñ ü é

---

Transcription with emoji 🎉`,
			expectError: false,
			validation: func(t *testing.T, s *Script) {
				if !strings.Contains(s.Title, "©") {
					t.Error("Expected copyright symbol in title")
				}
				if !strings.Contains(s.Slides[0].Title, "™") {
					t.Error("Expected trademark symbol in slide title")
				}
				if !strings.Contains(s.Slides[0].Transcription, "🎉") {
					t.Error("Expected emoji in transcription")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test*.md")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			result, err := ParseScript(tmpFile.Name())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validation != nil {
				tt.validation(t, result)
			}
		})
	}
}

func TestParseScriptLargeFile(t *testing.T) {
	// Test with many slides
	var content strings.Builder
	content.WriteString("# Large Presentation\n\n")
	
	numSlides := 100
	for i := 1; i <= numSlides; i++ {
		content.WriteString(fmt.Sprintf("## Slide %d\n\n", i))
		content.WriteString(fmt.Sprintf("Duration: %d\n\n", i%10+5))
		content.WriteString(fmt.Sprintf("Content for slide %d\n", i))
		content.WriteString("---\n")
		content.WriteString(fmt.Sprintf("Transcription for slide %d\n\n", i))
	}

	tmpFile, err := os.CreateTemp("", "test_large*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content.String()); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	result, err := ParseScript(tmpFile.Name())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if len(result.Slides) != numSlides {
		t.Errorf("Expected %d slides, got %d", numSlides, len(result.Slides))
	}

	// Check a few slides
	if result.Slides[0].Title != "Slide 1" {
		t.Errorf("Expected first slide title 'Slide 1', got %s", result.Slides[0].Title)
	}
	if result.Slides[numSlides-1].Title != fmt.Sprintf("Slide %d", numSlides) {
		t.Errorf("Expected last slide title 'Slide %d', got %s", numSlides, result.Slides[numSlides-1].Title)
	}
}