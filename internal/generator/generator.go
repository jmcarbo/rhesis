package generator

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmcarbo/rhesis/internal/script"
	"github.com/jmcarbo/rhesis/internal/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
)

type HTMLGenerator struct {
	template     *template.Template
	styleManager *styles.StyleManager
	markdown     goldmark.Markdown
}

func NewHTMLGenerator() *HTMLGenerator {
	tmpl := template.New("presentation").Funcs(template.FuncMap{
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	})

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
					chromahtml.WithLineNumbers(false),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	return &HTMLGenerator{
		template:     template.Must(tmpl.Parse(htmlTemplate)),
		styleManager: styles.NewStyleManager(),
		markdown:     md,
	}
}

func (h *HTMLGenerator) GeneratePresentation(s *script.Script, outputPath string, theme string) error {
	styleCSS, err := h.styleManager.GetStyle(theme)
	if err != nil {
		return fmt.Errorf("failed to get style: %w", err)
	}

	data := struct {
		Script *script.Script
		Slides []SlideData
		Style  template.CSS
	}{
		Script: s,
		Slides: h.processSlides(s.Slides),
		Style:  template.CSS(styleCSS),
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
	Index             int
	ImageSrc          string
	ContentHTML       template.HTML
	TranscriptionHTML template.HTML
}

func (h *HTMLGenerator) processSlides(slides []script.Slide) []SlideData {
	result := make([]SlideData, len(slides))
	for i, slide := range slides {
		result[i] = SlideData{
			Slide:             slide,
			Index:             i,
			ContentHTML:       h.renderMarkdown(slide.Content),
			TranscriptionHTML: h.renderMarkdown(slide.Transcription),
		}
		if slide.Image != "" {
			result[i].ImageSrc = h.imageToBase64(slide.Image)
		}
	}
	return result
}

func (h *HTMLGenerator) renderMarkdown(content string) template.HTML {
	var buf bytes.Buffer
	if err := h.markdown.Convert([]byte(content), &buf); err != nil {
		// If markdown rendering fails, return the original content escaped
		return template.HTML(template.HTMLEscapeString(content))
	}
	return template.HTML(buf.String())
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
        {{.Style}}
    </style>
</head>
<body>
    <div class="progress-bar" id="progressBar"></div>
    
    <div class="presentation-container">
        <div class="slide-area">
            {{range .Slides}}
            <div class="slide" data-duration="{{.Duration}}" data-index="{{.Index}}">
                <h1>{{.Title}}</h1>
                {{if .Content}}<div class="slide-content">{{.ContentHTML}}</div>{{end}}
                {{if .ImageSrc}}<img src="{{safeURL .ImageSrc}}" alt="{{.Title}}">{{end}}
            </div>
            {{end}}
        </div>
        
        <div class="transcription-area">
            <div class="transcription-title">Transcription</div>
            <div class="transcription-content" id="transcriptionContent">
                {{range .Slides}}
                <div class="transcription-slide" data-index="{{.Index}}" style="display: none;">
                    {{.TranscriptionHTML}}
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
