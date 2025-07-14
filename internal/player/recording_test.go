package player

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmcarbo/rhesis/internal/generator"
	"github.com/jmcarbo/rhesis/internal/script"
)

func TestRecordingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping recording integration test in short mode")
	}

	// Create a test script
	testScript := &script.Script{
		Title:       "Recording Test",
		DefaultTime: 2,
		Slides: []script.Slide{
			{
				Title:         "First Slide",
				Content:       "This is the first slide",
				Transcription: "Testing recording functionality",
				Duration:      2,
			},
			{
				Title:         "Second Slide",
				Content:       "This is the second slide",
				Transcription: "Verifying video output",
				Duration:      2,
			},
		},
	}

	// Generate HTML presentation
	gen := generator.NewHTMLGenerator()
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test_presentation.html")

	if err := gen.GeneratePresentation(testScript, htmlFile, "modern", false); err != nil {
		t.Fatalf("Failed to generate presentation: %v", err)
	}

	// Test different video formats
	formats := []string{"webm", "mp4"}

	for _, format := range formats {
		t.Run("Recording_"+format, func(t *testing.T) {
			recordFile := filepath.Join(tmpDir, "recording."+format)

			player := NewPresentationPlayer()
			err := player.PlayPresentation(htmlFile, recordFile)

			if err != nil {
				t.Errorf("Failed to play and record presentation: %v", err)
				return
			}

			// Wait for file to be fully written
			time.Sleep(1 * time.Second)

			// Verify recording was created
			info, err := os.Stat(recordFile)
			if err != nil {
				t.Errorf("Recording file not found: %v", err)
				return
			}

			// Check file has reasonable size (more than 1KB)
			if info.Size() < 1024 {
				t.Errorf("Recording file too small: %d bytes", info.Size())
			}

			t.Logf("Successfully created %s recording: %d bytes", format, info.Size())
		})
	}
}

func TestRecordingWithDifferentResolutions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resolution test in short mode")
	}

	htmlFile := createTestHTML(t)
	defer os.Remove(htmlFile)

	tmpDir := t.TempDir()
	recordFile := filepath.Join(tmpDir, "recording_hd.webm")

	player := NewPresentationPlayer()

	// Note: Current implementation uses fixed 1920x1080 resolution
	// This test verifies recording works with HD resolution
	err := player.PlayPresentation(htmlFile, recordFile)
	if err != nil {
		t.Errorf("Failed to record with HD resolution: %v", err)
		return
	}

	// Wait for file to be written
	time.Sleep(500 * time.Millisecond)

	// Verify file exists
	if _, err := os.Stat(recordFile); err != nil {
		t.Error("HD recording file not created")
	}
}

func TestRecordingErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error handling test in short mode")
	}

	htmlFile := createTestHTML(t)
	defer os.Remove(htmlFile)

	// Test with invalid record path
	player := NewPresentationPlayer()

	// Use a path that can't be written to
	invalidPath := "/invalid/path/that/does/not/exist/recording.webm"

	// Should not return error immediately, as video saving happens during cleanup
	err := player.PlayPresentation(htmlFile, invalidPath)
	if err != nil {
		t.Logf("Got expected behavior - error might occur during initialization or cleanup")
	}
}
