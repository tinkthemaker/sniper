package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Dialog grid. The editor, the recovery window and Fyne's dialogs sit on a 4 px
// grid with two vertical lines: boxed things (fields, buttons) start on the box
// line, and all text starts on the text line. Widgets inset their text by the
// theme's inner padding (4), so placing any widget on the box line puts its text
// on the text line. The main window has its own two lines (lcars.go).
const (
	boxLine  float32 = 12 // window edge to boxed controls, both sides
	textLine float32 = 16 // window edge to text: boxLine + inner padding
	gapArea  float32 = 12 // between content areas
	gapItem  float32 = 8  // between buttons; between a header and its content
)

// inset places o inside fixed margins (top, right, bottom, left).
func inset(o fyne.CanvasObject, top, right, bottom, left float32) fyne.CanvasObject {
	return container.New(layout.NewCustomPaddedLayout(top, bottom, left, right), o)
}

// Fyne v2.8 places custom dialog content padWidth/2 in from each side and
// padHeight below its heading (dialog/base.go), and draws the heading flush
// with the dialog's left edge. These are its constants.
const (
	fyneDialogPadWidth  float32 = 32
	fyneDialogPadHeight float32 = 16
)

// onDialogGrid puts custom dialog content on the same grid as the window,
// measured from the dialog's own edges: boxes 12 px in, text 16 px in. It
// cancels Fyne's fixed dialog padding and empty heading row, so give the
// dialog an empty title and draw the heading inside content.
func onDialogGrid(content fyne.CanvasObject) fyne.CanvasObject {
	headingRow := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}).MinSize().Height
	side := boxLine - fyneDialogPadWidth/2
	top := boxLine - headingRow - fyneDialogPadHeight
	bottom := boxLine - fyneDialogPadHeight - theme.Padding() // content ends a padding above the (empty) button row
	return inset(content, top, side, bottom, side)
}

// equalColumns lays its objects side by side in equal widths, gap apart.
// (Fyne's grid layout spaces by the theme padding, which is 4 here.)
type equalColumns struct{ gap float32 }

func (e *equalColumns) MinSize(objs []fyne.CanvasObject) fyne.Size {
	var w, h float32
	for _, o := range objs {
		w, h = max(w, o.MinSize().Width), max(h, o.MinSize().Height)
	}
	n := float32(len(objs))
	return fyne.NewSize(w*n+e.gap*(n-1), h)
}

func (e *equalColumns) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	n := float32(len(objs))
	w := (size.Width - e.gap*(n-1)) / n
	for i, o := range objs {
		o.Resize(fyne.NewSize(w, size.Height))
		o.Move(fyne.NewPos(float32(i)*(w+e.gap), 0))
	}
}

// Rounded text fields (see fieldTheme): the corner radius, and how far text and
// the caret sit from the field's edge, measured to the start of the text.
const (
	fieldRadius float32 = 14
	fieldInset  float32 = 8
	bodyPad     float32 = 8 // the multi-line field's inner padding (fieldTheme's trick needs a one-line field)
)

// blankIcon is a transparent 1×1 image. An entry with an icon insets its text
// for it, which is how fields get a wider left inset without growing taller.
var blankIcon = func() fyne.Resource {
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewNRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		panic(err)
	}
	return fyne.NewStaticResource("blank.png", b.Bytes())
}()

// insetField gives e a transparent icon on the left and an equally wide spacer
// on the right, so its text clears the rounded corners (fieldTheme sizes both).
func insetField(e *widget.Entry) {
	e.Icon = blankIcon
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(fieldInset, 1))
	e.ActionItem = spacer
}
