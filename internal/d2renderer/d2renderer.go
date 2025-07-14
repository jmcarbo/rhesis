package d2renderer

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

// Renderer handles D2 diagram rendering
type Renderer struct {
	ruler          *textmeasure.Ruler
	layoutResolver func(engine string) (d2graph.LayoutGraph, error)
}

// NewRenderer creates a new D2 renderer
func NewRenderer() (*Renderer, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, fmt.Errorf("failed to create text ruler: %w", err)
	}

	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return d2dagrelayout.DefaultLayout, nil
	}

	return &Renderer{
		ruler:          ruler,
		layoutResolver: layoutResolver,
	}, nil
}

// RenderToSVG renders D2 code to SVG
func (r *Renderer) RenderToSVG(d2Code string) (string, error) {
	renderOpts := &d2svg.RenderOpts{
		Pad:     go2.Pointer(int64(20)),
		ThemeID: &d2themescatalog.NeutralDefault.ID,
	}

	compileOpts := &d2lib.CompileOptions{
		LayoutResolver: r.layoutResolver,
		Ruler:          r.ruler,
	}

	ctx := log.WithDefault(context.Background())
	diagram, _, err := d2lib.Compile(ctx, d2Code, compileOpts, renderOpts)
	if err != nil {
		return "", fmt.Errorf("failed to compile D2 diagram: %w", err)
	}

	svg, err := d2svg.Render(diagram, renderOpts)
	if err != nil {
		return "", fmt.Errorf("failed to render D2 diagram: %w", err)
	}

	return string(svg), nil
}

// RenderToDataURL renders D2 code to a data URL for embedding in HTML
func (r *Renderer) RenderToDataURL(d2Code string) (string, error) {
	svg, err := r.RenderToSVG(d2Code)
	if err != nil {
		return "", err
	}

	// Encode SVG as base64 data URL
	encoded := base64.StdEncoding.EncodeToString([]byte(svg))
	dataURL := fmt.Sprintf("data:image/svg+xml;base64,%s", encoded)

	return dataURL, nil
}

// RenderToHTMLImage renders D2 code to an HTML img tag
func (r *Renderer) RenderToHTMLImage(d2Code string, alt string) (string, error) {
	dataURL, err := r.RenderToDataURL(d2Code)
	if err != nil {
		return "", err
	}

	// Create HTML img tag with the data URL
	html := fmt.Sprintf(`<img src="%s" alt="%s" class="d2-diagram" />`, dataURL, escapeHTML(alt))
	return html, nil
}

// escapeHTML escapes HTML special characters
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}