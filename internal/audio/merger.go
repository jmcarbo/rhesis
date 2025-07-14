package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// AudioVideoMerger handles merging audio tracks with video recordings
type AudioVideoMerger struct {
	ffmpegPath string
}

// NewAudioVideoMerger creates a new merger instance
func NewAudioVideoMerger() *AudioVideoMerger {
	return &AudioVideoMerger{
		ffmpegPath: "ffmpeg", // Default to PATH lookup
	}
}

// MergeAudioWithVideo merges audio files with a video recording based on slide timings
func (m *AudioVideoMerger) MergeAudioWithVideo(videoPath string, audioFiles []string, slideDurations []int, outputPath string) error {
	// Check if ffmpeg is available
	if err := m.checkFFmpeg(); err != nil {
		return fmt.Errorf("ffmpeg not available: %w", err)
	}

	// Create a temporary directory for intermediate files
	tempDir, err := os.MkdirTemp("", "rhesis_merge_*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create concatenated audio file with proper timing
	concatAudioPath := filepath.Join(tempDir, "concatenated_audio.mp3")
	if err := m.createTimedAudioTrack(audioFiles, slideDurations, concatAudioPath); err != nil {
		return fmt.Errorf("failed to create timed audio track: %w", err)
	}

	// Merge audio with video
	if err := m.mergeFiles(videoPath, concatAudioPath, outputPath); err != nil {
		return fmt.Errorf("failed to merge audio and video: %w", err)
	}

	return nil
}

// checkFFmpeg verifies that ffmpeg is available
func (m *AudioVideoMerger) checkFFmpeg() error {
	cmd := exec.Command(m.ffmpegPath, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg not found in PATH. Please install ffmpeg to use audio merging")
	}
	return nil
}

// createTimedAudioTrack creates a single audio file with proper timing for each slide
func (m *AudioVideoMerger) createTimedAudioTrack(audioFiles []string, slideDurations []int, outputPath string) error {
	// Create a complex filter to concatenate audio with silence padding
	var filterParts []string
	var inputs []string
	inputIndex := 0
	
	for i, duration := range slideDurations {
		if i < len(audioFiles) && audioFiles[i] != "" {
			// Add the audio file as input
			inputs = append(inputs, "-i", audioFiles[i])
			
			// Add silence padding if needed
			audioDuration, err := GetAudioDuration(audioFiles[i])
			if err != nil {
				// If we can't get duration, use the slide duration
				audioDuration = time.Duration(duration) * time.Second
			}
			
			audioSeconds := audioDuration.Seconds()
			slideSeconds := float64(duration)
			
			if audioSeconds < slideSeconds {
				// Need to add silence after the audio
				silenceDuration := slideSeconds - audioSeconds
				filterParts = append(filterParts, 
					fmt.Sprintf("[%d:a]apad=pad_dur=%.3f[a%d]", inputIndex, silenceDuration, i))
			} else {
				// Audio is longer or equal, just reference it
				filterParts = append(filterParts, fmt.Sprintf("[%d:a]acopy[a%d]", inputIndex, i))
			}
			inputIndex++
		} else {
			// No audio for this slide, create silence
			filterParts = append(filterParts,
				fmt.Sprintf("anullsrc=duration=%d:sample_rate=44100:channel_layout=stereo[a%d]", duration, i))
		}
	}
	
	// Concatenate all audio segments
	var concatInputs []string
	for i := range slideDurations {
		concatInputs = append(concatInputs, fmt.Sprintf("[a%d]", i))
	}
	
	filterComplex := strings.Join(filterParts, ";") + ";" +
		strings.Join(concatInputs, "") + fmt.Sprintf("concat=n=%d:v=0:a=1[out]", len(slideDurations))
	
	// Build ffmpeg command
	args := []string{"-y"} // Overwrite output
	args = append(args, inputs...)
	args = append(args, "-filter_complex", filterComplex)
	args = append(args, "-map", "[out]")
	args = append(args, "-codec:a", "libmp3lame")
	args = append(args, "-b:a", "192k")
	args = append(args, outputPath)
	
	cmd := exec.Command(m.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg audio concatenation failed: %w\nOutput: %s", err, string(output))
	}
	
	return nil
}

// mergeFiles merges the audio track with the video file
func (m *AudioVideoMerger) mergeFiles(videoPath, audioPath, outputPath string) error {
	// Determine output format based on file extension
	ext := strings.ToLower(filepath.Ext(outputPath))
	
	args := []string{
		"-y", // Overwrite output
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy", // Copy video codec
		"-c:a", "aac",  // Use AAC for audio (compatible with most formats)
		"-b:a", "192k",
		"-map", "0:v:0", // Map video from first input
		"-map", "1:a:0", // Map audio from second input
		"-shortest",     // End output when shortest input ends
	}
	
	// Add format-specific options
	switch ext {
	case ".mp4":
		args = append(args, "-movflags", "+faststart") // Optimize for streaming
	case ".webm":
		args = append(args, "-c:a", "libopus") // Use Opus for WebM
	}
	
	args = append(args, outputPath)
	
	cmd := exec.Command(m.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg merge failed: %w\nOutput: %s", err, string(output))
	}
	
	// Remove the original video file and rename the merged one
	tempOutput := outputPath + ".tmp"
	if err := os.Rename(outputPath, tempOutput); err != nil {
		// If rename fails, we've already created the merged file, so just log a warning
		fmt.Printf("Warning: Could not create temporary file: %v\n", err)
		return nil
	}
	
	// Remove original video
	if err := os.Remove(videoPath); err != nil {
		fmt.Printf("Warning: Could not remove original video: %v\n", err)
	}
	
	// Rename merged file to original video path
	if err := os.Rename(tempOutput, videoPath); err != nil {
		// Try to recover
		os.Rename(tempOutput, outputPath)
		return fmt.Errorf("failed to finalize merged video: %w", err)
	}
	
	return nil
}