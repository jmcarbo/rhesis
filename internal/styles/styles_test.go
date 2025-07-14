package styles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewStyleManager(t *testing.T) {
	sm := NewStyleManager()
	if sm == nil {
		t.Error("Expected StyleManager to be created")
	}
	
	// Check that default themes are loaded
	themes := sm.GetAvailableThemes()
	expectedThemes := []string{"modern", "minimal", "dark", "elegant"}
	
	if len(themes) != len(expectedThemes) {
		t.Errorf("Expected %d themes, got %d", len(expectedThemes), len(themes))
	}
}

func TestGetStyle(t *testing.T) {
	sm := NewStyleManager()
	
	tests := []struct {
		name      string
		themeName string
		shouldErr bool
		contains  string
	}{
		{
			name:      "modern theme",
			themeName: "modern",
			shouldErr: false,
			contains:  "Modern Theme",
		},
		{
			name:      "minimal theme",
			themeName: "minimal",
			shouldErr: false,
			contains:  "Minimal Theme",
		},
		{
			name:      "dark theme",
			themeName: "dark",
			shouldErr: false,
			contains:  "Dark Theme",
		},
		{
			name:      "elegant theme",
			themeName: "elegant",
			shouldErr: false,
			contains:  "Elegant Theme",
		},
		{
			name:      "non-existent theme defaults to modern",
			themeName: "nonexistent",
			shouldErr: false,
			contains:  "Modern Theme",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			css, err := sm.GetStyle(tt.themeName)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !strings.Contains(css, tt.contains) {
				t.Errorf("Expected CSS to contain '%s'", tt.contains)
			}
		})
	}
}

func TestGetStyleFromFile(t *testing.T) {
	// Create a temporary CSS file
	tmpDir := t.TempDir()
	cssFile := filepath.Join(tmpDir, "custom.css")
	customCSS := `/* Custom Theme */
body {
    background: #custom;
}`
	
	if err := os.WriteFile(cssFile, []byte(customCSS), 0644); err != nil {
		t.Fatalf("Failed to create test CSS file: %v", err)
	}
	
	sm := NewStyleManager()
	css, err := sm.GetStyle(cssFile)
	if err != nil {
		t.Errorf("Failed to get style from file: %v", err)
	}
	
	if css != customCSS {
		t.Error("Expected custom CSS content")
	}
}

func TestLoadCustomTheme(t *testing.T) {
	// Create a temporary CSS file
	tmpDir := t.TempDir()
	cssFile := filepath.Join(tmpDir, "custom.css")
	customCSS := `/* My Custom Theme */`
	
	if err := os.WriteFile(cssFile, []byte(customCSS), 0644); err != nil {
		t.Fatalf("Failed to create test CSS file: %v", err)
	}
	
	sm := NewStyleManager()
	
	// Load the custom theme
	if err := sm.LoadCustomTheme("mycustom", cssFile); err != nil {
		t.Errorf("Failed to load custom theme: %v", err)
	}
	
	// Verify it was loaded
	css, err := sm.GetStyle("mycustom")
	if err != nil {
		t.Errorf("Failed to get loaded custom theme: %v", err)
	}
	
	if css != customCSS {
		t.Error("Expected custom theme content to match")
	}
	
	// Check it appears in available themes
	themes := sm.GetAvailableThemes()
	found := false
	for _, theme := range themes {
		if theme == "mycustom" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'mycustom' to be in available themes")
	}
}