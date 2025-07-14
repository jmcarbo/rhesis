package generator

import (
	"bytes"
	"fmt"

	"github.com/jmcarbo/rhesis/internal/d2renderer"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// d2BlockKind is a NodeKind for D2 diagram blocks
var d2BlockKind = ast.NewNodeKind("D2Block")

// D2Block represents a D2 diagram block in the AST
type D2Block struct {
	ast.BaseBlock
	DataURI string
}

// Kind returns the NodeKind for D2Block
func (n *D2Block) Kind() ast.NodeKind {
	return d2BlockKind
}

// Dump dumps the D2Block node for debugging
func (n *D2Block) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"DataURI": n.DataURI,
	}, nil)
}

// NewD2Block creates a new D2Block node
func NewD2Block(dataURI string) *D2Block {
	return &D2Block{
		BaseBlock: ast.BaseBlock{},
		DataURI:   dataURI,
	}
}

// D2Extension is a Goldmark extension for rendering D2 diagrams
type D2Extension struct {
	renderer *d2renderer.Renderer
}

// NewD2Extension creates a new D2 extension
func NewD2Extension() (*D2Extension, error) {
	r, err := d2renderer.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create D2 renderer: %w", err)
	}
	
	return &D2Extension{
		renderer: r,
	}, nil
}

// Extend implements goldmark.Extender
func (e *D2Extension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&d2Transformer{renderer: e.renderer}, 100),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&d2HTMLRenderer{}, 100),
	))
}

// d2Transformer transforms D2 code blocks into D2Block nodes
type d2Transformer struct {
	renderer *d2renderer.Renderer
}

// Transform implements parser.ASTTransformer
func (t *d2Transformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		// Check if this is a code block
		codeBlock, ok := n.(*ast.FencedCodeBlock)
		if !ok {
			return ast.WalkContinue, nil
		}

		// Check if the language is d2
		lang := codeBlock.Language(reader.Source())
		if string(lang) != "d2" {
			return ast.WalkContinue, nil
		}

		// Extract the D2 code
		var buf bytes.Buffer
		for i := 0; i < codeBlock.Lines().Len(); i++ {
			line := codeBlock.Lines().At(i)
			buf.Write(line.Value(reader.Source()))
		}
		d2Code := buf.String()

		// Render the D2 code to SVG
		dataURI, err := t.renderer.RenderToBase64DataURI(d2Code)
		if err != nil {
			// If rendering fails, keep the code block as is
			return ast.WalkContinue, nil
		}

		// Create a new D2Block node
		d2Block := NewD2Block(dataURI)
		
		// Replace the code block with the D2Block
		parent := n.Parent()
		parent.ReplaceChild(parent, n, d2Block)

		return ast.WalkContinue, nil
	})
}

// d2HTMLRenderer is a custom renderer for D2Block nodes
type d2HTMLRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer
func (r *d2HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(d2BlockKind, r.renderD2Block)
}

func (r *d2HTMLRenderer) renderD2Block(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	
	block := n.(*D2Block)
	fmt.Fprintf(w, `<div class="d2-diagram"><img src="%s" alt="D2 Diagram"></div>`, block.DataURI)
	
	return ast.WalkContinue, nil
}