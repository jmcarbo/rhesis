package audio

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewElevenLabsGenerator(t *testing.T) {
	config := ElevenLabsConfig{
		APIKey: "test-key",
	}

	gen := NewElevenLabsGenerator(config)

	if gen.config.APIKey != "test-key" {
		t.Errorf("Expected API key to be 'test-key', got '%s'", gen.config.APIKey)
	}

	if gen.config.VoiceID == "" {
		t.Error("Expected default voice ID to be set")
	}

	if gen.config.ModelID == "" {
		t.Error("Expected default model ID to be set")
	}
}

func TestGenerateAudioNoAPIKey(t *testing.T) {
	config := ElevenLabsConfig{}
	gen := NewElevenLabsGenerator(config)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.mp3")

	_, err := gen.GenerateAudio("Test text", outputPath)
	if err == nil {
		t.Error("Expected error when API key is not configured")
	}
}

func TestGetAudioDuration(t *testing.T) {
	// Create a test file
	tmpFile, err := os.CreateTemp("", "test*.mp3")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write some test data
	testData := make([]byte, 16000) // 1 second at estimated bitrate
	if _, err := tmpFile.Write(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	duration, err := GetAudioDuration(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to get audio duration: %v", err)
	}

	// Check that duration is approximately 1 second
	if duration < 900*time.Millisecond || duration > 1100*time.Millisecond {
		t.Errorf("Expected duration around 1 second, got %v", duration)
	}
}

func TestGetAudioDurationFileNotFound(t *testing.T) {
	_, err := GetAudioDuration("/non/existent/file.mp3")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}
