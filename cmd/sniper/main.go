// Command sniper is a compact snippet search-and-copy utility.
package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/tinkthemaker/sniper/internal/clipboard"
	"github.com/tinkthemaker/sniper/internal/instance"
	"github.com/tinkthemaker/sniper/internal/store"
	"github.com/tinkthemaker/sniper/internal/ui"
)

//go:embed Icon.png
var iconPNG []byte

func main() {
	a := app.NewWithID("app.sniper.desktop")
	a.SetIcon(fyne.NewStaticResource("Icon.png", iconPNG))

	if lock := start(a); lock != nil {
		defer func() { _ = lock.Release() }()
	}
	a.Run()
}

// start shows the first window and returns the instance lock to hold while the
// app runs, or nil if a fatal or already-running window was shown instead.
func start(a fyne.App) *instance.Lock {
	dir, err := store.DefaultDir()
	if err == nil {
		err = os.MkdirAll(dir, 0o700)
	}
	if err != nil {
		ui.ShowFatal(a, "Sniper cannot create its data folder.", fmt.Sprintf("%s\n%v", dir, err))
		return nil
	}
	lock, err := instance.Acquire(dir)
	if errors.Is(err, instance.ErrAlreadyRunning) {
		ui.ShowAlreadyRunning(a)
		return nil
	}
	if err != nil {
		ui.ShowFatal(a, "Sniper cannot lock its data folder.", fmt.Sprintf("%s\n%v", dir, err))
		return nil
	}

	st := store.New(dir)
	res := st.Load()
	deps := ui.Deps{
		App:        a,
		Store:      st,
		Clipboard:  clipboard.New(a.Clipboard()),
		OpenFolder: ui.DefaultOpenFolder,
		Now:        time.Now,
	}
	if res.State == store.StateDamaged {
		ui.ShowRecovery(deps, res)
	} else {
		ui.ShowMain(deps, res.Library)
	}
	return lock
}
