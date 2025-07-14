package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jmcarbo/rhesis/internal/generator"
	"github.com/jmcarbo/rhesis/internal/player"
	"github.com/jmcarbo/rhesis/internal/script"
)

func main() {
	var (
		scriptPath    = flag.String("script", "", "Path to the presentation script file")
		outputPath    = flag.String("output", "presentation.html", "Output HTML file path")
		recordPath    = flag.String("record", "", "Path to save video recording (optional)")
		play          = flag.Bool("play", false, "Play the presentation after generating")
		style         = flag.String("style", "modern", "Presentation style (modern, minimal, dark, elegant, or path to custom CSS file)")
		transcription = flag.Bool("transcription", false, "Include transcription panel in presentation")
	)
	flag.Parse()

	if *scriptPath == "" {
		fmt.Println("Usage: rhesis -script <script-file> [-output <html-file>] [-style <style-name|css-file>] [-record <video-file>] [-play] [-transcription]")
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

	if *play {
		p := player.NewPresentationPlayer()
		if err := p.PlayPresentation(*outputPath, *recordPath); err != nil {
			log.Fatalf("Failed to play presentation: %v", err)
		}
	}
}
