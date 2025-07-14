package script

import (
	"os"
	"path/filepath"
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
			content: `title: "Test Presentation"
duration: 60
default_time: 5
slides:
  - title: "Slide 1"
    content: "Content 1"
    transcription: "Transcription 1"
    duration: 10
  - title: "Slide 2"
    content: "Content 2"
    transcription: "Transcription 2"`,
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
			content: `title: "Test"
slides:
  - title: "Slide 1"
    content: "Content 1"
    transcription: "Transcription 1"`,
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
			name:        "invalid yaml",
			content:     `invalid: yaml: content`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test*.yaml")
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
	content := `title: "Test"
slides:
  - title: "Slide 1"
    image: "./test/image.png"
    transcription: "Test transcription"`

	tmpFile, err := os.CreateTemp("", "test*.yaml")
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
	_, err := ParseScript("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}
