package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Layout grid. Everything sits on a 4 px grid with two vertical lines down the
// window: boxed things (search box, list selection, buttons) start on the box
// line, and all text starts on the text line. Widgets inset their text by the
// theme's inner padding (4), so placing any widget on the box line puts its text
// on the text line. Windows asks for 16 px between a surface edge and its text,
// 12 px between content areas and 8 px between buttons.
const (
	boxLine     float32 = 12 // window edge to boxed controls, both sides
	textLine    float32 = 16 // window edge to text: boxLine + inner padding
	gapArea     float32 = 12 // between content areas
	gapItem     float32 = 8  // between buttons; between a header and its content
	stripHeight float32 = 32 // header and key strips
)

// inset places o inside fixed margins (top, right, bottom, left).
func inset(o fyne.CanvasObject, top, right, bottom, left float32) fyne.CanvasObject {
	return container.New(layout.NewCustomPaddedLayout(top, bottom, left, right), o)
}

// stripLayout draws a full-width strip of fixed height: objects[0] is the
// background and fills it; objects[1] sits on the box line, centered vertically.
type stripLayout struct{}

func (stripLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(objs[1].MinSize().Width+2*boxLine, stripHeight)
}

func (stripLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	objs[0].Resize(size)
	objs[0].Move(fyne.NewPos(0, 0))
	h := objs[1].MinSize().Height
	objs[1].Resize(fyne.NewSize(size.Width-2*boxLine, h))
	objs[1].Move(fyne.NewPos(boxLine, float32(int((size.Height-h)/2))))
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

// strip lays content on a raised iron surface, like the header and key strips.
func strip(content fyne.CanvasObject) fyne.CanvasObject {
	return container.New(stripLayout{}, canvas.NewRectangle(ironSurface), content)
}

// edgeRule is the one-pixel metal edge between regions.
func edgeRule() fyne.CanvasObject {
	r := canvas.NewRectangle(ironEdge)
	r.SetMinSize(fyne.NewSize(1, 1))
	return r
}

// microText returns a small run of text in a theme color.
func microText(text string, color fyne.ThemeColorName, bold bool) *widget.TextSegment {
	return &widget.TextSegment{
		Text: text,
		Style: widget.RichTextStyle{
			ColorName: color,
			SizeName:  theme.SizeNameCaptionText,
			TextStyle: fyne.TextStyle{Bold: bold},
			Inline:    true,
		},
	}
}

// keyHints is the key strip: each key in red, its action dim.
func keyHints() *widget.RichText {
	var segs []widget.RichTextSegment
	for i, h := range keyHintList {
		if i > 0 {
			segs = append(segs, microText("  ", theme.ColorNamePlaceHolder, false))
		}
		segs = append(segs,
			microText(h.key, theme.ColorNameError, true),
			microText(" "+h.action, theme.ColorNamePlaceHolder, false))
	}
	return widget.NewRichText(segs...)
}

var keyHintList = []struct{ key, action string }{
	{"Enter", "copy"},
	{"Ctrl+N", "new"},
	{"Ctrl+E", "edit"},
	{"Ctrl+Shift+D", "delete"},
}

// rowRefs are the parts of one list row that updateRow fills in.
type rowRefs struct {
	bar     *canvas.Rectangle
	title   *widget.Label
	excerpt *widget.TextSegment
	rich    *widget.RichText
}

// newRow builds a two-line row: a bold title over a dim one-line excerpt, and a
// blood-red bar that shows when selected. The bar is overlaid on the row's left
// edge rather than placed beside the text, so row text stays on the text line.
func (m *MainWindow) newRow() fyne.CanvasObject {
	refs := &rowRefs{bar: canvas.NewRectangle(color.Transparent)}
	refs.bar.SetMinSize(fyne.NewSize(3, 1))
	refs.title = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	refs.title.Truncation = fyne.TextTruncateEllipsis
	refs.excerpt = microText("", theme.ColorNamePlaceHolder, false)
	refs.rich = widget.NewRichText(refs.excerpt)
	refs.rich.Truncation = fyne.TextTruncateEllipsis
	text := container.New(layout.NewCustomPaddedVBoxLayout(0), refs.title, refs.rich)
	row := container.NewThemeOverride(
		container.NewStack(text, container.NewBorder(nil, nil, refs.bar, nil)), denseTheme{NewTheme()})
	m.rows[row] = refs
	return row
}
