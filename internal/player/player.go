package player

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/playwright-community/playwright-go"
)

type PresentationPlayer struct {
	pw         *playwright.Playwright
	browser    playwright.Browser
	context    playwright.BrowserContext
	page       playwright.Page
	recordPath string
}

func NewPresentationPlayer() *PresentationPlayer {
	return &PresentationPlayer{}
}

func (p *PresentationPlayer) PlayPresentation(htmlPath, recordPath string) error {
	return p.PlayPresentationWithOptions(htmlPath, recordPath, false)
}

func (p *PresentationPlayer) PlayPresentationWithOptions(htmlPath, recordPath string, headless bool) error {
	p.recordPath = recordPath

	if err := p.initializeWithOptions(headless); err != nil {
		return fmt.Errorf("failed to initialize player: %w", err)
	}
	defer p.cleanup()

	absolutePath, err := filepath.Abs(htmlPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fileURL := fmt.Sprintf("file://%s", absolutePath)

	if _, err := p.page.Goto(fileURL); err != nil {
		return fmt.Errorf("failed to load presentation: %w", err)
	}

	if _, err := p.page.WaitForSelector("[data-ready='true']"); err != nil {
		return fmt.Errorf("failed to wait for presentation ready: %w", err)
	}

	if err := p.playPresentation(); err != nil {
		return fmt.Errorf("failed to play presentation: %w", err)
	}

	return nil
}

func (p *PresentationPlayer) initialize() error {
	return p.initializeWithOptions(false)
}

func (p *PresentationPlayer) initializeWithOptions(headless bool) error {
	pw, err := playwright.Run()
	if err != nil {
		return err
	}
	p.pw = pw

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
		Args: []string{
			"--enable-web-bluetooth",
			"--use-fake-ui-for-media-stream",
			"--use-fake-device-for-media-stream",
		},
	})
	if err != nil {
		return err
	}
	p.browser = browser

	contextOptions := playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
	}

	// Enable video recording if record path is specified
	if p.recordPath != "" {
		contextOptions.RecordVideo = &playwright.RecordVideo{
			Dir: filepath.Dir(p.recordPath),
			Size: &playwright.Size{
				Width:  1920,
				Height: 1080,
			},
		}
	}

	context, err := browser.NewContext(contextOptions)
	if err != nil {
		return err
	}
	p.context = context

	page, err := context.NewPage()
	if err != nil {
		return err
	}
	p.page = page

	return nil
}

func (p *PresentationPlayer) cleanup() {
	// Save video if recording was enabled
	if p.recordPath != "" && p.page != nil {
		video := p.page.Video()
		if video != nil {
			// Get the actual video path from the page
			videoPath, err := video.Path()
			if err != nil {
				fmt.Printf("Warning: failed to get video path: %v\n", err)
				return
			}

			// Close the page to finalize the video
			p.page.Close()
			p.page = nil

			// Close context to ensure video is saved
			if p.context != nil {
				p.context.Close()
				p.context = nil
			}

			// Wait a moment for video to be written
			time.Sleep(500 * time.Millisecond)

			// Move video from temporary location to desired path
			if videoPath != "" && videoPath != p.recordPath {
				if err := os.Rename(videoPath, p.recordPath); err != nil {
					// Try copying if rename fails (e.g., across filesystems)
					if err := copyFile(videoPath, p.recordPath); err != nil {
						fmt.Printf("Warning: failed to save video to %s: %v\n", p.recordPath, err)
					}
				}
			}
		}
	}

	// Close page if not already closed
	if p.page != nil {
		p.page.Close()
	}

	// Close context if not already closed
	if p.context != nil {
		p.context.Close()
	}

	if p.browser != nil {
		p.browser.Close()
	}
	if p.pw != nil {
		p.pw.Stop()
	}
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

func (p *PresentationPlayer) playPresentation() error {
	playBtn, err := p.page.QuerySelector("#playBtn")
	if err != nil {
		return err
	}

	// Get total duration before starting
	totalDuration, err := p.page.Evaluate(`() => {
		let total = 0;
		document.querySelectorAll('.slide').forEach(slide => {
			total += parseInt(slide.dataset.duration) * 1000;
		});
		return total;
	}`)
	if err != nil {
		return err
	}

	// Handle both float64 and int types
	var totalDurationMs float64
	switch v := totalDuration.(type) {
	case float64:
		totalDurationMs = v
	case int:
		totalDurationMs = float64(v)
	default:
		return fmt.Errorf("unexpected duration type: %T", totalDuration)
	}
	fmt.Printf("Total presentation duration: %.1f seconds\n", totalDurationMs/1000)

	if err := playBtn.Click(); err != nil {
		return err
	}

	// Wait a moment for playback to start
	time.Sleep(500 * time.Millisecond)

	// Monitor playback status
	startTime := time.Now()
	for {
		isPlaying, err := p.page.Evaluate(`() => window.isPlaying`)
		if err != nil {
			return err
		}

		// Get current slide index
		slideIndex, _ := p.page.Evaluate(`() => window.currentSlideIndex`)
		totalSlides, _ := p.page.Evaluate(`() => document.querySelectorAll('.slide').length`)

		if idx, ok := slideIndex.(float64); ok {
			if total, ok := totalSlides.(float64); ok {
				fmt.Printf("\rPlaying slide %d of %d", int(idx)+1, int(total))
			}
		}

		// Handle case where isPlaying might be nil or not a bool
		playing, ok := isPlaying.(bool)
		if !ok || !playing {
			// Check if we've played for at least most of the expected duration
			elapsed := time.Since(startTime)
			if elapsed < time.Duration(totalDurationMs*0.9)*time.Millisecond {
				fmt.Printf("\nPresentation stopped early after %.1f seconds\n", elapsed.Seconds())
			} else {
				fmt.Printf("\nPresentation completed\n")
			}
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Wait a bit more to ensure video capture completes
	time.Sleep(2 * time.Second)

	// Check if we might have hit a recording limit
	if totalDurationMs > 180000 { // More than 3 minutes
		fmt.Printf("\nWARNING: Presentation is longer than 3 minutes (%.1fs).\n", totalDurationMs/1000)
		fmt.Printf("Some browsers/codecs may truncate long video recordings.\n")
		fmt.Printf("If the video is shorter than expected, consider:\n")
		fmt.Printf("- Breaking the presentation into smaller segments\n")
		fmt.Printf("- Using a different output format (WebM vs MP4)\n")
		fmt.Printf("- Recording without video and adding audio separately\n")
	}

	return nil
}

func (p *PresentationPlayer) GetPresentationDuration() (time.Duration, error) {
	duration, err := p.page.Evaluate(`() => window.totalDuration`)
	if err != nil {
		return 0, err
	}

	return time.Duration(duration.(float64)) * time.Millisecond, nil
}
