package subtitle

import (
	"fmt"
	"strings"
	"time"
)

// SubtitleFormat represents the format of the subtitle file
type SubtitleFormat string

const (
	FormatSRT    SubtitleFormat = "srt"
	FormatWebVTT SubtitleFormat = "vtt"
)

// Subtitle represents a single subtitle entry
type Subtitle struct {
	Index     int
	StartTime time.Duration
	EndTime   time.Duration
	Text      string
}

// Generator handles subtitle generation from slide transcriptions
type Generator struct {
	format SubtitleFormat
}

// NewGenerator creates a new subtitle generator
func NewGenerator(format SubtitleFormat) *Generator {
	return &Generator{
		format: format,
	}
}

// Generate creates subtitle content from slide transcriptions
func (g *Generator) Generate(transcriptions []string, durations []int, defaultDuration int) string {
	var subtitles []Subtitle
	currentTime := time.Duration(0)

	for i, transcription := range transcriptions {
		if transcription == "" {
			continue
		}

		// Get duration for this slide
		duration := defaultDuration
		if i < len(durations) && durations[i] > 0 {
			duration = durations[i]
		}
		slideDuration := time.Duration(duration) * time.Second

		// Split transcription into subtitle chunks (max 2 lines, ~50 chars per line)
		chunks := g.splitIntoChunks(transcription, 100)

		// Calculate time per chunk
		chunkDuration := slideDuration / time.Duration(len(chunks))

		for _, chunk := range chunks {
			subtitle := Subtitle{
				Index:     len(subtitles) + 1,
				StartTime: currentTime,
				EndTime:   currentTime + chunkDuration,
				Text:      chunk,
			}
			subtitles = append(subtitles, subtitle)
			currentTime = subtitle.EndTime
		}
	}

	// Format subtitles based on the selected format
	switch g.format {
	case FormatSRT:
		return g.formatSRT(subtitles)
	case FormatWebVTT:
		return g.formatWebVTT(subtitles)
	default:
		return g.formatSRT(subtitles)
	}
}

// splitIntoChunks splits text into subtitle-friendly chunks
func (g *Generator) splitIntoChunks(text string, maxCharsPerChunk int) []string {
	// Remove markdown formatting
	text = g.cleanMarkdown(text)

	words := strings.Fields(text)
	var chunks []string
	var currentChunk []string
	currentLength := 0

	for _, word := range words {
		wordLength := len(word)

		// If adding this word would exceed the limit, start a new chunk
		if currentLength > 0 && currentLength+wordLength+1 > maxCharsPerChunk {
			chunks = append(chunks, strings.Join(currentChunk, " "))
			currentChunk = []string{}
			currentLength = 0
		}

		currentChunk = append(currentChunk, word)
		currentLength += wordLength
		if currentLength > 0 {
			currentLength++ // Space between words
		}
	}

	// Add the last chunk
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, " "))
	}

	return chunks
}

// cleanMarkdown removes common markdown formatting
func (g *Generator) cleanMarkdown(text string) string {
	// Remove code blocks
	text = strings.ReplaceAll(text, "```", "")

	// Remove bold/italic markers
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "__", "")
	text = strings.ReplaceAll(text, "*", "")
	text = strings.ReplaceAll(text, "_", "")

	// Remove headers
	lines := strings.Split(text, "\n")
	var cleanedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") && line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	return strings.Join(cleanedLines, " ")
}

// formatSRT formats subtitles in SRT format
func (g *Generator) formatSRT(subtitles []Subtitle) string {
	var result strings.Builder

	for _, sub := range subtitles {
		result.WriteString(fmt.Sprintf("%d\n", sub.Index))
		result.WriteString(fmt.Sprintf("%s --> %s\n",
			g.formatTimeSRT(sub.StartTime),
			g.formatTimeSRT(sub.EndTime)))
		result.WriteString(fmt.Sprintf("%s\n\n", sub.Text))
	}

	return result.String()
}

// formatWebVTT formats subtitles in WebVTT format
func (g *Generator) formatWebVTT(subtitles []Subtitle) string {
	var result strings.Builder

	result.WriteString("WEBVTT\n\n")

	for _, sub := range subtitles {
		result.WriteString(fmt.Sprintf("%s --> %s\n",
			g.formatTimeWebVTT(sub.StartTime),
			g.formatTimeWebVTT(sub.EndTime)))
		result.WriteString(fmt.Sprintf("%s\n\n", sub.Text))
	}

	return result.String()
}

// formatTimeSRT formats time for SRT format (HH:MM:SS,mmm)
func (g *Generator) formatTimeSRT(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := int(d.Milliseconds()) % 1000

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, seconds, milliseconds)
}

// formatTimeWebVTT formats time for WebVTT format (HH:MM:SS.mmm)
func (g *Generator) formatTimeWebVTT(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := int(d.Milliseconds()) % 1000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}

// DetectFormat detects subtitle format from file extension
func DetectFormat(filename string) SubtitleFormat {
	if strings.HasSuffix(strings.ToLower(filename), ".vtt") {
		return FormatWebVTT
	}
	return FormatSRT
}
