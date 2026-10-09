package ui

import (
	"image/color"
	"time"

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

// rowHeight is the height of a list row, which is also the row pitch: the list
// adds no gap between rows (see listTheme). Fixed, so every row is the same.
const rowHeight float32 = 24

// doubleClickWindow is how close two clicks on one row must be to count as a
// double-click. It is Windows' default double-click time.
const doubleClickWindow = 500 * time.Millisecond

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

// Copy is the primary action, so it is bigger than the other buttons: at least
// copyMinWidth by copyMinHeight, with 16 px text.
const (
	copyMinWidth  float32 = 112
	copyMinHeight float32 = 40
	copyTextSize  float32 = 16
)

// buttonRowLayout lays out the action buttons side by side, gapItem apart. The
// first is the primary button and is at least copyMinWidth by copyMinHeight; the
// rest keep their own size. All are centered vertically in the row.
type buttonRowLayout struct{}

func (buttonRowLayout) sizeOf(i int, o fyne.CanvasObject) fyne.Size {
	s := o.MinSize()
	if i == 0 {
		s = fyne.NewSize(max(s.Width, copyMinWidth), max(s.Height, copyMinHeight))
	}
	return s
}

func (l buttonRowLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	var w, h float32
	for i, o := range objs {
		s := l.sizeOf(i, o)
		w, h = w+s.Width, max(h, s.Height)
	}
	return fyne.NewSize(w+gapItem*float32(len(objs)-1), h)
}

func (l buttonRowLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	x := float32(0)
	for i, o := range objs {
		s := l.sizeOf(i, o)
		o.Resize(s)
		o.Move(fyne.NewPos(x, float32(int((size.Height-s.Height)/2))))
		x += s.Width + gapItem
	}
}

// vCenterLayout gives its one object the full width and its own height, centered
// vertically.
type vCenterLayout struct{}

func (vCenterLayout) MinSize(objs []fyne.CanvasObject) fyne.Size { return objs[0].MinSize() }

func (vCenterLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	h := objs[0].MinSize().Height
	objs[0].Resize(fyne.NewSize(size.Width, h))
	objs[0].Move(fyne.NewPos(0, (size.Height-h)/2))
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
	bar   *canvas.Rectangle
	title *widget.Label
	id    string // the snippet this row currently shows
}

// rowLayout draws a list row: the selection bar down the left edge and the
// title on the text line, centered vertically. objs are bar, title.
type rowLayout struct{}

func (rowLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(objs[1].MinSize().Width, rowHeight)
}

func (rowLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	objs[0].Resize(fyne.NewSize(objs[0].MinSize().Width, size.Height))
	objs[0].Move(fyne.NewPos(0, 0))
	h := objs[1].MinSize().Height
	objs[1].Resize(fyne.NewSize(size.Width, h))
	objs[1].Move(fyne.NewPos(0, (size.Height-h)/2))
}

// listRow is one list row. It takes the click itself instead of leaving it to
// the list's own item, so a double-click can be told apart (see tapRow).
type listRow struct {
	widget.BaseWidget
	m       *MainWindow
	refs    *rowRefs
	content fyne.CanvasObject
}

func (r *listRow) CreateRenderer() fyne.WidgetRenderer { return widget.NewSimpleRenderer(r.content) }

func (r *listRow) Tapped(*fyne.PointEvent) { r.m.tapRow(r.refs.id) }

// newRow builds a one-line row: a bold title and a blood-red bar that shows when
// selected. The bar is overlaid on the row's left edge rather than placed beside
// the text, so row text stays on the text line.
func (m *MainWindow) newRow() fyne.CanvasObject {
	refs := &rowRefs{bar: canvas.NewRectangle(color.Transparent)}
	refs.bar.SetMinSize(fyne.NewSize(3, 1))
	refs.title = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	refs.title.Truncation = fyne.TextTruncateEllipsis
	content := container.New(rowLayout{}, refs.bar, refs.title)
	row := &listRow{m: m, refs: refs, content: content}
	row.ExtendBaseWidget(row)
	m.rows[row] = refs
	return row
}
