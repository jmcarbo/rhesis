package styles

import (
	_ "embed"
	"fmt"
	"os"
)

// Embedded default styles
//
//go:embed themes/modern.css
var ModernTheme string

//go:embed themes/minimal.css
var MinimalTheme string

//go:embed themes/dark.css
var DarkTheme string

//go:embed themes/elegant.css
var ElegantTheme string

// StyleManager handles presentation styles
type StyleManager struct {
	themes map[string]string
}

// NewStyleManager creates a new style manager with embedded themes
func NewStyleManager() *StyleManager {
	sm := &StyleManager{
		themes: make(map[string]string),
	}

	// Register embedded themes
	sm.themes["modern"] = ModernTheme
	sm.themes["minimal"] = MinimalTheme
	sm.themes["dark"] = DarkTheme
	sm.themes["elegant"] = ElegantTheme

	return sm
}

// GetStyle returns the CSS for the specified theme
func (sm *StyleManager) GetStyle(themeName string) (string, error) {
	// Check if it's a built-in theme
	if css, ok := sm.themes[themeName]; ok {
		return css, nil
	}

	// Check if it's a file path
	if _, err := os.Stat(themeName); err == nil {
		content, err := os.ReadFile(themeName)
		if err != nil {
			return "", fmt.Errorf("failed to read style file: %w", err)
		}
		return string(content), nil
	}

	// Default to modern theme
	return sm.themes["modern"], nil
}

// GetAvailableThemes returns a list of available built-in themes
func (sm *StyleManager) GetAvailableThemes() []string {
	themes := make([]string, 0, len(sm.themes))
	for name := range sm.themes {
		themes = append(themes, name)
	}
	return themes
}

// LoadCustomTheme loads a custom theme from a file
func (sm *StyleManager) LoadCustomTheme(name string, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read custom theme: %w", err)
	}

	sm.themes[name] = string(content)
	return nil
}
