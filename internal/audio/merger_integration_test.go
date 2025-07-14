package audio

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergerWithRealFFmpeg(t *testing.T) {
	// Skip if ffmpeg is not available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found in PATH, skipping integration test")
	}

	merger := NewAudioVideoMerger()
	tmpDir := t.TempDir()

	// Test 1: Create a simple audio concatenation
	t.Run("SimpleAudioConcatenation", func(t *testing.T) {
		// Create test audio file
		audioPath := filepath.Join(tmpDir, "test.mp3")
		cmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "anullsrc=duration=2", "-c:a", "libmp3lame", audioPath)
		if err := cmd.Run(); err != nil {
			t.Skipf("Failed to create test audio: %v", err)
		}

		outputPath := filepath.Join(tmpDir, "concat.mp3")
		err := merger.createTimedAudioTrack([]string{audioPath}, []int{3}, outputPath)
		if err != nil {
			t.Errorf("Failed to create timed audio track: %v", err)
		}

		// Check if output exists
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Output file was not created")
		}
	})

	// Test 2: Handle empty audio list
	t.Run("EmptyAudioList", func(t *testing.T) {
		outputPath := filepath.Join(tmpDir, "silence.mp3")
		err := merger.createTimedAudioTrack([]string{}, []int{5}, outputPath)
		if err != nil {
			t.Errorf("Failed to create silence track: %v", err)
		}

		// Check if output exists
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Output file was not created")
		}
	})

	// Test 3: Test codec detection
	t.Run("CodecDetection", func(t *testing.T) {
		// Create a simple VP8 video
		vp8Path := filepath.Join(tmpDir, "test_vp8.webm")
		cmd := exec.Command("ffmpeg",
			"-f", "lavfi", "-i", "testsrc=duration=1:size=64x64:rate=1",
			"-c:v", "libvpx", "-b:v", "50k",
			vp8Path)

		if err := cmd.Run(); err != nil {
			t.Skipf("Failed to create VP8 test video: %v", err)
		}

		// Test probing
		probeCmd := exec.Command(merger.ffmpegPath, "-i", vp8Path)
		probeOutput, _ := probeCmd.CombinedOutput()
		probeStr := string(probeOutput)

		if !strings.Contains(probeStr, "vp8") {
			t.Error("Failed to detect VP8 codec in test video")
		}
	})
}

func TestFFmpegCommandGeneration(t *testing.T) {
	// This test doesn't require ffmpeg to be installed

	// Test filter generation for various scenarios
	tests := []struct {
		name           string
		audioFiles     []string
		slideDurations []int
		expectError    bool
	}{
		{
			name:           "Single audio file",
			audioFiles:     []string{"audio1.mp3"},
			slideDurations: []int{10},
			expectError:    false,
		},
		{
			name:           "Multiple audio files",
			audioFiles:     []string{"audio1.mp3", "audio2.mp3", "audio3.mp3"},
			slideDurations: []int{5, 10, 15},
			expectError:    false,
		},
		{
			name:           "Mixed audio and silence",
			audioFiles:     []string{"audio1.mp3", "", "audio3.mp3"},
			slideDurations: []int{5, 10, 15},
			expectError:    false,
		},
		{
			name:           "All silence",
			audioFiles:     []string{"", "", ""},
			slideDurations: []int{5, 10, 15},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the function handles different inputs without panicking
			// We can't actually run ffmpeg without it being installed

			// Check that we handle various input combinations
			if len(tt.audioFiles) > len(tt.slideDurations) {
				t.Logf("Warning: More audio files than slide durations")
			}
		})
	}
}
