package player

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmcarbo/rhesis/internal/generator"
	"github.com/jmcarbo/rhesis/internal/script"
)

func TestNewPresentationPlayer(t *testing.T) {
	player := NewPresentationPlayer()
	if player == nil {
		t.Error("Expected player to be created")
	}
}

func TestPresentationPlayerInvalidHTML(t *testing.T) {
	player := NewPresentationPlayer()
	err := player.PlayPresentation("nonexistent.html", "")
	if err == nil {
		t.Error("Expected error for nonexistent HTML file")
	}
}

func createTestHTML(t *testing.T) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Test Presentation</title>
</head>
<body data-ready="true">
    <div id="playBtn">Play</div>
    <script>
        window.isPlaying = false;
        window.totalDuration = 5000;
        
        document.getElementById('playBtn').addEventListener('click', function() {
            window.isPlaying = true;
            setTimeout(function() {
                window.isPlaying = false;
            }, 1000);
        });
    </script>
</body>
</html>`

	tmpFile, err := os.CreateTemp("", "test*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(html); err != nil {
		t.Fatalf("Failed to write HTML: %v", err)
	}
	tmpFile.Close()

	return tmpFile.Name()
}

func TestPlayPresentationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	htmlFile := createTestHTML(t)
	defer os.Remove(htmlFile)

	player := NewPresentationPlayer()

	done := make(chan error, 1)
	go func() {
		done <- player.PlayPresentation(htmlFile, "")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Error("Test timed out")
	}
}

func TestPlayPresentationWithRecording(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	htmlFile := createTestHTML(t)
	defer os.Remove(htmlFile)

	recordDir := t.TempDir()
	recordFile := filepath.Join(recordDir, "test_recording.webm")

	player := NewPresentationPlayer()

	done := make(chan error, 1)
	go func() {
		done <- player.PlayPresentation(htmlFile, recordFile)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Wait a moment for file to be written
		time.Sleep(500 * time.Millisecond)

		// Check if recording file was created
		if _, err := os.Stat(recordFile); os.IsNotExist(err) {
			t.Error("Expected recording file to be created")
		} else {
			// Verify file has content
			info, _ := os.Stat(recordFile)
			if info.Size() == 0 {
				t.Error("Recording file is empty")
			}
		}
	case <-time.After(30 * time.Second):
		t.Error("Test timed out")
	}
}

func TestPresentationPlayerCleanup(t *testing.T) {
	player := NewPresentationPlayer()
	player.cleanup()
}

// createMultiSlideTestHTML creates a test HTML with multiple slides
func createMultiSlideTestHTML(t *testing.T, numSlides int, slideDuration int) string {
	var slidesHTML strings.Builder
	var transcriptionHTML strings.Builder
	totalDuration := numSlides * slideDuration * 1000

	for i := 0; i < numSlides; i++ {
		slidesHTML.WriteString(fmt.Sprintf(`
			<div class="slide" data-duration="%d" data-index="%d">
				<h1>Slide %d</h1>
				<p>Content for slide %d</p>
			</div>`, slideDuration, i, i+1, i+1))

		transcriptionHTML.WriteString(fmt.Sprintf(`
			<div class="transcription-slide" data-index="%d" style="display: none;">
				Transcription for slide %d
			</div>`, i, i+1))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<title>Multi-Slide Test Presentation</title>
	<style>
		.slide { display: none; }
		.slide.active { display: block; }
	</style>
</head>
<body>
	<div class="progress-bar" id="progressBar"></div>
	<div class="presentation-container">
		<div class="slide-area">%s</div>
		<div class="transcription-area">
			<div class="transcription-content" id="transcriptionContent">%s</div>
		</div>
	</div>
	<div class="slide-counter">
		<span id="currentSlide">1</span> / <span id="totalSlides">%d</span>
	</div>
	<div class="controls">
		<button class="btn" id="playBtn" onclick="togglePlayback()">Play</button>
	</div>
	<script>
		let currentSlideIndex = 0;
		let isPlaying = false;
		let slideTimer = null;
		let startTime = null;
		let totalDuration = %d;
		let slidesPlayed = [];
		
		const slides = document.querySelectorAll('.slide');
		const transcriptionSlides = document.querySelectorAll('.transcription-slide');
		const progressBar = document.getElementById('progressBar');
		const currentSlideSpan = document.getElementById('currentSlide');
		const playBtn = document.getElementById('playBtn');
		
		function showSlide(index) {
			slides.forEach(slide => slide.classList.remove('active'));
			transcriptionSlides.forEach(trans => trans.style.display = 'none');
			
			if (index >= 0 && index < slides.length) {
				slides[index].classList.add('active');
				transcriptionSlides[index].style.display = 'block';
				currentSlideIndex = index;
				currentSlideSpan.textContent = index + 1;
				slidesPlayed.push(index);
			}
		}
		
		function nextSlide() {
			if (currentSlideIndex < slides.length - 1) {
				showSlide(currentSlideIndex + 1);
			}
		}
		
		function togglePlayback() {
			if (isPlaying) {
				stopPresentation();
			} else {
				startPresentation();
			}
		}
		
		function startPresentation() {
			isPlaying = true;
			startTime = Date.now();
			playBtn.textContent = 'Pause';
			
			function advanceSlide() {
				if (!isPlaying) return;
				
				const currentSlide = slides[currentSlideIndex];
				const duration = parseInt(currentSlide.dataset.duration) * 1000;
				
				slideTimer = setTimeout(() => {
					if (currentSlideIndex < slides.length - 1) {
						nextSlide();
						advanceSlide();
					} else {
						stopPresentation();
					}
				}, duration);
				
				updateProgress();
			}
			
			advanceSlide();
		}
		
		function stopPresentation() {
			isPlaying = false;
			playBtn.textContent = 'Play';
			if (slideTimer) {
				clearTimeout(slideTimer);
				slideTimer = null;
			}
		}
		
		function updateProgress() {
			if (!isPlaying || !startTime) return;
			
			const elapsed = Date.now() - startTime;
			const progress = Math.min(elapsed / totalDuration * 100, 100);
			progressBar.style.width = progress + '%%';
			
			if (isPlaying) {
				requestAnimationFrame(updateProgress);
			}
		}
		
		// Initialize
		showSlide(0);
		
		// Auto-start indicator
		window.addEventListener('load', () => {
			document.body.setAttribute('data-ready', 'true');
		});
		
		// Expose for testing
		window.isPlaying = false;
		window.totalDuration = totalDuration;
		window.getCurrentSlideIndex = () => currentSlideIndex;
		window.getSlidesPlayed = () => slidesPlayed;
	</script>
</body>
</html>`, slidesHTML.String(), transcriptionHTML.String(), numSlides, totalDuration)

	tmpFile, err := os.CreateTemp("", "test_multi_slide*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(html); err != nil {
		t.Fatalf("Failed to write HTML: %v", err)
	}
	tmpFile.Close()

	return tmpFile.Name()
}

