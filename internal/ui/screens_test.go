package ui

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2"

	"github.com/tinkthemaker/sniper/internal/snippets"
	"github.com/tinkthemaker/sniper/internal/store"
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
	t.Run("editor with validation error", func(t *testing.T) {
		h := newHarness(t, sampleLibrary())
		h.m.win.Resize(fyne.NewSize(520, 600))
		h.shortcut(shortcutNew)
		h.typeText("Draft title")
		h.shortcut(shortcutSave)
		shot("04-editor-error", h.m.win)
	})
	t.Run("delete confirmation", func(t *testing.T) {
		h := newHarness(t, sampleLibrary())
		h.m.win.Resize(fyne.NewSize(520, 600))
		h.shortcut(shortcutDelete)
		shot("05-delete-confirm", h.m.win)
	})
	t.Run("recovery", func(t *testing.T) {
		bak := goodBackup(t)
		rh := newRecoveryHarness(t, ptr("not json"), &bak)
		rh.r.win.Resize(fyne.NewSize(520, 320))
		shot("06-recovery", rh.r.win)
	})
	t.Run("recovery without backup", func(t *testing.T) {
		rh := newRecoveryHarness(t, ptr("not json"), nil)
		rh.r.win.Resize(fyne.NewSize(520, 320))
		shot("07-recovery-no-backup", rh.r.win)
	})
	t.Run("about", func(t *testing.T) {
		h := newHarness(t, sampleLibrary())
		h.m.win.Resize(fyne.NewSize(520, 600))
		h.m.showAbout()
		shot("08-about", h.m.win)
	})
	t.Run("import confirm", func(t *testing.T) {
		h := newHarness(t, sampleLibrary())
		h.m.win.Resize(fyne.NewSize(520, 600))
		h.m.importFrom(strings.NewReader(importJSON), "sample-import.json")
		shot("09-import-confirm", h.m.win)
	})
	t.Run("corpus", func(t *testing.T) {
		f, err := os.Open(filepath.Join("..", "..", "testdata", "sample-import.json"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		items, err := store.DecodeImport(f)
		if err != nil {
			t.Fatal(err)
		}
		lib, err := snippets.Append(snippets.NewLibrary(), items)
		if err != nil {
			t.Fatal(err)
		}
		h := newHarness(t, lib)
		h.m.win.Resize(fyne.NewSize(520, 600))
		for i := 0; i < 4; i++ { // select the emoji snippet
			h.key(fyne.KeyDown)
		}
		shot("10-corpus-emoji", h.m.win)
		h.key(fyne.KeyDown)
		h.key(fyne.KeyDown)
		shot("11-corpus-cjk", h.m.win)
		h.key(fyne.KeyDown)
		shot("12-corpus-rtl", h.m.win)
	})
}
