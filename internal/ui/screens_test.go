package ui

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"

	"sniper/internal/snippets"
)

// TestScreens renders key states to PNGs for human review. It only runs when
// SNIPER_SCREENS=1, and the output is not asserted on because font rendering
// differs between machines.
func TestScreens(t *testing.T) {
	if os.Getenv("SNIPER_SCREENS") != "1" {
		t.Skip("set SNIPER_SCREENS=1 to render screenshots")
	}
	dir := filepath.Join("testdata", "screens")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	shot := func(name string, w fyne.Window) {
		t.Helper()
		img := w.Canvas().Capture()
		f, err := os.Create(filepath.Join(dir, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("empty library", func(t *testing.T) {
		h := newHarness(t, snippets.NewLibrary())
		h.m.win.Resize(fyne.NewSize(520, 600))
		shot("01-empty-library", h.m.win)
	})
	t.Run("populated", func(t *testing.T) {
		h := newHarness(t, sampleLibrary())
		h.m.win.Resize(fyne.NewSize(520, 600))
		h.key(fyne.KeyDown)
		h.key(fyne.KeyReturn)
		shot("02-populated", h.m.win)
	})
	t.Run("no matches", func(t *testing.T) {
		h := newHarness(t, sampleLibrary())
		h.m.win.Resize(fyne.NewSize(520, 600))
		h.typeText("nomatch")
		shot("03-no-matches", h.m.win)
	})
}
