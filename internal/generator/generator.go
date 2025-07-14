package generator

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmcarbo/rhesis/internal/script"
)

type HTMLGenerator struct {
	template *template.Template
}

func NewHTMLGenerator() *HTMLGenerator {
	tmpl := template.New("presentation").Funcs(template.FuncMap{
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
	})
	return &HTMLGenerator{
		template: template.Must(tmpl.Parse(htmlTemplate)),
	}
}

func (h *HTMLGenerator) GeneratePresentation(s *script.Script, outputPath string) error {
	data := struct {
		Script *script.Script
		Slides []SlideData
	}{
		Script: s,
		Slides: h.processSlides(s.Slides),
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return h.template.Execute(file, data)
}

type SlideData struct {
	script.Slide
	Index    int
	ImageSrc string
}

func (h *HTMLGenerator) processSlides(slides []script.Slide) []SlideData {
	result := make([]SlideData, len(slides))
	for i, slide := range slides {
		result[i] = SlideData{
			Slide: slide,
			Index: i,
		}
		if slide.Image != "" {
			result[i].ImageSrc = h.imageToBase64(slide.Image)
		}
	}
	return result
}

func (h *HTMLGenerator) imageToBase64(imagePath string) string {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return ""
	}

	if len(data) == 0 {
		return ""
	}

	ext := strings.ToLower(filepath.Ext(imagePath))
	var mimeType string
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	default:
		mimeType = "image/png"
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Encode(data))
}

