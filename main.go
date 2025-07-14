package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var (
		scriptPath = flag.String("script", "", "Path to the presentation script file")
		outputPath = flag.String("output", "presentation.html", "Output HTML file path")
		recordPath = flag.String("record", "", "Path to save video recording (optional)")
		play       = flag.Bool("play", false, "Play the presentation after generating")
	)
	flag.Parse()

	if *scriptPath == "" {
		fmt.Println("Usage: rhesis -script <script-file> [-output <html-file>] [-record <video-file>] [-play]")
		os.Exit(1)
	}

	script, err := ParseScript(*scriptPath)
	if err != nil {
		log.Fatalf("Failed to parse script: %v", err)
	}

	generator := NewHTMLGenerator()
	if err := generator.GeneratePresentation(script, *outputPath); err != nil {
		log.Fatalf("Failed to generate presentation: %v", err)
	}

	fmt.Printf("Presentation generated: %s\n", *outputPath)

	if *play {
		player := NewPresentationPlayer()
		if err := player.PlayPresentation(*outputPath, *recordPath); err != nil {
			log.Fatalf("Failed to play presentation: %v", err)
		}
	}
}