// TestPlaybackCompletesAllSlides verifies that all slides are played
func TestPlaybackCompletesAllSlides(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	numSlides := 3
	slideDuration := 1 // 1 second per slide
	htmlFile := createMultiSlideTestHTML(t, numSlides, slideDuration)
	defer os.Remove(htmlFile)

	player := NewPresentationPlayer()

	// Initialize player to get access to page
	if err := player.initialize(); err != nil {
		t.Fatalf("Failed to initialize player: %v", err)
	}
	defer player.cleanup()

	// Navigate to the test page
	absolutePath, _ := filepath.Abs(htmlFile)
	fileURL := fmt.Sprintf("file://%s", absolutePath)
	if _, err := player.page.Goto(fileURL); err != nil {
		t.Fatalf("Failed to navigate to test page: %v", err)
	}

	// Wait for page to be ready
	if _, err := player.page.WaitForSelector("[data-ready='true']"); err != nil {
		t.Fatalf("Failed to wait for page ready: %v", err)
	}

	// Click play button
	playBtn, err := player.page.QuerySelector("#playBtn")
	if err != nil {
		t.Fatalf("Failed to find play button: %v", err)
	}
	if err := playBtn.Click(); err != nil {
		t.Fatalf("Failed to click play button: %v", err)
	}

	// Wait for presentation to complete (with buffer)
	expectedDuration := time.Duration(numSlides*slideDuration)*time.Second + 2*time.Second
	time.Sleep(expectedDuration)

	// Check if presentation stopped
	isPlaying, err := player.page.Evaluate(`() => window.isPlaying`)
	if err != nil {
		t.Fatalf("Failed to check isPlaying: %v", err)
	}
	if playing, ok := isPlaying.(bool); ok && playing {
		t.Error("Presentation should have stopped after all slides")
	}

	// Check current slide index
	currentIndex, err := player.page.Evaluate(`() => window.getCurrentSlideIndex()`)
	if err != nil {
		t.Fatalf("Failed to get current slide index: %v", err)
	}
	// The getCurrentSlideIndex function might return the value directly as int
	switch v := currentIndex.(type) {
	case float64:
		if int(v) != numSlides-1 {
			t.Errorf("Expected to be on last slide (index %d), but on slide index %d", numSlides-1, int(v))
		}
	case int:
		if v != numSlides-1 {
			t.Errorf("Expected to be on last slide (index %d), but on slide index %d", numSlides-1, v)
		}
	default:
		t.Errorf("Unexpected type for slide index: %T, value: %v", currentIndex, currentIndex)
	}

	// Check slides played
	slidesPlayed, err := player.page.Evaluate(`() => window.getSlidesPlayed()`)
	if err != nil {
		t.Fatalf("Failed to get slides played: %v", err)
	}
	if played, ok := slidesPlayed.([]interface{}); ok {
		if len(played) < numSlides {
			t.Errorf("Expected at least %d slides to be played, but only %d were played", numSlides, len(played))
		}
	}
}

