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
	scriptContent := `title: "Integration Test"
default_time: 5
slides:
  - title: "Test Slide 1"
    content: "This is test content"
    transcription: "This is the transcription for slide 1"
    duration: 8
  - title: "Test Slide 2"
    content: "More test content"
    transcription: "This is the transcription for slide 2"
    duration: 6`

	tmpDir, err := os.MkdirTemp("", "rhesis_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	scriptFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to write script file: %v", err)
	}

	parsedScript, err := script.ParseScript(scriptFile)
	if err != nil {
		t.Fatalf("Failed to parse script: %v", err)
	}

	gen := generator.NewHTMLGenerator()
	outputFile := filepath.Join(tmpDir, "presentation.html")

	err = gen.GeneratePresentation(parsedScript, outputFile)
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

	scriptContent := `title: "Test with Image"
slides:
  - title: "Image Slide"
    image: "` + imageFile + `"
    transcription: "This slide contains an image"
    duration: 10`

	scriptFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to write script file: %v", err)
	}

	parsedScript, err := script.ParseScript(scriptFile)
	if err != nil {
		t.Fatalf("Failed to parse script: %v", err)
	}

	gen := generator.NewHTMLGenerator()
	outputFile := filepath.Join(tmpDir, "presentation.html")

	err = gen.GeneratePresentation(parsedScript, outputFile)
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

	scriptContent := `title: "Workflow Test"
default_time: 3
slides:
  - title: "Quick Test"
    content: "Fast test content"
    transcription: "This is a quick test"
    duration: 2`

	scriptFile := filepath.Join(tmpDir, "workflow.yaml")
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to write script file: %v", err)
	}

	parsedScript, err := script.ParseScript(scriptFile)
	if err != nil {
		t.Fatalf("Failed to parse script: %v", err)
	}

	gen := generator.NewHTMLGenerator()
	outputFile := filepath.Join(tmpDir, "workflow.html")

	err = gen.GeneratePresentation(parsedScript, outputFile)
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
			scriptPath:  "nonexistent.yaml",
			outputPath:  "output.html",
			expectError: true,
		},
		{
			name:        "invalid output path",
			scriptPath:  "example.yaml",
			outputPath:  "/invalid/path/output.html",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.scriptPath != "nonexistent.yaml" {
				if _, err := os.Stat(tt.scriptPath); os.IsNotExist(err) {
					t.Skip("Skipping test - example.yaml not found")
				}
			}

			var parsedScript *script.Script
			var err error

			if tt.scriptPath == "nonexistent.yaml" {
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
			err = gen.GeneratePresentation(parsedScript, tt.outputPath)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
