package d2renderer

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// d2Block is a custom AST node for D2 diagrams
type d2Block struct {
	ast.BaseBlock
	HTML []byte
}

// Kind returns the kind of this node
func (n *d2Block) Kind() ast.NodeKind {
	return ast.KindHTMLBlock
}

// Dump implements ast.Node
func (n *d2Block) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// D2 is a goldmark extension for rendering D2 diagrams
type D2 struct {
	renderer *Renderer
}

// NewD2Extension creates a new D2 extension for goldmark
func NewD2Extension() (goldmark.Extender, error) {
	renderer, err := NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to create D2 renderer: %w", err)
	}

	return &D2{
		renderer: renderer,
	}, nil
}

// Extend implements goldmark.Extender
func (d *D2) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&d2Transformer{renderer: d.renderer}, 100),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&d2BlockRenderer{}, 100),
		),
	)
}

// d2Transformer transforms D2 code blocks into D2 blocks
type d2Transformer struct {
	renderer *Renderer
}

// Transform implements parser.ASTTransformer
func (t *d2Transformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		// Check if this is a code block
		codeBlock, ok := n.(*ast.FencedCodeBlock)
		if !ok {
			return ast.WalkContinue, nil
		}

		// Check if it's a D2 code block
		language := codeBlock.Language(reader.Source())
		if string(language) != "d2" {
			return ast.WalkContinue, nil
		}

		// Extract the D2 code
		var d2Code bytes.Buffer
		for i := 0; i < codeBlock.Lines().Len(); i++ {
			line := codeBlock.Lines().At(i)
			d2Code.Write(line.Value(reader.Source()))
		}

		// Render the D2 diagram to HTML
		htmlImg, err := t.renderer.RenderToHTMLImage(d2Code.String(), "D2 Diagram")
		if err != nil {
			// If rendering fails, replace with error message
			errorHTML := fmt.Sprintf(`<div class="d2-error">Error rendering D2 diagram: %s</div>`, err.Error())
			htmlImg = errorHTML
		}

		// Create a new D2 block
		d2BlockNode := &d2Block{
			HTML: []byte(htmlImg),
		}

		// Replace the code block with the D2 block
		parent := n.Parent()
		if parent != nil {
			parent.ReplaceChild(parent, n, d2BlockNode)
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		// Handle error if needed
		_ = err
	}
}

// d2BlockRenderer renders D2 blocks
type d2BlockRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer
func (r *d2BlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHTMLBlock, r.renderD2Block)
}

func (r *d2BlockRenderer) renderD2Block(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	// Check if this is a D2 block
	if d2, ok := node.(*d2Block); ok {
		_, _ = w.Write(d2.HTML)
		return ast.WalkSkipChildren, nil
	}

	// Handle regular HTML blocks
	if n, ok := node.(*ast.HTMLBlock); ok {
		l := n.Lines().Len()
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			_, _ = w.Write(line.Value(source))
		}
		return ast.WalkContinue, nil
	}

	return ast.WalkContinue, nil
}