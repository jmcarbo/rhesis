package script

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Script struct {
	Title       string
	Duration    int
	Slides      []Slide
	DefaultTime int
}

type Slide struct {
	Title         string
	Content       string
	Image         string
	Transcription string
	Duration      int
}

func ParseScript(path string) (*Script, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	script := &Script{
		DefaultTime: 10, // Default to 10 seconds if not specified
	}

	scanner := bufio.NewScanner(file)
	var currentSlide *Slide
	var inContent bool
	var inTranscription bool
	var contentBuilder, transcriptionBuilder strings.Builder
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check for title (first H1)
		if strings.HasPrefix(line, "# ") && script.Title == "" {
			script.Title = strings.TrimPrefix(line, "# ")
			continue
		}

		// Check for metadata
		if strings.HasPrefix(trimmedLine, "Duration:") && currentSlide == nil {
			durationStr := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "Duration:"))
			if duration, err := strconv.Atoi(durationStr); err == nil {
				script.Duration = duration
			}
			continue
		}

		if strings.HasPrefix(trimmedLine, "Default time:") {
			defaultTimeStr := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "Default time:"))
			if defaultTime, err := strconv.Atoi(defaultTimeStr); err == nil {
				script.DefaultTime = defaultTime
			}
			continue
		}

		// Check for slide title (H2)
		if strings.HasPrefix(line, "## ") {
			// Save previous slide if exists
			if currentSlide != nil {
				if inContent {
					currentSlide.Content = strings.TrimSpace(contentBuilder.String())
					contentBuilder.Reset()
					inContent = false
				}
				if inTranscription {
					currentSlide.Transcription = strings.TrimSpace(transcriptionBuilder.String())
					transcriptionBuilder.Reset()
					inTranscription = false
				}
				script.Slides = append(script.Slides, *currentSlide)
			}

			// Start new slide
			currentSlide = &Slide{
				Title:    strings.TrimPrefix(line, "## "),
				Duration: script.DefaultTime,
			}
			inContent = true
			continue
		}

		// Check for horizontal rule (transcription separator)
		if trimmedLine == "---" && currentSlide != nil {
			if inContent {
				currentSlide.Content = strings.TrimSpace(contentBuilder.String())
				contentBuilder.Reset()
				inContent = false
			}
			inTranscription = true
			continue
		}

		// Check for slide duration
		if currentSlide != nil && strings.HasPrefix(trimmedLine, "Duration:") {
			durationStr := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "Duration:"))
			if duration, err := strconv.Atoi(durationStr); err == nil {
				currentSlide.Duration = duration
			}
			continue
		}

		// Check for image
		if currentSlide != nil && strings.HasPrefix(trimmedLine, "Image:") {
			imagePath := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "Image:"))
			currentSlide.Image = filepath.Clean(imagePath)
			continue
		}

		// Accumulate content or transcription
		if currentSlide != nil {
			if inTranscription {
				if transcriptionBuilder.Len() > 0 {
					transcriptionBuilder.WriteString("\n")
				}
				transcriptionBuilder.WriteString(line)
			} else if inContent {
				if contentBuilder.Len() > 0 {
					contentBuilder.WriteString("\n")
				}
				contentBuilder.WriteString(line)
			}
		}
	}

	// Save last slide
	if currentSlide != nil {
		if inContent {
			currentSlide.Content = strings.TrimSpace(contentBuilder.String())
		}
		if inTranscription {
			currentSlide.Transcription = strings.TrimSpace(transcriptionBuilder.String())
		}
		script.Slides = append(script.Slides, *currentSlide)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Validate
	if script.Title == "" {
		return nil, fmt.Errorf("presentation must have a title (# Title)")
	}

	return script, nil
}

func (s *Script) GetTotalDuration() time.Duration {
	total := 0
	for _, slide := range s.Slides {
		total += slide.Duration
	}
	return time.Duration(total) * time.Second
}
