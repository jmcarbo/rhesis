package main

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Script struct {
	Title       string  `yaml:"title"`
	Duration    int     `yaml:"duration"`
	Slides      []Slide `yaml:"slides"`
	DefaultTime int     `yaml:"default_time"`
}

type Slide struct {
	Title        string `yaml:"title"`
	Content      string `yaml:"content"`
	Image        string `yaml:"image"`
	Transcription string `yaml:"transcription"`
	Duration     int    `yaml:"duration"`
}

func ParseScript(path string) (*Script, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var script Script
	if err := yaml.Unmarshal(data, &script); err != nil {
		return nil, err
	}

	if script.DefaultTime == 0 {
		script.DefaultTime = 10
	}

	for i := range script.Slides {
		if script.Slides[i].Duration == 0 {
			script.Slides[i].Duration = script.DefaultTime
		}
		if script.Slides[i].Image != "" {
			script.Slides[i].Image = filepath.Clean(script.Slides[i].Image)
		}
	}

	return &script, nil
}

func (s *Script) GetTotalDuration() time.Duration {
	total := 0
	for _, slide := range s.Slides {
		total += slide.Duration
	}
	return time.Duration(total) * time.Second
}