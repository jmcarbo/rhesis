package generator

import (
	"os"
	"strings"
	"testing"

	"github.com/jmcarbo/rhesis/internal/script"
)

func TestNewHTMLGenerator(t *testing.T) {
	generator := NewHTMLGenerator()
	if generator == nil {
		t.Error("Expected generator to be created")
	}
	if generator.template == nil {
		t.Error("Expected template to be initialized")
	}
}

func TestGeneratePresentation(t *testing.T) {
	testScript := &script.Script{
		Title: "Test Presentation",
		Slides: []script.Slide{
			{
				Title:         "Slide 1",
				Content:       "Content 1",
				Transcription: "Transcription 1",
				Duration:      10,
			},
			{
				Title:         "Slide 2",
				Content:       "Content 2",
				Transcription: "Transcription 2",
				Duration:      15,
			},
		},
	}

	generator := NewHTMLGenerator()
	tmpFile, err := os.CreateTemp("", "test*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	err = generator.GeneratePresentation(testScript, tmpFile.Name())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	html := string(content)

	if !strings.Contains(html, "Test Presentation") {
		t.Error("Expected presentation title in HTML")
	}

	if !strings.Contains(html, "Slide 1") {
		t.Error("Expected slide 1 title in HTML")
	}

	if !strings.Contains(html, "Slide 2") {
		t.Error("Expected slide 2 title in HTML")
	}

	if !strings.Contains(html, "Content 1") {
		t.Error("Expected slide 1 content in HTML")
	}

	if !strings.Contains(html, "Transcription 1") {
		t.Error("Expected slide 1 transcription in HTML")
	}

	if !strings.Contains(html, "data-duration=\"10\"") {
		t.Error("Expected slide 1 duration in HTML")
	}

	if !strings.Contains(html, "data-duration=\"15\"") {
		t.Error("Expected slide 2 duration in HTML")
	}
}

func TestGeneratePresentationWithImage(t *testing.T) {
	tmpImageFile, err := os.CreateTemp("", "test*.png")
	if err != nil {
		t.Fatalf("Failed to create temp image file: %v", err)
	}
	defer os.Remove(tmpImageFile.Name())

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
	if _, err := tmpImageFile.Write(pngData); err != nil {
		t.Fatalf("Failed to write PNG data: %v", err)
	}
	tmpImageFile.Close()

	testScript := &script.Script{
		Title: "Test with Image",
		Slides: []script.Slide{
			{
				Title:         "Slide with Image",
				Image:         tmpImageFile.Name(),
				Transcription: "This slide has an image",
				Duration:      10,
			},
		},
	}

	generator := NewHTMLGenerator()
	tmpFile, err := os.CreateTemp("", "test*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	err = generator.GeneratePresentation(testScript, tmpFile.Name())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	html := string(content)

	if !strings.Contains(html, "data:image/png;base64,") {
		t.Errorf("Expected base64 encoded image in HTML. HTML content: %s", html)
	}
}

func TestBase64Encode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "simple string",
			input:    []byte("hello"),
			expected: "aGVsbG8=",
		},
		{
			name:     "empty string",
			input:    []byte(""),
			expected: "",
		},
		{
			name:     "binary data",
			input:    []byte{0x00, 0x01, 0x02},
			expected: "AAEC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := base64Encode(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestProcessSlides(t *testing.T) {
	generator := NewHTMLGenerator()
	slides := []script.Slide{
		{Title: "Slide 1", Content: "Content 1"},
		{Title: "Slide 2", Content: "Content 2"},
	}

	result := generator.processSlides(slides)

	if len(result) != 2 {
		t.Errorf("Expected 2 processed slides, got %d", len(result))
	}

	for i, slide := range result {
		if slide.Index != i {
			t.Errorf("Expected slide %d to have index %d, got %d", i, i, slide.Index)
		}
		if slide.Title != slides[i].Title {
			t.Errorf("Expected slide %d title %s, got %s", i, slides[i].Title, slide.Title)
		}
	}
}

func TestGeneratePresentationInvalidPath(t *testing.T) {
	testScript := &script.Script{
		Title:  "Test",
		Slides: []script.Slide{{Title: "Slide 1"}},
	}

	generator := NewHTMLGenerator()
	err := generator.GeneratePresentation(testScript, "/invalid/path/file.html")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}
