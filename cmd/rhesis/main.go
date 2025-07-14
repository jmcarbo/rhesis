package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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
	)
	flag.Parse()

	if *scriptPath == "" {
		fmt.Println("Usage: rhesis -script <script-file> [-output <html-file>] [-style <style-name|css-file>] [-record <video-file>] [-play] [-transcription] [-subtitle <subtitle-file>]")
		os.Exit(1)
	}

	parsedScript, err := script.ParseScript(*scriptPath)
	if err != nil {
		log.Fatalf("Failed to parse script: %v", err)
	}

	gen := generator.NewHTMLGenerator()
	if err := gen.GeneratePresentation(parsedScript, *outputPath, *style, *transcription); err != nil {
		log.Fatalf("Failed to generate presentation: %v", err)
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
		p := player.NewPresentationPlayer()
		if err := p.PlayPresentation(*outputPath, *recordPath); err != nil {
			log.Fatalf("Failed to play presentation: %v", err)
		}
	}
}
