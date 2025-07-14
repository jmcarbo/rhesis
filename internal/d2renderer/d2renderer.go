package d2renderer

import (
	"context"
	"encoding/base64"
	"fmt"

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
	ruler *textmeasure.Ruler
}

// New creates a new D2 renderer
func New() (*Renderer, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, fmt.Errorf("failed to create text ruler: %w", err)
	}
	
	return &Renderer{
		ruler: ruler,
	}, nil
}

// RenderToSVG converts D2 code to SVG
func (r *Renderer) RenderToSVG(d2Code string) ([]byte, error) {
	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return d2dagrelayout.DefaultLayout, nil
	}
	
	renderOpts := &d2svg.RenderOpts{
		Pad:     go2.Pointer(int64(10)),
		ThemeID: &d2themescatalog.GrapeSoda.ID,
	}
	
	compileOpts := &d2lib.CompileOptions{
		LayoutResolver: layoutResolver,
		Ruler:          r.ruler,
	}
	
	ctx := log.WithDefault(context.Background())
	diagram, _, err := d2lib.Compile(ctx, d2Code, compileOpts, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to compile D2 diagram: %w", err)
	}
	
	svg, err := d2svg.Render(diagram, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to render D2 diagram: %w", err)
	}
	
	return svg, nil
}

// RenderToBase64DataURI converts D2 code to a base64 data URI
func (r *Renderer) RenderToBase64DataURI(d2Code string) (string, error) {
	svg, err := r.RenderToSVG(d2Code)
	if err != nil {
		return "", err
	}
	
	encoded := base64.StdEncoding.EncodeToString(svg)
	return fmt.Sprintf("data:image/svg+xml;base64,%s", encoded), nil
}