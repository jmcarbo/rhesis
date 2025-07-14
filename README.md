# Rhesis

A Go program that reads a Markdown script and generates, plays, and records presentations in HTML with Playwright.

## Features

- **Markdown-based presentations**: Define slides with simple Markdown scripts
- **HTML generation**: Creates beautiful HTML presentations with CSS styling
- **Timing control**: Configurable slide durations for smooth transitions
- **Transcription support**: Display explanatory text alongside each slide
- **Image support**: Embed images directly in slides (PNG, JPG, GIF, WebP)
- **Automatic playback**: Play presentations automatically with proper timing
- **Recording capability**: Record presentations to video files (WebM, MP4) using Playwright
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
./bin/rhesis -script presentation.md -output presentation.html

# Generate and play presentation
./bin/rhesis -script presentation.md -output presentation.html -play

# Generate, play, and record presentation (WebM format)
./bin/rhesis -script presentation.md -output presentation.html -play -record video.webm

# Generate, play, and record presentation (MP4 format)
./bin/rhesis -script presentation.md -output presentation.html -play -record video.mp4
```

### Command Line Options

- `-script`: Path to the presentation script file (required)
- `-output`: Output HTML file path (default: "presentation.html")
- `-play`: Play the presentation after generating (optional)
- `-record`: Path to save video recording (optional, requires -play)

## Script Format

Scripts are written in Markdown format with a simple structure:

```markdown
# Presentation Title

Duration: 120              # Total duration in seconds (optional)
Default time: 10           # Default slide duration in seconds (optional, default: 10)

## Slide Title

Duration: 15               # Override default duration for this slide

This is the slide content. You can use **bold**, *italic*, and other Markdown formatting.

- Bullet points
- Work great too

---

This is the transcription text that appears in the side panel. 
It provides additional context or narration for the slide.

## Another Slide

Image: path/to/image.png   # Optional image

Code blocks are supported:

` + "```python" + `
def hello():
    print("Hello, World!")
` + "```" + `

---

Transcription for the second slide goes here.
```

### Markdown Structure

1. **Presentation Title**: Use a single H1 (`# Title`) for the presentation title
2. **Metadata**: Place optional metadata after the title:
   - `Duration: N` - Total presentation duration in seconds
   - `Default time: N` - Default slide duration in seconds
3. **Slides**: Each H2 (`## Slide Title`) starts a new slide
4. **Slide Options**: Place these after the slide title:
   - `Duration: N` - Override duration for this specific slide
   - `Image: path/to/image` - Add an image to the slide
5. **Content**: Everything after the slide options until `---` is slide content
6. **Transcription**: Text after `---` until the next slide is the transcription

### Features

- **Markdown Formatting**: Full support for Markdown in slide content
- **Code Blocks**: Syntax-highlighted code blocks with language specification
- **Lists**: Both ordered and unordered lists, including nested lists
- **Images**: Embed images using the `Image:` directive
- **Links**: Standard Markdown links are supported
- **Blockquotes**: Use `>` for quotations

## Examples

See `example.md` for a complete example presentation about Go programming.

### Simple Example

```markdown
# My First Presentation

Default time: 8

## Welcome

Welcome to my presentation!

---

Thank you for joining me today. In this presentation, we'll explore...

## Main Points

Duration: 12

- First important point
- Second important point
- Third important point

---

Let me elaborate on each of these points...
```

### Code-Heavy Presentation

```markdown
# Programming Workshop

## Hello World Examples

Duration: 15

Let's compare Hello World in different languages:

` + "```go" + `
// Go
package main
import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

` + "```python" + `
# Python
print("Hello, World!")
` + "```" + `

---

Notice how Go requires more boilerplate but provides type safety...
```

## Controls

When playing a presentation in the browser:

- **Arrow Right / Spacebar**: Next slide
- **Arrow Left**: Previous slide
- **Enter**: Toggle play/pause
- **Play button**: Start/pause automatic playback
- **Previous/Next buttons**: Manual navigation

## Recording Presentations

The recording feature allows you to capture your presentations as video files:

- **Supported formats**: WebM and MP4
- **Resolution**: Full HD (1920x1080)
- **Requirements**: Playwright browsers must be installed
- **Usage**: Use the `-record` flag with a filename ending in `.webm` or `.mp4`

The recording captures the entire presentation playback, including slide transitions and timing. Videos are saved in the specified format and location.

Example:
```bash
# Record a presentation as WebM
./bin/rhesis -script demo.md -play -record output/demo.webm

# Record a presentation as MP4
./bin/rhesis -script demo.md -play -record output/demo.mp4
```

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
- `internal/script/`: Markdown script parsing and validation
- `internal/generator/`: HTML presentation generation with templates
- `internal/player/`: Playwright integration for playback and recording
- `internal/version/`: Version information
- `Makefile`: Build automation and development tasks

## License

MIT License