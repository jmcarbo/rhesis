# Rhesis

A Go program that reads a script and generates, plays, and records presentations in HTML with Playwright.

## Features

- **Script-based presentations**: Define slides with YAML scripts
- **HTML generation**: Creates beautiful HTML presentations with CSS styling
- **Timing control**: Configurable slide durations for smooth transitions
- **Transcription support**: Display explanatory text alongside each slide
- **Image support**: Embed images directly in slides (PNG, JPG, GIF, WebP)
- **Automatic playback**: Play presentations automatically with proper timing
- **Recording capability**: Record presentations to video files using Playwright
- **Keyboard controls**: Navigate slides with arrow keys and spacebar
- **Responsive design**: Works on different screen sizes

## Installation

```bash
go get github.com/jmcarbo/rhesis
```

Or clone and build:

```bash
git clone https://github.com/jmcarbo/rhesis.git
cd rhesis
make build
```

## Prerequisites

For video recording functionality, you need to install Playwright browsers:

```bash
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install
```

## Usage

### Basic Usage

```bash
# Generate HTML presentation
./bin/rhesis -script presentation.yaml -output presentation.html

# Generate and play presentation
./bin/rhesis -script presentation.yaml -output presentation.html -play

# Generate, play, and record presentation
./bin/rhesis -script presentation.yaml -output presentation.html -play -record video.webm
```

### Command Line Options

- `-script`: Path to the presentation script file (required)
- `-output`: Output HTML file path (default: "presentation.html")
- `-play`: Play the presentation after generating (optional)
- `-record`: Path to save video recording (optional, requires -play)

## Script Format

Scripts are written in YAML format:

```yaml
title: "My Presentation"
duration: 120              # Total duration in seconds (optional)
default_time: 10           # Default slide duration in seconds (optional, default: 10)
slides:
  - title: "Welcome"
    content: "Welcome to my presentation"
    transcription: "This is the welcome slide where I introduce the topic..."
    duration: 8            # Override default duration for this slide
    image: "path/to/image.png"  # Optional image
  
  - title: "Main Content"
    content: "Key points:\n• Point 1\n• Point 2\n• Point 3"
    transcription: "In this slide, I'll cover the main points..."
    duration: 15
```

### Script Fields

- `title`: Presentation title
- `duration`: Total presentation duration (optional)
- `default_time`: Default duration for slides without explicit duration
- `slides`: Array of slide objects

### Slide Fields

- `title`: Slide title (required)
- `content`: Main slide content (optional)
- `transcription`: Explanatory text shown in transcription panel (optional)
- `duration`: Slide duration in seconds (optional, uses default_time if not specified)
- `image`: Path to image file (optional, supports PNG, JPG, GIF, WebP)

## Examples

See `example.yaml` for a complete example presentation.

## Controls

When playing a presentation in the browser:

- **Arrow Right / Spacebar**: Next slide
- **Arrow Left**: Previous slide
- **Enter**: Toggle play/pause
- **Play button**: Start/pause automatic playback
- **Previous/Next buttons**: Manual navigation

## Development

### Building

```bash
# Build the application
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linter (if available)
make lint

# Clean build artifacts
make clean
```

### Testing

Run the test suite:

```bash
# Run unit tests
go test -v -short

# Run all tests including integration tests
go test -v

# Run tests with coverage
make test-coverage
```

## Architecture

The project follows standard Go project structure:

- `cmd/rhesis/`: CLI application entry point
- `internal/script/`: YAML script parsing and validation
- `internal/generator/`: HTML presentation generation with templates
- `internal/player/`: Playwright integration for playback and recording
- `internal/version/`: Version information
- `Makefile`: Build automation and development tasks

## License

MIT License