// TestPlaybackStopsAtFirstSlide tests if presentation incorrectly stops after first slide
func TestPlaybackStopsAtFirstSlide(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	numSlides := 3
	slideDuration := 2 // 2 seconds per slide
	htmlFile := createMultiSlideTestHTML(t, numSlides, slideDuration)
	defer os.Remove(htmlFile)

	player := NewPresentationPlayer()

	// Initialize player
	if err := player.initialize(); err != nil {
		t.Fatalf("Failed to initialize player: %v", err)
	}
	defer player.cleanup()

	// Navigate to the test page
	absolutePath, _ := filepath.Abs(htmlFile)
	fileURL := fmt.Sprintf("file://%s", absolutePath)
	if _, err := player.page.Goto(fileURL); err != nil {
		t.Fatalf("Failed to navigate to test page: %v", err)
	}

	// Wait for page to be ready
	if _, err := player.page.WaitForSelector("[data-ready='true']"); err != nil {
		t.Fatalf("Failed to wait for page ready: %v", err)
	}

	// Click play button
	playBtn, err := player.page.QuerySelector("#playBtn")
	if err != nil {
		t.Fatalf("Failed to find play button: %v", err)
	}
	if err := playBtn.Click(); err != nil {
		t.Fatalf("Failed to click play button: %v", err)
	}

	// Wait for more than first slide duration but less than full presentation
	time.Sleep(time.Duration(slideDuration+1) * time.Second)

	// Check if still playing
	isPlaying, err := player.page.Evaluate(`() => window.isPlaying`)
	if err != nil {
		t.Fatalf("Failed to check isPlaying: %v", err)
	}

	// Should still be playing if not stuck on first slide
	if playing, ok := isPlaying.(bool); !ok || !playing {
		// Check which slide we're on
		currentIndex, _ := player.page.Evaluate(`() => window.getCurrentSlideIndex()`)
		if idx, ok := currentIndex.(float64); ok && int(idx) == 0 {
			t.Error("Presentation stopped after first slide - this is the bug!")
		}
	}

	// Check current slide index - should be on second slide or later
	currentIndex, err := player.page.Evaluate(`() => window.getCurrentSlideIndex()`)
	if err != nil {
		t.Fatalf("Failed to get current slide index: %v", err)
	}
	if idx, ok := currentIndex.(float64); ok && int(idx) == 0 {
		t.Error("Presentation is still on first slide after waiting - slides are not advancing")
	}
}

