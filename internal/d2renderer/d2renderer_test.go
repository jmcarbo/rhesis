package d2renderer

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	renderer, err := New()
	if err != nil {
		t.Fatalf("Failed to create D2 renderer: %v", err)
	}
	if renderer == nil {
		t.Fatal("Expected non-nil renderer")
	}
	if renderer.ruler == nil {
		t.Fatal("Expected non-nil ruler")
	}
}

func TestRenderToSVG(t *testing.T) {
	renderer, err := New()
	if err != nil {
		t.Fatalf("Failed to create D2 renderer: %v", err)
	}

	tests := []struct {
		name    string
		d2Code  string
		wantErr bool
	}{
		{
			name:    "Simple diagram",
			d2Code:  "x -> y",
			wantErr: false,
		},
		{
			name:    "Multiple connections",
			d2Code:  "a -> b\nb -> c\nc -> d",
			wantErr: false,
		},
		{
			name:    "With labels",
			d2Code:  "server: Web Server\ndatabase: PostgreSQL\nserver -> database: HTTP Request",
			wantErr: false,
		},
		{
			name:    "Empty diagram",
			d2Code:  "",
			wantErr: false,
		},
		{
			name:    "Invalid syntax",
			d2Code:  "x -> -> y",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := renderer.RenderToSVG(tt.d2Code)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderToSVG() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(svg) == 0 {
					t.Error("Expected non-empty SVG output")
				}
				if !strings.Contains(string(svg), "<svg") {
					t.Error("Output should contain SVG tag")
				}
			}
		})
	}
}

func TestRenderToBase64DataURI(t *testing.T) {
	renderer, err := New()
	if err != nil {
		t.Fatalf("Failed to create D2 renderer: %v", err)
	}

	tests := []struct {
		name    string
		d2Code  string
		wantErr bool
	}{
		{
			name:    "Simple diagram",
			d2Code:  "start -> end",
			wantErr: false,
		},
		{
			name:    "Complex diagram",
			d2Code:  "user -> api: request\napi -> db: query\ndb -> api: result\napi -> user: response",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataURI, err := renderer.RenderToBase64DataURI(tt.d2Code)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderToBase64DataURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !strings.HasPrefix(dataURI, "data:image/svg+xml;base64,") {
					t.Errorf("Expected data URI to start with 'data:image/svg+xml;base64,', got %s", dataURI[:50])
				}
			}
		})
	}
}