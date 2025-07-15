package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jmcarbo/rhesis/internal/audio"
	"github.com/jmcarbo/rhesis/internal/generator"
	"github.com/jmcarbo/rhesis/internal/player"
	"github.com/jmcarbo/rhesis/internal/script"
	"github.com/jmcarbo/rhesis/internal/subtitle"
)

func main() {
	var (
		scriptPath    = flag.String("script", "", "Path to the presentation script file")
		outputPath    = flag.String("output", "presentation.html", "Output HTML file path")
		recordPath    = flag.String("record", "", "Path to save video recording (optional)")
		play          = flag.Bool("play", false, "Play the presentation after generating")
		style         = flag.String("style", "modern", "Presentation style (modern, minimal, dark, elegant, or path to custom CSS file)")
		transcription = flag.Bool("transcription", false, "Include transcription panel in presentation")
		subtitlePath  = flag.String("subtitle", "", "Generate subtitle file (optional, .srt or .vtt)")
		sound         = flag.Bool("sound", false, "Generate audio from transcriptions using ElevenLabs")
		apiKey        = flag.String("elevenlabs-key", os.Getenv("ELEVENLABS_API_KEY"), "ElevenLabs API key (or set ELEVENLABS_API_KEY env var)")
		voiceID       = flag.String("voice", "", "ElevenLabs voice ID (optional, defaults to Rachel)")
		skipAudioGen  = flag.Bool("skip-audio-creation", false, "Skip audio generation if audio files already exist")
		background    = flag.Bool("background", false, "Run presentation in background (headless mode)")
		fuse          = flag.Bool("fuse", false, "Fuse mode: merge video and audio files (requires -video, -audio, and -output)")
		videoPath     = flag.String("video", "", "Input video file path (for fuse mode)")
		audioPath     = flag.String("audio", "", "Input audio file path or directory (for fuse mode)")
		durations     = flag.String("durations", "", "Comma-separated slide durations in seconds (for fuse mode with audio directory)")
	)
	flag.Parse()

	// Handle fuse mode separately
	if *fuse {
		if *videoPath == "" || *audioPath == "" || *outputPath == "" {
			fmt.Println("Usage for fuse mode: rhesis -fuse -video <video-file> -audio <audio-file-or-directory> -output <output-file> [-durations <comma-separated-durations>]")
			fmt.Println("  -video: Input video file path")
			fmt.Println("  -audio: Input audio file path (single file) or directory (multiple audio files)")
			fmt.Println("  -output: Output video file path")
			fmt.Println("  -durations: Comma-separated slide durations in seconds (required when -audio is a directory)")
			os.Exit(1)
		}

		// Run fuse mode
		if err := runFuseMode(*videoPath, *audioPath, *outputPath, *durations); err != nil {
			log.Fatalf("Fuse operation failed: %v", err)
		}
		return
	}

	// Normal presentation mode
	if *scriptPath == "" {
		fmt.Println("Usage: rhesis -script <script-file> [-output <html-file>] [-style <style-name|css-file>] [-record <video-file>] [-play] [-background] [-transcription] [-subtitle <subtitle-file>] [-sound] [-skip-audio-creation] [-elevenlabs-key <api-key>] [-voice <voice-id>]")
		fmt.Println("\nOr for fuse mode:")
		fmt.Println("  rhesis -fuse -video <video-file> -audio <audio-file-or-directory> -output <output-file> [-durations <comma-separated-durations>]")
		os.Exit(1)
	}

	parsedScript, err := script.ParseScript(*scriptPath)
	if err != nil {
		log.Fatalf("Failed to parse script: %v", err)
	}

	// Generate audio if requested
	audioFiles := make([]string, 0)
	if *sound {
		// Only require API key if we're not skipping audio generation entirely
		if *apiKey == "" && !*skipAudioGen {
			log.Fatal("ElevenLabs API key is required when using -sound flag. Use -elevenlabs-key or set ELEVENLABS_API_KEY environment variable.")
		}

		audioGen := audio.NewElevenLabsGenerator(audio.ElevenLabsConfig{
			APIKey:  *apiKey,
			VoiceID: *voiceID,
		})

		// Create audio output directory
		audioDir := strings.TrimSuffix(*outputPath, filepath.Ext(*outputPath)) + "_audio"
		if err := os.MkdirAll(audioDir, 0755); err != nil {
			log.Fatalf("Failed to create audio directory: %v", err)
		}

		fmt.Println("Processing audio files...")
		for i, slide := range parsedScript.Slides {
			if slide.Transcription != "" {
				audioPath := filepath.Join(audioDir, fmt.Sprintf("slide_%02d.mp3", i+1))

				// Check if we should skip audio generation
				if *skipAudioGen {
					if _, err := os.Stat(audioPath); err == nil {
						// Audio file exists, skip generation
						fmt.Printf("Using existing audio file for slide %d: %s\n", i+1, audioPath)

						// Get audio duration to adjust slide timing if needed
						audioDuration, err := audio.GetAudioDuration(audioPath)
						if err == nil {
							audioDurationSeconds := int(audioDuration.Seconds())
							originalDuration := slide.Duration
							if audioDurationSeconds > slide.Duration {
								parsedScript.Slides[i].Duration = audioDurationSeconds + 1 // Add 1 second buffer
								fmt.Printf("Adjusted slide %d duration from %ds to %ds to accommodate audio\n",
									i+1, originalDuration, parsedScript.Slides[i].Duration)
							}
						} else {
							fmt.Printf("Warning: Could not get duration for audio file %s: %v\n", audioPath, err)
						}

						audioFiles = append(audioFiles, audioPath)
						continue
					} else {
						fmt.Printf("Audio file not found for slide %d, generating...\n", i+1)
					}
				}

				// Generate audio
				audioDuration, err := audioGen.GenerateAudio(slide.Transcription, audioPath)
				if err != nil {
					log.Printf("Warning: Failed to generate audio for slide %d: %v", i+1, err)
					continue
				}

				// Get actual audio duration if possible
				actualDuration, err := audio.GetAudioDuration(audioPath)
				if err == nil {
					audioDuration = actualDuration
				}

				// Adjust slide duration if audio is longer
				audioDurationSeconds := int(audioDuration.Seconds())
				originalDuration := slide.Duration
				if audioDurationSeconds > slide.Duration {
					parsedScript.Slides[i].Duration = audioDurationSeconds + 1 // Add 1 second buffer
					fmt.Printf("Adjusted slide %d duration from %ds to %ds to accommodate audio\n",
						i+1, originalDuration, parsedScript.Slides[i].Duration)
				}

				audioFiles = append(audioFiles, audioPath)
				fmt.Printf("Generated audio for slide %d (duration: %v)\n", i+1, audioDuration)
			}
		}
	}

	gen := generator.NewHTMLGenerator()
	if *sound && len(audioFiles) > 0 {
		if err := gen.GeneratePresentationWithOptions(parsedScript, *outputPath, *style, *transcription, audioFiles, *background); err != nil {
			log.Fatalf("Failed to generate presentation: %v", err)
		}
	} else {
		if err := gen.GeneratePresentation(parsedScript, *outputPath, *style, *transcription); err != nil {
			log.Fatalf("Failed to generate presentation: %v", err)
		}
	}

	fmt.Printf("Presentation generated: %s\n", *outputPath)

	// Generate subtitle file if requested
	if *subtitlePath != "" {
		// Extract transcriptions and durations from slides
		var transcriptions []string
		var durations []int
		for _, slide := range parsedScript.Slides {
			transcriptions = append(transcriptions, slide.Transcription)
			durations = append(durations, slide.Duration)
		}

		// Generate subtitles
		format := subtitle.DetectFormat(*subtitlePath)
		gen := subtitle.NewGenerator(format)
		subtitleContent := gen.Generate(transcriptions, durations, parsedScript.DefaultTime)

		// Write subtitle file
		if err := os.WriteFile(*subtitlePath, []byte(subtitleContent), 0644); err != nil {
			log.Fatalf("Failed to write subtitle file: %v", err)
		}
		fmt.Printf("Subtitle file generated: %s\n", *subtitlePath)
	}

	if *play {
		if *background {
			fmt.Println("Running presentation in background mode (headless)...")
		}
		p := player.NewPresentationPlayer()
		if err := p.PlayPresentationWithOptions(*outputPath, *recordPath, *background); err != nil {
			log.Fatalf("Failed to play presentation: %v", err)
		}

		// If both recording and sound were enabled, merge audio with video
		if *recordPath != "" && *sound && len(audioFiles) > 0 {
			fmt.Println("Merging audio with video recording...")
			merger := audio.NewAudioVideoMerger()

			// Extract slide durations
			durations := make([]int, len(parsedScript.Slides))
			totalExpectedDuration := 0
			for i, slide := range parsedScript.Slides {
				durations[i] = slide.Duration
				totalExpectedDuration += slide.Duration
			}
			fmt.Printf("Expected total duration: %d seconds\n", totalExpectedDuration)

			// Create output path for merged video
			mergedPath := strings.TrimSuffix(*recordPath, filepath.Ext(*recordPath)) + "_with_audio" + filepath.Ext(*recordPath)

			if err := merger.MergeAudioWithVideo(*recordPath, audioFiles, durations, mergedPath); err != nil {
				log.Printf("Warning: Failed to merge audio with video: %v", err)
				log.Printf("Original video saved without audio to: %s", *recordPath)
			} else {
				fmt.Printf("Original video (no audio): %s\n", *recordPath)
				fmt.Printf("Video with audio: %s\n", mergedPath)
			}
		}
	}
}

