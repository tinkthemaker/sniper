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
// can land on: 4.5:1 or better.
func TestThemeContrastFloors(t *testing.T) {
	surfaces := map[string]color.Color{
		"canvas": colCanvas, "dialog": colSurface, "raised button": colRaised,
		"selected text": colSelect,
		// Hover is translucent, so test what it actually produces.
		"hovered canvas":        blend(colCanvas, colHover),
		"hovered raised button": blend(colRaised, colHover),
	}
	text := map[string]color.Color{
		"ink": colInk, "tan": colTan, "dim": colDim, "orange": colOrange,
		"lavender": colLavender, "periwinkle": colPeri, "red": colRed, "gold": colGold,
	}
	// Colored text is for the canvas and dialogs; on a selection or a button it
	// is only the warm white and tan that appear.
	flat := map[string]bool{"canvas": true, "dialog": true, "hovered canvas": true}
	always := map[string]bool{"ink": true, "tan": true}
	for sn, s := range surfaces {
		for tn, c := range text {
			if !flat[sn] && !always[tn] {
				continue
			}
			if tn == "dim" && sn != "canvas" && sn != "dialog" {
				continue // placeholders sit on inputs, which are the canvas
			}
			if got := contrast(c, s); got < 4.5 {
				t.Errorf("%s on %s = %.2f:1, want at least 4.5:1", tn, sn, got)
			}
		}
	}
	// Black text on every block color, and on those colors under the hover and
	// focus veils.
	blocks := map[string]color.Color{
		"orange": colOrange, "peach": colPeach, "gold": colGold, "tan": colTan,
		"lavender": colLavender, "periwinkle": colPeri, "red": colRed,
	}
	for n, bg := range blocks {
		for veil, v := range map[string]color.Color{"": color.Transparent, " hovered": colHover, " focused": colFocus} {
			if got := contrast(colCanvas, blend(bg, v)); got < 4.5 {
				t.Errorf("black on the%s %s block = %.2f:1, want at least 4.5:1", veil, n, got)
			}
		}
	}
}

func TestThemeIsAlwaysDark(t *testing.T) {
	th := NewTheme()
	for _, v := range []fyne.ThemeVariant{theme.VariantDark, theme.VariantLight} {
		if got := th.Color(theme.ColorNameBackground, v); got != colCanvas {
			t.Errorf("variant %v background = %v, want the black canvas", v, got)
		}
	}
}

func TestThemeMapsBoldToAntonioAndTextToJost(t *testing.T) {
	th := NewTheme()
	for style, want := range map[fyne.TextStyle]string{
		{}:                         "Jost-500-Medium.ttf",
		{Italic: true}:             "Jost-500-MediumItalic.ttf",
		{Bold: true}:               "Antonio-SemiBold.ttf",
		{Bold: true, Italic: true}: "Jost-600-SemiItalic.ttf",
	} {
		if got := th.Font(style).Name(); got != want {
			t.Errorf("font for %+v = %s, want %s", style, got, want)
		}
	}
	// Symbol and monospace text keep Fyne's own fonts.
	base := theme.DefaultTheme()
	for _, s := range []fyne.TextStyle{{Symbol: true}, {Monospace: true}} {
		if got, want := th.Font(s).Name(), base.Font(s).Name(); got != want {
			t.Errorf("font for %+v = %s, want the base theme's %s", s, got, want)
		}
	}
	// The embedded files are real TrueType fonts (sfnt version 0x00010000).
	for _, r := range []fyne.Resource{fontRegular, fontItalic, fontBold, fontBoldItalic} {
		if c := r.Content(); len(c) < 1024 || string(c[:4]) != "\x00\x01\x00\x00" {
			t.Errorf("%s is not a TrueType font (%d bytes)", r.Name(), len(c))
		}
	}
}

func TestThemeButtonsArePills(t *testing.T) {
	th := NewTheme()
	if got := th.Size(theme.SizeNameButtonRadius); got < 30 {
		t.Errorf("button radius = %v, want one large enough to round a button into a pill", got)
	}
	if got := th.Color(theme.ColorNameForegroundOnPrimary, theme.VariantDark); got != colCanvas {
		t.Errorf("text on an orange button = %v, want black", got)
	}
}
