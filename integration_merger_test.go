// +build integration

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jmcarbo/rhesis/internal/audio"
)

// TestAudioVideoMergingE2E tests the complete audio/video merging flow
func TestAudioVideoMergingE2E(t *testing.T) {
	// Skip if ffmpeg is not available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found in PATH, skipping integration test")
	}

	tmpDir := t.TempDir()
	merger := audio.NewAudioVideoMerger()

	// Create a simple test video with VP8 codec (WebM)
	videoPath := filepath.Join(tmpDir, "test.webm")
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", "color=c=blue:s=320x240:d=5",
		"-c:v", "libvpx",
		"-b:v", "200k",
		videoPath,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create test video: %v", err)
	}

	// Create simple audio files
	audioFiles := []string{}
	for i := 0; i < 2; i++ {
		audioPath := filepath.Join(tmpDir, fmt.Sprintf("audio_%d.mp3", i))
		cmd := exec.Command("ffmpeg",
			"-f", "lavfi",
			"-i", "sine=frequency=440:duration=2",
			"-c:a", "libmp3lame",
			audioPath,
		)
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create audio %d: %v", i, err)
		}
		audioFiles = append(audioFiles, audioPath)
	}

	// Test merging to MP4 (should transcode VP8 to H264)
	t.Run("VP8_to_MP4", func(t *testing.T) {
		outputPath := filepath.Join(tmpDir, "output.mp4")
		durations := []int{3, 3} // Two 3-second slides
		
		err := merger.MergeAudioWithVideo(videoPath, audioFiles, durations, outputPath)
		if err != nil {
			t.Fatalf("Failed to merge: %v", err)
		}

		// Verify output exists
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Output MP4 was not created")
		}

		// Verify it's a valid MP4 with both streams
		cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "stream=codec_type", "-of", "csv=p=0", outputPath)
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to probe output: %v", err)
		}

		outputStr := string(output)
		if !contains(outputStr, "video") || !contains(outputStr, "audio") {
			t.Errorf("Output missing expected streams: %s", outputStr)
		}
	})
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}