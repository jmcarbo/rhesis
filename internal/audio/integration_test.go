// +build integration

package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAudioVideoMergerIntegration(t *testing.T) {
	// Skip if ffmpeg is not available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found in PATH, skipping integration test")
	}

	merger := NewAudioVideoMerger()
	tmpDir := t.TempDir()

	// Test cases for different video formats
	tests := []struct {
		name          string
		videoFormat   string
		outputFormat  string
		expectSuccess bool
	}{
		{
			name:          "VP8 to MP4 with transcoding",
			videoFormat:   "vp8",
			outputFormat:  ".mp4",
			expectSuccess: true,
		},
		{
			name:          "VP8 to WebM without transcoding",
			videoFormat:   "vp8",
			outputFormat:  ".webm",
			expectSuccess: true,
		},
		{
			name:          "H264 to MP4 without transcoding",
			videoFormat:   "h264",
			outputFormat:  ".mp4",
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test video file
			videoPath := filepath.Join(tmpDir, "test_video"+tt.outputFormat)
			if err := createTestVideo(videoPath, tt.videoFormat); err != nil {
				t.Skipf("Failed to create test video: %v", err)
			}

			// Create test audio files
			audioFiles := []string{}
			durations := []int{5, 5, 5} // 3 slides, 5 seconds each
			
			for i := 0; i < 3; i++ {
				audioPath := filepath.Join(tmpDir, fmt.Sprintf("audio_%d.mp3", i))
				if err := createTestAudio(audioPath, 3); err != nil {
					t.Fatalf("Failed to create test audio %d: %v", i, err)
				}
				audioFiles = append(audioFiles, audioPath)
			}

			// Test merging
			outputPath := filepath.Join(tmpDir, "output"+tt.outputFormat)
			err := merger.MergeAudioWithVideo(videoPath, audioFiles, durations, outputPath)

			if tt.expectSuccess && err != nil {
				t.Errorf("Expected success but got error: %v", err)
			} else if !tt.expectSuccess && err == nil {
				t.Error("Expected error but got success")
			}

			// Verify output exists if successful
			if tt.expectSuccess && err == nil {
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Error("Output file was not created")
				}

				// Verify the output has both video and audio streams
				if hasAudio, hasVideo := verifyMediaStreams(outputPath); !hasAudio || !hasVideo {
					t.Errorf("Output file missing streams - audio: %v, video: %v", hasAudio, hasVideo)
				}
			}
		})
	}
}

// createTestVideo creates a simple test video using ffmpeg
func createTestVideo(outputPath string, codec string) error {
	args := []string{
		"-f", "lavfi",
		"-i", "testsrc=duration=15:size=320x240:rate=10",
		"-pix_fmt", "yuv420p",
	}

	switch codec {
	case "vp8":
		args = append(args, "-c:v", "libvpx", "-b:v", "500k")
	case "h264":
		args = append(args, "-c:v", "libx264", "-preset", "ultrafast")
	default:
		args = append(args, "-c:v", "libx264", "-preset", "ultrafast")
	}

	args = append(args, outputPath)
	
	cmd := exec.Command("ffmpeg", args...)
	return cmd.Run()
}

// createTestAudio creates a simple test audio file using ffmpeg
func createTestAudio(outputPath string, duration int) error {
	args := []string{
		"-f", "lavfi",
		"-i", fmt.Sprintf("sine=frequency=440:duration=%d", duration),
		"-c:a", "libmp3lame",
		"-b:a", "128k",
		outputPath,
	}
	
	cmd := exec.Command("ffmpeg", args...)
	return cmd.Run()
}

// verifyMediaStreams checks if a media file has audio and video streams
func verifyMediaStreams(filePath string) (hasAudio, hasVideo bool) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_type",
		"-of", "csv=p=0",
		filePath,
	)
	
	output, err := cmd.Output()
	if err == nil && string(output) == "video\n" {
		hasVideo = true
	}

	cmd = exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_type",
		"-of", "csv=p=0",
		filePath,
	)
	
	output, err = cmd.Output()
	if err == nil && string(output) == "audio\n" {
		hasAudio = true
	}

	return hasAudio, hasVideo
}

func TestCreateTimedAudioTrack(t *testing.T) {
	// Skip if ffmpeg is not available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found in PATH, skipping test")
	}

	merger := NewAudioVideoMerger()
	tmpDir := t.TempDir()

	// Create test audio files with different durations
	audioFiles := make([]string, 3)
	for i := 0; i < 3; i++ {
		if i == 1 {
			// Skip second audio file to test silence generation
			audioFiles[i] = ""
			continue
		}
		audioPath := filepath.Join(tmpDir, fmt.Sprintf("test_audio_%d.mp3", i))
		if err := createTestAudio(audioPath, 2); err != nil {
			t.Fatalf("Failed to create test audio: %v", err)
		}
		audioFiles[i] = audioPath
	}

	// Test with different slide durations
	slideDurations := []int{5, 3, 4} // First and third longer than audio, second has no audio
	outputPath := filepath.Join(tmpDir, "concatenated.mp3")

	err := merger.createTimedAudioTrack(audioFiles, slideDurations, outputPath)
	if err != nil {
		t.Fatalf("Failed to create timed audio track: %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Concatenated audio file was not created")
	}

	// Verify approximate duration (should be ~12 seconds)
	duration := getAudioDuration(t, outputPath)
	expectedDuration := 12.0 // 5 + 3 + 4 seconds
	tolerance := 1.0 // Allow 1 second tolerance

	if duration < expectedDuration-tolerance || duration > expectedDuration+tolerance {
		t.Errorf("Expected duration ~%.1f seconds, got %.1f seconds", expectedDuration, duration)
	}
}

// getAudioDuration gets the duration of an audio file in seconds using ffprobe
func getAudioDuration(t *testing.T, filePath string) float64 {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)
	
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get audio duration: %v", err)
		return 0
	}

	var duration float64
	fmt.Sscanf(string(output), "%f", &duration)
	return duration
}