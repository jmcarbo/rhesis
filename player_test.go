package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewPresentationPlayer(t *testing.T) {
	player := NewPresentationPlayer()
	if player == nil {
		t.Error("Expected player to be created")
	}
}

func TestPresentationPlayerInvalidHTML(t *testing.T) {
	player := NewPresentationPlayer()
	err := player.PlayPresentation("nonexistent.html", "")
	if err == nil {
		t.Error("Expected error for nonexistent HTML file")
	}
}

func createTestHTML(t *testing.T) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Test Presentation</title>
</head>
<body data-ready="true">
    <div id="playBtn">Play</div>
    <script>
        window.isPlaying = false;
        window.totalDuration = 5000;
        
        document.getElementById('playBtn').addEventListener('click', function() {
            window.isPlaying = true;
            setTimeout(function() {
                window.isPlaying = false;
            }, 1000);
        });
    </script>
</body>
</html>`

	tmpFile, err := os.CreateTemp("", "test*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(html); err != nil {
		t.Fatalf("Failed to write HTML: %v", err)
	}
	tmpFile.Close()

	return tmpFile.Name()
}

func TestPlayPresentationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	htmlFile := createTestHTML(t)
	defer os.Remove(htmlFile)

	player := NewPresentationPlayer()
	
	done := make(chan error, 1)
	go func() {
		done <- player.PlayPresentation(htmlFile, "")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Error("Test timed out")
	}
}

func TestPlayPresentationWithRecording(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	htmlFile := createTestHTML(t)
	defer os.Remove(htmlFile)

	recordFile := filepath.Join(os.TempDir(), "test_recording.webm")
	defer os.Remove(recordFile)

	player := NewPresentationPlayer()
	
	done := make(chan error, 1)
	go func() {
		done <- player.PlayPresentation(htmlFile, recordFile)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if _, err := os.Stat(recordFile); os.IsNotExist(err) {
			t.Error("Expected recording file to be created")
		}
	case <-time.After(30 * time.Second):
		t.Error("Test timed out")
	}
}

func TestPresentationPlayerCleanup(t *testing.T) {
	player := NewPresentationPlayer()
	player.cleanup()
}

func TestPresentationPlayerInitializeAndCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	player := NewPresentationPlayer()
	
	err := player.initialize()
	if err != nil {
		t.Errorf("Failed to initialize player: %v", err)
		return
	}

	if player.pw == nil {
		t.Error("Expected playwright to be initialized")
	}
	if player.browser == nil {
		t.Error("Expected browser to be initialized")
	}
	if player.page == nil {
		t.Error("Expected page to be initialized")
	}

	player.cleanup()
}