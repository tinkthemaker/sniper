// Temporary M1 spike: proves Fyne builds, packages and copies. Replaced in M4.
package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"sniper/internal/clipboard"
)

//go:embed Icon.png
var iconPNG []byte

var items = []struct{ title, body string }{
	{"Greeting", "Hello,\nThanks for reaching out."},
	{"Emoji", "emoji \U0001F44B\U0001F3FD family \U0001F468‍\U0001F469‍\U0001F467"},
	{"Japanese", "日本語のテキスト"},
}

func main() {
	a := app.NewWithID("app.sniper.desktop")
	a.SetIcon(fyne.NewStaticResource("Icon.png", iconPNG))
	w := a.NewWindow("Sniper")
	cb := clipboard.New(a.Clipboard())

	sel := 0
	preview := widget.NewLabel(items[0].body)
	preview.Selectable = true
	list := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject { return widget.NewLabel("title") },
		func(i widget.ListItemID, o fyne.CanvasObject) { o.(*widget.Label).SetText(items[i].title) },
	)
	list.OnSelected = func(i widget.ListItemID) { sel = i; preview.SetText(items[i].body) }
	list.Select(0)
	status := widget.NewLabel("")
	copyBtn := widget.NewButton("Copy", func() {
		cb.SetText(items[sel].body)
		status.SetText("Copied “" + items[sel].title + "”.")
	})
	w.SetContent(container.NewBorder(nil, container.NewVBox(copyBtn, status), nil, nil,
		container.NewVSplit(list, container.NewVScroll(preview))))
	w.Resize(fyne.NewSize(520, 600))
	w.ShowAndRun()
}