// runFuseMode handles the fuse operation to merge video and audio files
func runFuseMode(videoPath, audioPath, outputPath, durationsStr string) error {
	// Check if audioPath is a file or directory
	fileInfo, err := os.Stat(audioPath)
	if err != nil {
		return fmt.Errorf("failed to stat audio path: %w", err)
	}

	merger := audio.NewAudioVideoMerger()

	if fileInfo.IsDir() {
		// Multiple audio files in a directory
		if durationsStr == "" {
			return fmt.Errorf("durations parameter is required when audio path is a directory")
		}

		// Parse durations
		durationStrs := strings.Split(durationsStr, ",")
		durations := make([]int, len(durationStrs))
		for i, dStr := range durationStrs {
			d, err := strconv.Atoi(strings.TrimSpace(dStr))
			if err != nil {
				return fmt.Errorf("invalid duration at position %d: %w", i+1, err)
			}
			durations[i] = d
		}

		// Find audio files in the directory
		audioFiles := make([]string, 0)
		entries, err := os.ReadDir(audioPath)
		if err != nil {
			return fmt.Errorf("failed to read audio directory: %w", err)
		}

		// Sort entries by name to ensure consistent ordering
		for _, entry := range entries {
			if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".mp3") || strings.HasSuffix(entry.Name(), ".wav") || strings.HasSuffix(entry.Name(), ".m4a")) {
				audioFiles = append(audioFiles, filepath.Join(audioPath, entry.Name()))
			}
		}

		if len(audioFiles) == 0 {
			return fmt.Errorf("no audio files found in directory: %s", audioPath)
		}

		fmt.Printf("Found %d audio files in directory\n", len(audioFiles))

		// Ensure we have enough durations
		if len(durations) < len(audioFiles) {
			return fmt.Errorf("not enough durations provided: have %d durations for %d audio files", len(durations), len(audioFiles))
		}

		// Merge with multiple audio files
		return merger.MergeAudioWithVideo(videoPath, audioFiles, durations, outputPath)

	} else {
		// Single audio file
		fmt.Println("Merging single audio file with video...")

		// Get video duration to use as the single slide duration
		videoDurationTime, err := audio.GetVideoDuration(videoPath)
		var videoDuration int
		if err != nil {
			// If we can't get video duration, use a large default
			fmt.Printf("Warning: Could not get video duration, using default: %v\n", err)
			videoDuration = 300 // 5 minutes default
		} else {
			videoDuration = int(videoDurationTime.Seconds())
		}

		// For a single audio file, treat it as one slide with the video's duration
		return merger.MergeAudioWithVideo(videoPath, []string{audioPath}, []int{videoDuration}, outputPath)
	}
}
