package audio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type Generator interface {
	GenerateAudio(text string, outputPath string) (duration time.Duration, err error)
}

type ElevenLabsConfig struct {
	APIKey  string
	VoiceID string
	ModelID string
}

type ElevenLabsGenerator struct {
	config     ElevenLabsConfig
	httpClient *http.Client
}

func NewElevenLabsGenerator(config ElevenLabsConfig) *ElevenLabsGenerator {
	if config.VoiceID == "" {
		config.VoiceID = "21m00Tcm4TlvDq8ikWAM" // Default to Rachel voice
	}
	if config.ModelID == "" {
		config.ModelID = "eleven_multilingual_v2"
	}

	return &ElevenLabsGenerator{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type ttsRequest struct {
	Text          string        `json:"text"`
	ModelID       string        `json:"model_id"`
	VoiceSettings voiceSettings `json:"voice_settings"`
}

type voiceSettings struct {
	Stability       float64 `json:"stability"`
	SimilarityBoost float64 `json:"similarity_boost"`
}

func (g *ElevenLabsGenerator) GenerateAudio(text string, outputPath string) (time.Duration, error) {
	if g.config.APIKey == "" {
		return 0, fmt.Errorf("ElevenLabs API key not configured")
	}

	// Create request payload
	payload := ttsRequest{
		Text:    text,
		ModelID: g.config.ModelID,
		VoiceSettings: voiceSettings{
			Stability:       0.5,
			SimilarityBoost: 0.5,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", g.config.VoiceID)
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("xi-api-key", g.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")

	// Send request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("ElevenLabs API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save audio file
	file, err := os.Create(outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	if _, err := file.Write(audioData); err != nil {
		return 0, fmt.Errorf("failed to write audio file: %w", err)
	}

	// Estimate duration based on text length and average speech rate
	// Average speech rate is about 150 words per minute
	// This is a rough estimate; for accurate duration, we'd need to decode the MP3
	wordCount := float64(len(text)) / 5.0 // Rough estimate: 5 characters per word
	minutes := wordCount / 150.0
	duration := time.Duration(minutes * float64(time.Minute))

	// Ensure minimum duration of 1 second
	if duration < time.Second {
		duration = time.Second
	}

	return duration, nil
}

// GetAudioDuration gets the actual duration of an audio file using ffmpeg
func GetAudioDuration(filePath string) (time.Duration, error) {
	cmd := exec.Command("ffmpeg", "-i", filePath)
	output, _ := cmd.CombinedOutput()
	outputStr := string(output)

	// Extract duration using regex
	durationRegex := regexp.MustCompile(`Duration: (\d{2}):(\d{2}):(\d{2}\.\d+)`)
	if matches := durationRegex.FindStringSubmatch(outputStr); len(matches) > 3 {
		hours, _ := strconv.ParseFloat(matches[1], 64)
		minutes, _ := strconv.ParseFloat(matches[2], 64)
		seconds, _ := strconv.ParseFloat(matches[3], 64)
		totalSeconds := hours*3600 + minutes*60 + seconds
		return time.Duration(totalSeconds * float64(time.Second)), nil
	}

	// Fall back to file size estimate if ffmpeg fails
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("could not extract duration from audio file and stat failed: %w", err)
	}

	// Rough estimate: file size in bytes / (128000 bits/sec / 8 bits/byte)
	seconds := float64(info.Size()) / 16000.0
	return time.Duration(seconds * float64(time.Second)), nil
}
