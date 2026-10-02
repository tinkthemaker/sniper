package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ShowAlreadyRunning tells the user another instance holds the library, then quits.
func ShowAlreadyRunning(a fyne.App) {
	w := a.NewWindow(appTitle)
	w.SetMaster()
	label := widget.NewLabel(alreadyRunningText)
	label.Wrapping = fyne.TextWrapWord
	w.SetContent(container.NewBorder(nil, container.NewCenter(widget.NewButton(labelOK, a.Quit)), nil, nil, label))
	w.Resize(fyne.NewSize(380, 140))
	w.Show()
}

// ShowFatal shows an unrecoverable startup error with selectable detail text.
func ShowFatal(a fyne.App, title, detail string) {
	w := a.NewWindow(appTitle)
	w.SetMaster()
	heading := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	body := widget.NewLabel(detail)
	body.Selectable = true
	body.Wrapping = fyne.TextWrapWord
	w.SetContent(container.NewBorder(heading, widget.NewButton(labelQuit, a.Quit), nil, nil, container.NewVScroll(body)))
	w.Resize(fyne.NewSize(480, 220))
	w.Show()
}