// TestPlayPresentationMethod tests the actual PlayPresentation method behavior
func TestPlayPresentationMethod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	numSlides := 2
	slideDuration := 1 // 1 second per slide
	htmlFile := createMultiSlideTestHTML(t, numSlides, slideDuration)
	defer os.Remove(htmlFile)

	player := NewPresentationPlayer()

	// Run PlayPresentation in a goroutine
	done := make(chan error, 1)
	start := time.Now()
	go func() {
		done <- player.PlayPresentation(htmlFile, "")
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		elapsed := time.Since(start)
		if err != nil {
			t.Errorf("PlayPresentation returned error: %v", err)
		}

		// Check if it completed too quickly (might indicate early exit)
		minExpectedDuration := time.Duration(numSlides*slideDuration) * time.Second
		if elapsed < minExpectedDuration {
			t.Errorf("PlayPresentation completed too quickly: %v (expected at least %v)", elapsed, minExpectedDuration)
		}
	case <-time.After(30 * time.Second):
		t.Error("Test timed out")
	}
}

// TestProgressBarUpdate tests if progress bar updates correctly during playback
func TestProgressBarUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	numSlides := 2
	slideDuration := 2 // 2 seconds per slide
	htmlFile := createMultiSlideTestHTML(t, numSlides, slideDuration)
	defer os.Remove(htmlFile)

	player := NewPresentationPlayer()

	// Initialize player
	if err := player.initialize(); err != nil {
		t.Fatalf("Failed to initialize player: %v", err)
	}
	defer player.cleanup()

	// Navigate to the test page
	absolutePath, _ := filepath.Abs(htmlFile)
	fileURL := fmt.Sprintf("file://%s", absolutePath)
	if _, err := player.page.Goto(fileURL); err != nil {
		t.Fatalf("Failed to navigate to test page: %v", err)
	}

	// Wait for page to be ready
	if _, err := player.page.WaitForSelector("[data-ready='true']"); err != nil {
		t.Fatalf("Failed to wait for page ready: %v", err)
	}

	// Click play button
	playBtn, err := player.page.QuerySelector("#playBtn")
	if err != nil {
		t.Fatalf("Failed to find play button: %v", err)
	}
	if err := playBtn.Click(); err != nil {
		t.Fatalf("Failed to click play button: %v", err)
	}

	// Wait a bit for progress to start
	time.Sleep(1 * time.Second)

	// Check progress bar width
	progressWidth, err := player.page.Evaluate(`() => document.getElementById('progressBar').style.width`)
	if err != nil {
		t.Fatalf("Failed to get progress bar width: %v", err)
	}

	if width, ok := progressWidth.(string); ok {
		if width == "" || width == "0%" {
			t.Error("Progress bar is not updating during playback")
		}
	}
}

