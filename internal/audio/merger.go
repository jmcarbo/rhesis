package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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

			fmt.Printf("Slide %d: audio=%.2fs, slide=%.2fs\n", i+1, audioSeconds, slideSeconds)

			if audioSeconds < slideSeconds {
				// Need to add silence after the audio
				silenceDuration := slideSeconds - audioSeconds
				fmt.Printf("  Adding %.2fs of silence\n", silenceDuration)
				filterParts = append(filterParts,
					fmt.Sprintf("[%d:a]apad=pad_dur=%.3f[a%d]", inputIndex, silenceDuration, i))
			} else if audioSeconds > slideSeconds {
				// Audio is longer than slide, trim it to match slide duration
				fmt.Printf("  WARNING: Audio (%.2fs) is longer than slide duration (%.2fs)\n", audioSeconds, slideSeconds)
				fmt.Printf("  This should not happen if slide durations were adjusted properly!\n")
				fmt.Printf("  Trimming audio to %.2fs\n", slideSeconds)
				filterParts = append(filterParts,
					fmt.Sprintf("[%d:a]atrim=0:%.3f,asetpts=PTS-STARTPTS[a%d]", inputIndex, slideSeconds, i))
			} else {
				// Audio matches slide duration exactly
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
	// Get video duration
	videoDurationCmd := exec.Command(m.ffmpegPath, "-i", videoPath)
	videoDurationOutput, _ := videoDurationCmd.CombinedOutput()
	videoDurationStr := string(videoDurationOutput)

	// Extract video duration using regex
	var videoDuration float64
	durationRegex := regexp.MustCompile(`Duration: (\d{2}):(\d{2}):(\d{2}\.\d+)`)
	if matches := durationRegex.FindStringSubmatch(videoDurationStr); len(matches) > 3 {
		hours, _ := strconv.ParseFloat(matches[1], 64)
		minutes, _ := strconv.ParseFloat(matches[2], 64)
		seconds, _ := strconv.ParseFloat(matches[3], 64)
		videoDuration = hours*3600 + minutes*60 + seconds
	}

	// Get audio duration
	audioDurationCmd := exec.Command(m.ffmpegPath, "-i", audioPath)
	audioDurationOutput, _ := audioDurationCmd.CombinedOutput()
	audioDurationStr := string(audioDurationOutput)

	var audioDuration float64
	if matches := durationRegex.FindStringSubmatch(audioDurationStr); len(matches) > 3 {
		hours, _ := strconv.ParseFloat(matches[1], 64)
		minutes, _ := strconv.ParseFloat(matches[2], 64)
		seconds, _ := strconv.ParseFloat(matches[3], 64)
		audioDuration = hours*3600 + minutes*60 + seconds
	}

	// Calculate delay needed
	delay := videoDuration - audioDuration
	fmt.Printf("Video duration: %.2fs, Audio duration: %.2fs, Delay: %.2fs\n", videoDuration, audioDuration, delay)

	// First, detect the input video codec
	probeCmd := exec.Command(m.ffmpegPath, "-i", videoPath)
	probeOutput, _ := probeCmd.CombinedOutput()
	probeStr := string(probeOutput)

	// Determine output format based on file extension
	outputExt := strings.ToLower(filepath.Ext(outputPath))
	inputExt := strings.ToLower(filepath.Ext(videoPath))

	// Check the actual video codec in the input file
	isVP8Input := strings.Contains(probeStr, "Video: vp8") || strings.Contains(probeStr, "Video: vp9")
	isH264Input := strings.Contains(probeStr, "Video: h264")

	// Debug: log detected codec
	if isVP8Input {
		fmt.Printf("Detected VP8/VP9 video codec in input file\n")
	} else if isH264Input {
		fmt.Printf("Detected H.264 video codec in input file\n")
	}

	// Determine if we need to transcode based on codec compatibility
	needsTranscode := false
	if outputExt == ".mp4" && (isVP8Input || (!isH264Input && inputExt == ".webm")) {
		needsTranscode = true
		fmt.Printf("Transcoding required: VP8/WebM to MP4\n")
	} else if outputExt == ".webm" && isH264Input {
		needsTranscode = true
		fmt.Printf("Transcoding required: H.264 to WebM\n")
	}

	// Determine how to handle the sync
	args := []string{"-y"} // Overwrite output

	if delay > 0.5 { // If there's significant delay, trim the video start
		// Trim the beginning of the video to match when audio should start
		args = append(args, "-ss", fmt.Sprintf("%.3f", delay))
		args = append(args, "-i", videoPath)
		args = append(args, "-i", audioPath)
		fmt.Printf("Trimming %.2fs from video start to sync with audio\n", delay)
	} else {
		// No significant delay, use files as-is
		args = append(args, "-i", videoPath)
		args = append(args, "-i", audioPath)
	}

	// Handle video codec
	if needsTranscode {
		if outputExt == ".mp4" {
			// Transcode to H.264 for MP4
			args = append(args,
				"-c:v", "libx264",
				"-preset", "fast",
				"-crf", "23", // Good quality
			)
		} else if outputExt == ".webm" {
			// Transcode to VP8 for WebM
			args = append(args,
				"-c:v", "libvpx",
				"-b:v", "1M",
				"-crf", "10",
			)
		}
	} else {
		// Copy video codec if compatible
		args = append(args, "-c:v", "copy")
	}

	// Audio codec based on output format
	switch outputExt {
	case ".mp4":
		args = append(args, "-c:a", "aac", "-b:a", "192k")
		if !needsTranscode {
			args = append(args, "-movflags", "+faststart") // Optimize for streaming
		}
	case ".webm":
		args = append(args, "-c:a", "libopus", "-b:a", "128k")
	default:
		args = append(args, "-c:a", "aac", "-b:a", "192k") // Default to AAC
	}

	args = append(args,
		"-map", "0:v:0", // Map video from first input
		"-map", "1:a:0", // Map audio from second input
		"-shortest", // End output when shortest input ends
		outputPath,
	)

	cmd := exec.Command(m.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg merge failed: %w\nOutput: %s", err, string(output))
	}

	// The merged file is created at outputPath
	// The caller (main.go) will handle any file renaming if needed
	return nil
}
