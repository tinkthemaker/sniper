package ui

import (
	"fmt"
	"hash/fnv"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// doubleClickWindow is how close two clicks on one row must be to count as a
// double-click. It is Windows' default double-click time.
const doubleClickWindow = 500 * time.Millisecond

// codeWidth is the room a row gives its record code, from the text line to the
// start of the title.
const codeWidth float32 = 62

// recordCode is the display code of a snippet, "NN-NNNN", derived only from its
// ID, so a snippet keeps its code however the list is filtered or ordered.
func recordCode(id string) string {
	h := fnv.New32a()
	h.Write([]byte(id))
	n := h.Sum32()
	return fmt.Sprintf("%02d-%04d", n%100, n/100%10000)
}

// rowRefs are the parts of one list row that updateRow fills in.
type rowRefs struct {
	pill      *canvas.Rectangle // orange when selected
	code      *widget.TextSegment
	codeText  *widget.RichText
	title     *widget.TextSegment
	titleText *widget.RichText
	id        string          // the snippet this row currently shows
	fade      *fyne.Animation // the highlight fade in progress, if any
	gen       int             // counts fades, so a superseded one does nothing
}

// rowLayout draws a list row: the selection pill across the whole row, the
// record code on the text line, and the title after it, each centered
// vertically. objs are pill, code, title.
type rowLayout struct{}

func (rowLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(textInset+codeWidth, rowHeight)
}

func (rowLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	objs[0].Resize(size)
	objs[0].Move(fyne.NewPos(0, 0))
	for i, x := range []float32{textInset - textPad, textInset + codeWidth - textPad} {
		o := objs[i+1]
		h := o.MinSize().Height
		w := size.Width - x - 4
		if i == 0 {
			w = codeWidth
		}
		o.Resize(fyne.NewSize(max(w, 0), h))
		o.Move(fyne.NewPos(x, round((size.Height-h)/2)-inkLift))
	}
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

// newRow builds a one-line row: a record code, the title in capitals, and an
// orange pill behind both when selected. The pill is overlaid on the row
// rather than placed beside the text, so row text stays on the text line.
func (m *MainWindow) newRow() fyne.CanvasObject {
	refs := &rowRefs{pill: canvas.NewRectangle(color.Transparent)}
	refs.pill.CornerRadius = rowHeight / 2
	refs.code = capsSeg("", namePeri, sizeCode)
	refs.codeText = widget.NewRichText(refs.code)
	refs.title = capsSeg("", nameTan, sizeRow)
	refs.titleText = widget.NewRichText(refs.title)
	refs.titleText.Truncation = fyne.TextTruncateEllipsis
	content := container.New(rowLayout{}, refs.pill, refs.codeText, refs.titleText)
	row := &listRow{m: m, refs: refs, content: content}
	row.ExtendBaseWidget(row)
	m.rows[row] = refs
	return row
}

// show fills the row for the snippet id with title, selected or not. Capitals
// are for display only; the stored title is never changed.
func (refs *rowRefs) show(id, title string, selected bool) {
	refs.id = id
	refs.code.Text = recordCode(id)
	refs.title.Text = strings.ToUpper(title)
	if selected {
		refs.paint(1)
	} else {
		refs.paint(0)
	}
}

// paint draws the row's highlight at fraction f: 0 is unselected, 1 is the full
// orange pill with black text. Text turns black once the pill is half there, so
// it is never black on the dark canvas nor tan on bright orange for long.
func (refs *rowRefs) paint(f float32) {
	refs.pill.FillColor = fade(colOrange, f)
	if f >= 0.5 {
		refs.code.Style.ColorName, refs.title.Style.ColorName = nameBlack, nameBlack
	} else {
		refs.code.Style.ColorName, refs.title.Style.ColorName = namePeri, nameTan
	}
	refs.pill.Refresh()
	refs.codeText.Refresh()
	refs.titleText.Refresh()
}
