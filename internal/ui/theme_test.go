package ui

import (
	"image/color"
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	lin := func(v uint32) float64 {
		x := float64(v) / 65535
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

func contrast(a, b color.Color) float64 {
	la, lb := luminance(a), luminance(b)
	return (math.Max(la, lb) + 0.05) / (math.Min(la, lb) + 0.05)
}

// blend draws over onto an opaque under, the way Fyne blends hover colors.
func blend(under, over color.Color) color.Color {
	ur, ug, ub, _ := under.RGBA()
	or, og, ob, oa := over.RGBA() // alpha-premultiplied
	mix := func(u, o uint32) uint8 { return uint8((o + u*(0xFFFF-oa)/0xFFFF) >> 8) }
	return color.NRGBA{R: mix(ur, or), G: mix(ug, og), B: mix(ub, ob), A: 0xFF}
}

// TestThemeContrastFloors keeps every text color readable on every surface it
// can land on. Blood red is deliberately absent: it never carries small text.
func TestThemeContrastFloors(t *testing.T) {
	surfaces := map[string]color.Color{
		"canvas": ironCanvas, "panel": ironPanel, "surface": ironSurface,
		"selection": ironSelect,
		// Hover is translucent, so test what it actually produces.
		"hovered row": blend(ironCanvas, ironHover),
	}
	text := map[string]color.Color{"bright": textBright, "mid": textMid, "dim": textDim, "red text": redText, "ember": ember}
	// Red text only appears on the flat surfaces (error labels, key chips), never
	// on the selection or hover fills.
	redSurfaces := map[string]bool{"canvas": true, "panel": true, "surface": true}
	for sn, s := range surfaces {
		for tn, c := range text {
			if tn == "red text" && !redSurfaces[sn] {
				continue
			}
			if got := contrast(c, s); got < 4.5 {
				t.Errorf("%s on %s = %.2f:1, want at least 4.5:1", tn, sn, got)
			}
		}
	}
	// Secondary button labels are always the bright foreground, at rest and hovered.
	for name, bg := range map[string]color.Color{"button": ironRaised, "hovered button": blend(ironRaised, ironHover)} {
		if got := contrast(textBright, bg); got < 4.5 {
			t.Errorf("bright on %s = %.2f:1, want at least 4.5:1", name, got)
		}
	}
	// Text on the two filled buttons.
	if got := contrast(ironCanvas, ember); got < 4.5 {
		t.Errorf("text on the ember button = %.2f:1", got)
	}
	if got := contrast(ironCanvas, redText); got < 4.5 {
		t.Errorf("text on the danger button = %.2f:1", got)
	}
}

func TestThemeIsAlwaysDark(t *testing.T) {
	th := NewTheme()
	for _, v := range []fyne.ThemeVariant{theme.VariantDark, theme.VariantLight} {
		if got := th.Color(theme.ColorNameBackground, v); got != ironCanvas {
			t.Errorf("variant %v background = %v, want iron canvas", v, got)
		}
	}
}

func TestThemeFontFallback(t *testing.T) {
	th := &slayerTheme{base: theme.DefaultTheme()} // no system fonts found
	for _, s := range []fyne.TextStyle{{}, {Bold: true}, {Monospace: true}, {Symbol: true}} {
		if th.Font(s) == nil {
			t.Errorf("no font for %+v when system fonts are missing", s)
		}
	}
	// Symbol text always uses Fyne's own font, even with Consolas loaded.
	withFonts := &slayerTheme{base: theme.DefaultTheme(), font: func(s fyne.TextStyle) fyne.Resource {
		if s.Symbol {
			return nil
		}
		return fyne.NewStaticResource("x.ttf", []byte{1})
	}}
	if withFonts.Font(fyne.TextStyle{Symbol: true}).Name() == "x.ttf" {
		t.Error("symbol font should come from the base theme")
	}
}