func base64Encode(data []byte) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result strings.Builder

	for i := 0; i < len(data); i += 3 {
		var b1, b2, b3 byte
		b1 = data[i]
		if i+1 < len(data) {
			b2 = data[i+1]
		}
		if i+2 < len(data) {
			b3 = data[i+2]
		}

		result.WriteByte(base64Chars[b1>>2])
		result.WriteByte(base64Chars[((b1&0x03)<<4)|((b2&0xf0)>>4)])

		if i+1 < len(data) {
			result.WriteByte(base64Chars[((b2&0x0f)<<2)|((b3&0xc0)>>6)])
		} else {
			result.WriteByte('=')
		}

		if i+2 < len(data) {
			result.WriteByte(base64Chars[b3&0x3f])
		} else {
			result.WriteByte('=')
		}
	}

	return result.String()
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Script.Title}}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 0;
            background: #1a1a1a;
            color: #fff;
            overflow: hidden;
        }
        .presentation-container {
            display: flex;
            height: 100vh;
        }
        .slide-area {
            flex: 2;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            padding: 40px;
            background: linear-gradient(135deg, #2c3e50, #3498db);
        }
        .transcription-area {
            flex: 1;
            background: #2c3e50;
            padding: 20px;
            border-left: 2px solid #3498db;
            display: flex;
            flex-direction: column;
        }
        .slide {
            display: none;
            text-align: center;
            max-width: 100%;
        }
        .slide.active {
            display: block;
        }
        .slide h1 {
            font-size: 3em;
            margin-bottom: 30px;
            color: #fff;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.5);
        }
        .slide p {
            font-size: 1.5em;
            line-height: 1.6;
            margin-bottom: 20px;
            max-width: 800px;
        }
        .slide img {
            max-width: 100%;
            max-height: 60vh;
            border-radius: 10px;
            box-shadow: 0 4px 8px rgba(0,0,0,0.3);
        }
        .transcription-title {
            font-size: 1.2em;
            color: #3498db;
            margin-bottom: 10px;
            border-bottom: 1px solid #3498db;
            padding-bottom: 5px;
        }
        .transcription-content {
            font-size: 1em;
            line-height: 1.6;
            color: #ecf0f1;
            opacity: 0;
            transition: opacity 0.5s ease-in-out;
        }
        .transcription-content.active {
            opacity: 1;
        }
        .progress-bar {
            position: fixed;
            top: 0;
            left: 0;
            height: 4px;
            background: #3498db;
            transition: width 0.3s ease;
            z-index: 1000;
        }
        .slide-counter {
            position: fixed;
            bottom: 20px;
            right: 20px;
            background: rgba(0,0,0,0.7);
            padding: 10px 15px;
            border-radius: 20px;
            font-size: 0.9em;
        }
        .controls {
            position: fixed;
            bottom: 20px;
            left: 50%;
            transform: translateX(-50%);
            display: flex;
            gap: 10px;
            opacity: 1;
            transition: opacity 0.3s ease-in-out;
        }
        .controls.hidden {
            opacity: 0;
            pointer-events: none;
        }
        .btn {
            background: #3498db;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            font-size: 1em;
        }
        .btn:hover {
            background: #2980b9;
        }
        .btn:disabled {
            background: #7f8c8d;
            cursor: not-allowed;
        }
    </style>
</head>
<body>
    <div class="progress-bar" id="progressBar"></div>
    
    <div class="presentation-container">
        <div class="slide-area">
            {{range .Slides}}
            <div class="slide" data-duration="{{.Duration}}" data-index="{{.Index}}">
                <h1>{{.Title}}</h1>
                {{if .Content}}<p>{{.Content}}</p>{{end}}
                {{if .ImageSrc}}<img src="{{safeURL .ImageSrc}}" alt="{{.Title}}">{{end}}
            </div>
            {{end}}
        </div>
        
        <div class="transcription-area">
            <div class="transcription-title">Transcription</div>
            <div class="transcription-content" id="transcriptionContent">
                {{range .Slides}}
                <div class="transcription-slide" data-index="{{.Index}}" style="display: none;">
                    {{.Transcription}}
                </div>
                {{end}}
            </div>
        </div>
    </div>
    
    <div class="slide-counter">
        <span id="currentSlide">1</span> / <span id="totalSlides">{{len .Slides}}</span>
    </div>
    
    <div class="controls">
        <button class="btn" id="prevBtn" onclick="previousSlide()">Previous</button>
        <button class="btn" id="playBtn" onclick="togglePlayback()">Play</button>
        <button class="btn" id="nextBtn" onclick="nextSlide()">Next</button>
    </div>

    <script>
        let currentSlideIndex = 0;
        let isPlaying = false;
        let slideTimer = null;
        let startTime = null;
        let totalDuration = 0;
        
        // Expose variables to window for player to monitor
        window.isPlaying = false;
        window.currentSlideIndex = 0;
        
        const slides = document.querySelectorAll('.slide');
        const transcriptionSlides = document.querySelectorAll('.transcription-slide');
        const progressBar = document.getElementById('progressBar');
        const currentSlideSpan = document.getElementById('currentSlide');
        const playBtn = document.getElementById('playBtn');
        
        // Calculate total duration
        slides.forEach(slide => {
            totalDuration += parseInt(slide.dataset.duration) * 1000;
        });
        
        // Expose totalDuration to window for player
        window.totalDuration = totalDuration;
        
        function showSlide(index) {
            slides.forEach(slide => slide.classList.remove('active'));
            transcriptionSlides.forEach(trans => trans.style.display = 'none');
            
            // Remove active class from transcription content
            const transcriptionContent = document.getElementById('transcriptionContent');
            transcriptionContent.classList.remove('active');
            
            if (index >= 0 && index < slides.length) {
                slides[index].classList.add('active');
                transcriptionSlides[index].style.display = 'block';
                currentSlideIndex = index;
                window.currentSlideIndex = index;
                currentSlideSpan.textContent = index + 1;
                
                // Fade in transcription content
                setTimeout(() => {
                    transcriptionContent.classList.add('active');
                }, 100);
            }
        }
        
        function nextSlide() {
            if (currentSlideIndex < slides.length - 1) {
                showSlide(currentSlideIndex + 1);
            }
        }
        
        function previousSlide() {
            if (currentSlideIndex > 0) {
                showSlide(currentSlideIndex - 1);
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
            window.isPlaying = true;
            startTime = Date.now();
            playBtn.textContent = 'Pause';
            
            // Hide controls during automatic playback
            const controls = document.querySelector('.controls');
            controls.classList.add('hidden');
            
            function advanceSlide() {
                if (!isPlaying) return;
                
                const currentSlide = slides[currentSlideIndex];
                const duration = parseInt(currentSlide.dataset.duration) * 1000;
                
                console.log('Playing slide', currentSlideIndex + 1, 'of', slides.length, 'for', duration, 'ms');
                
                slideTimer = setTimeout(() => {
                    if (currentSlideIndex < slides.length - 1) {
                        console.log('Advancing to next slide');
                        nextSlide();
                        advanceSlide();
                    } else {
                        console.log('Reached last slide, stopping presentation');
                        stopPresentation();
                    }
                }, duration);
                
                updateProgress();
            }
            
            advanceSlide();
        }
        
        function stopPresentation() {
            isPlaying = false;
            window.isPlaying = false;
            playBtn.textContent = 'Play';
            if (slideTimer) {
                clearTimeout(slideTimer);
                slideTimer = null;
            }
            
            // Show controls when playback stops
            const controls = document.querySelector('.controls');
            controls.classList.remove('hidden');
        }
        
        function updateProgress() {
            if (!isPlaying || !startTime) return;
            
            const elapsed = Date.now() - startTime;
            const progress = Math.min(elapsed / totalDuration * 100, 100);
            progressBar.style.width = progress + '%';
            
            if (isPlaying) {
                requestAnimationFrame(updateProgress);
            }
        }
        
        // Keyboard controls
        document.addEventListener('keydown', (e) => {
            switch(e.key) {
                case 'ArrowRight':
                case ' ':
                    e.preventDefault();
                    nextSlide();
                    break;
                case 'ArrowLeft':
                    e.preventDefault();
                    previousSlide();
                    break;
                case 'Enter':
                    e.preventDefault();
                    togglePlayback();
                    break;
            }
        });
        
        // Initialize
        showSlide(0);
        
        // Auto-start indicator
        window.addEventListener('load', () => {
            document.body.setAttribute('data-ready', 'true');
        });
    </script>
</body>
</html>`