// TestPlaybackWithGeneratedHTML tests playback with HTML generated by the actual generator
func TestPlaybackWithGeneratedHTML(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Import required packages for this test
	scriptContent := &script.Script{
		Title:       "Test Presentation",
		Duration:    10,
		DefaultTime: 2,
		Slides: []script.Slide{
			{
				Title:         "First Slide",
				Content:       "This is the first slide",
				Transcription: "Welcome to the first slide",
				Duration:      2,
			},
			{
				Title:         "Second Slide",
				Content:       "This is the second slide",
				Transcription: "Now on the second slide",
				Duration:      2,
			},
			{
				Title:         "Third Slide",
				Content:       "This is the third slide",
				Transcription: "Finally, the third slide",
				Duration:      2,
			},
		},
	}

	// Generate HTML using the actual generator
	generator := generator.NewHTMLGenerator()
	htmlFile := filepath.Join(t.TempDir(), "test_generated.html")
	if err := generator.GeneratePresentation(scriptContent, htmlFile); err != nil {
		t.Fatalf("Failed to generate HTML: %v", err)
	}

	player := NewPresentationPlayer()

	// Run PlayPresentation
	done := make(chan error, 1)
	start := time.Now()
	go func() {
		done <- player.PlayPresentation(htmlFile, "")
	}()

	// Wait for completion
	select {
	case err := <-done:
		elapsed := time.Since(start)
		if err != nil {
			t.Errorf("PlayPresentation returned error: %v", err)
		}

		// Should take at least 6 seconds (3 slides * 2 seconds each)
		if elapsed < 6*time.Second {
			t.Errorf("Presentation completed too quickly: %v (expected at least 6s)", elapsed)
		}
	case <-time.After(30 * time.Second):
		t.Error("Test timed out")
	}
}

// TestPlaybackInterruption tests stopping presentation mid-playback
func TestPlaybackInterruption(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	numSlides := 5
	slideDuration := 3 // 3 seconds per slide
	htmlFile := createMultiSlideTestHTML(t, numSlides, slideDuration)
	defer os.Remove(htmlFile)

	player := NewPresentationPlayer()

	// Initialize player
	if err := player.initialize(); err != nil {
		t.Fatalf("Failed to initialize player: %v", err)
	}
	defer player.cleanup()

	// Navigate to the test page
	absolutePath, _ := filepath.Abs(htmlFile)
	fileURL := fmt.Sprintf("file://%s", absolutePath)
	if _, err := player.page.Goto(fileURL); err != nil {
		t.Fatalf("Failed to navigate to test page: %v", err)
	}

	// Wait for page to be ready
	if _, err := player.page.WaitForSelector("[data-ready='true']"); err != nil {
		t.Fatalf("Failed to wait for page ready: %v", err)
	}

	// Click play button
	playBtn, err := player.page.QuerySelector("#playBtn")
	if err != nil {
		t.Fatalf("Failed to find play button: %v", err)
	}
	if err := playBtn.Click(); err != nil {
		t.Fatalf("Failed to click play button: %v", err)
	}

	// Wait until we're on second or third slide
	time.Sleep(time.Duration(slideDuration+1) * time.Second)

	// Click play button again to pause
	if err := playBtn.Click(); err != nil {
		t.Fatalf("Failed to click play button to pause: %v", err)
	}

	// Check if paused
	isPlaying, err := player.page.Evaluate(`() => window.isPlaying`)
	if err != nil {
		t.Fatalf("Failed to check isPlaying: %v", err)
	}
	if playing, ok := isPlaying.(bool); !ok || playing {
		t.Error("Presentation should be paused after clicking play button again")
	}

	// Get current slide to ensure we stopped mid-presentation
	currentIndex, err := player.page.Evaluate(`() => window.getCurrentSlideIndex()`)
	if err != nil {
		t.Fatalf("Failed to get current slide index: %v", err)
	}
	if idx, ok := currentIndex.(float64); ok {
		if int(idx) == 0 || int(idx) == numSlides-1 {
			t.Errorf("Expected to be on a middle slide, but on slide %d", int(idx))
		}
	}
}

func TestPresentationPlayerInitializeAndCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	player := NewPresentationPlayer()

	err := player.initialize()
	if err != nil {
		t.Errorf("Failed to initialize player: %v", err)
		return
	}

	if player.pw == nil {
		t.Error("Expected playwright to be initialized")
	}
	if player.browser == nil {
		t.Error("Expected browser to be initialized")
	}
	if player.context == nil {
		t.Error("Expected context to be initialized")
	}
	if player.page == nil {
		t.Error("Expected page to be initialized")
	}

	player.cleanup()
}
