package player

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/playwright-community/playwright-go"
)

type PresentationPlayer struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	page    playwright.Page
}

func NewPresentationPlayer() *PresentationPlayer {
	return &PresentationPlayer{}
}

func (p *PresentationPlayer) PlayPresentation(htmlPath, recordPath string) error {
	if err := p.initialize(); err != nil {
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

	if recordPath != "" {
		if err := p.startRecording(recordPath); err != nil {
			return fmt.Errorf("failed to start recording: %w", err)
		}
	}

	if err := p.playPresentation(); err != nil {
		return fmt.Errorf("failed to play presentation: %w", err)
	}

	if recordPath != "" {
		if err := p.stopRecording(); err != nil {
			return fmt.Errorf("failed to stop recording: %w", err)
		}
	}

	return nil
}

func (p *PresentationPlayer) initialize() error {
	pw, err := playwright.Run()
	if err != nil {
		return err
	}
	p.pw = pw

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
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

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
	})
	if err != nil {
		return err
	}

	page, err := context.NewPage()
	if err != nil {
		return err
	}
	p.page = page

	return nil
}

func (p *PresentationPlayer) cleanup() {
	if p.page != nil {
		p.page.Close()
	}
	if p.browser != nil {
		p.browser.Close()
	}
	if p.pw != nil {
		p.pw.Stop()
	}
}

func (p *PresentationPlayer) startRecording(recordPath string) error {
	// Recording is not supported in the current implementation
	// This would require setting up video recording during context creation
	return nil
}

func (p *PresentationPlayer) stopRecording() error {
	return nil
}

func (p *PresentationPlayer) playPresentation() error {
	playBtn, err := p.page.QuerySelector("#playBtn")
	if err != nil {
		return err
	}

	if err := playBtn.Click(); err != nil {
		return err
	}

	for {
		isPlaying, err := p.page.Evaluate(`() => window.isPlaying`)
		if err != nil {
			return err
		}

		if !isPlaying.(bool) {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(2 * time.Second)
	return nil
}

func (p *PresentationPlayer) GetPresentationDuration() (time.Duration, error) {
	duration, err := p.page.Evaluate(`() => window.totalDuration`)
	if err != nil {
		return 0, err
	}

	return time.Duration(duration.(float64)) * time.Millisecond, nil
}
