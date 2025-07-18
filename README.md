# Rhesis

A Go program that reads a Markdown script and generates, plays, and records presentations in HTML with Playwright.

## Features

- **Markdown-based presentations**: Define slides with simple Markdown scripts
- **HTML generation**: Creates beautiful HTML presentations with CSS styling
- **Timing control**: Configurable slide durations for smooth transitions
- **Transcription support**: Display explanatory text alongside each slide
- **Audio generation**: Generate voice narration from transcriptions using ElevenLabs
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

For audio/video merging when using `-sound` with `-record`, you need ffmpeg:

```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt-get install ffmpeg

# Windows (using chocolatey)
choco install ffmpeg
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

# Generate presentation with audio narration
./bin/rhesis -script presentation.md -output presentation.html -sound -elevenlabs-key YOUR_API_KEY

# Generate with audio using environment variable for API key
export ELEVENLABS_API_KEY=your_api_key
./bin/rhesis -script presentation.md -output presentation.html -sound

# Generate, play, and record with audio narration
./bin/rhesis -script presentation.md -output presentation.html -sound -play -record video.mp4 -elevenlabs-key YOUR_API_KEY

# Reuse existing audio files (skip generation)
./bin/rhesis -script presentation.md -output presentation.html -sound -skip-audio-creation -play

# Generate and record in background mode (no visible browser)
./bin/rhesis -script presentation.md -output presentation.html -sound -play -record video.mp4 -background -elevenlabs-key YOUR_API_KEY

# Fuse existing video and audio files (without presentation generation)
./bin/rhesis -fuse -video input.webm -audio audio.mp3 -output output.mp4

# Fuse video with multiple audio files from a directory
./bin/rhesis -fuse -video input.webm -audio audio_dir/ -output output.mp4 -durations 10,15,20,10
```

### Command Line Options

#### Presentation Mode
- `-script`: Path to the presentation script file (required)
- `-output`: Output HTML file path (default: "presentation.html")
- `-play`: Play the presentation after generating (optional)
- `-background`: Run presentation in background/headless mode (optional, use with -play)
- `-record`: Path to save video recording (optional, requires -play)
- `-sound`: Generate audio narration from transcriptions using ElevenLabs (optional)
- `-skip-audio-creation`: Skip audio generation if audio files already exist (optional, use with -sound)
- `-elevenlabs-key`: ElevenLabs API key (optional, can also use ELEVENLABS_API_KEY env var)
- `-voice`: ElevenLabs voice ID (optional, defaults to Rachel voice)

#### Fuse Mode
- `-fuse`: Enable fuse mode to merge existing video and audio files (optional)
- `-video`: Input video file path (required in fuse mode)
- `-audio`: Input audio file path or directory containing audio files (required in fuse mode)
- `-output`: Output video file path (required in fuse mode)
- `-durations`: Comma-separated slide durations in seconds (required when `-audio` is a directory)

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

### Audio Generation

When using the `-sound` flag, the tool will:
1. Generate audio narration for each slide's transcription text using ElevenLabs API
2. Automatically adjust slide duration if the audio is longer than the specified duration
3. Play the audio synchronized with slide transitions during presentation playback
4. When combined with `-record`, automatically merge the audio with the video recording using ffmpeg

To use audio generation:
- Sign up for an ElevenLabs account and get an API key
- Set the API key via `-elevenlabs-key` flag or `ELEVENLABS_API_KEY` environment variable
- Optionally specify a voice ID with `-voice` flag (defaults to Rachel voice)
- Install ffmpeg if you want to record videos with audio narration

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

## Fuse Mode

The fuse mode allows you to merge existing video and audio files without generating a presentation. This is useful when you have:
- A video recording and want to add audio narration
- Multiple audio clips that need to be synchronized with video segments
- Pre-recorded content that needs audio/video synchronization

### Usage Examples

#### Single Audio File
Merge a single audio file with a video:
```bash
./bin/rhesis -fuse -video recording.webm -audio narration.mp3 -output final.mp4
```

#### Multiple Audio Files
Merge multiple audio files from a directory with specified timings:
```bash
./bin/rhesis -fuse -video recording.webm -audio audio_clips/ -output final.mp4 -durations 10,15,20,10
```

The `-durations` parameter specifies how long each audio clip should play for (in seconds). Audio files in the directory are processed in alphabetical order.

### Supported Formats
- **Video**: WebM, MP4
- **Audio**: MP3, WAV, M4A
- **Output**: WebM, MP4 (format determined by file extension)

The tool will automatically:
- Synchronize audio with video
- Adjust video framerate if audio is significantly longer than video
- Handle codec conversions between WebM and MP4 formats
- Add appropriate padding or trimming to match durations

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