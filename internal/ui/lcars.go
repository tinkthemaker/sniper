package ui

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// The main window is an LCARS-style frame (docs/design/lcars-mockup-v3.html):
// a top elbow with the SNIPER title, a column of four labelled blocks down the
// left, a bottom elbow that carries the status readout and two key hints, and a
// black panel in the middle with the search, list, preview and buttons. The
// frame is flat colored rectangles (Fyne 2.8 rectangles take a radius per
// corner), laid out by frameLayout from the window size alone.
//
// Inside the panel two vertical lines run down the window. Pills and highlight
// bars (FIND, the selected row, the divider, the buttons) start on the pill
// line; all text starts textInset further in. Widgets inset their text by the
// theme's 4 px inner padding, so a text widget sits textInset-4 right of the
// pill line.
const (
	frameMargin float32 = 8  // window edge to the frame
	columnWidth float32 = 96 // the left column and the elbows' caps
	segGap      float32 = 4  // between blocks, segments and buttons

	pillLine   float32 = frameMargin + columnWidth + 16 // x where pills start
	textInset  float32 = 12                             // pill line to text
	panelRight float32 = 12                             // window edge to the panel's right edge
	textPad    float32 = 4                              // widgets inset their text by this much

	topCapHeight float32 = 56 // top elbow
	topBarHeight float32 = 24
	botCapHeight float32 = 40 // bottom elbow
	botBarHeight float32 = 22
	joinWidth    float32 = 24 // the elbow's neck, where it meets its bar
	endCapWidth  float32 = 28

	searchHeight float32 = 28
	rowHeight    float32 = 24 // a list row, which is also the row pitch
	dividerH     float32 = 10
	listShare    float32 = 0.55 // the list's share of the room between the search and the buttons

	pillPad float32 = 12 // between a button's label and its right end

	// Antonio's capitals sit about 2 px below the middle of their line box, so
	// text that is meant to be centered in a bar is lifted by this much.
	inkLift float32 = 2
)

// column blocks, top to bottom; the third takes the rest of the column.
var blockHeights = [4]float32{70, 46, 0, 40}
var blockColors = [4]color.Color{colPeri, colGold, colLavender, colRed}

// caps is a run of text in Antonio (bold) in a palette color and size.
func caps(text string, c fyne.ThemeColorName, size fyne.ThemeSizeName) *widget.RichText {
	return widget.NewRichText(capsSeg(text, c, size))
}

func capsSeg(text string, c fyne.ThemeColorName, size fyne.ThemeSizeName) *widget.TextSegment {
	return &widget.TextSegment{
		Text: text,
		Style: widget.RichTextStyle{
			ColorName: c, SizeName: size, TextStyle: fyne.TextStyle{Bold: true}, Inline: true,
		},
	}
}

func block(c color.Color) *canvas.Rectangle { return canvas.NewRectangle(c) }

// rounded returns a rectangle with its own radius at each corner.
func rounded(c color.Color, topLeft, topRight, bottomRight, bottomLeft float32) *canvas.Rectangle {
	r := canvas.NewRectangle(c)
	r.TopLeftCornerRadius, r.TopRightCornerRadius = topLeft, topRight
	r.BottomRightCornerRadius, r.BottomLeftCornerRadius = bottomRight, bottomLeft
	return r
}

func round(v float32) float32 { return float32(math.Round(float64(v))) }

// frame is the decoration around the panel, and the layout that places it.
type frame struct {
	topBar, topCap, topJoin, topCorner, topEnd *canvas.Rectangle
	brand                                      *widget.RichText

	blocks     [4]*canvas.Rectangle
	blockTexts [4]*widget.RichText

	botCap, botJoin, botCorner, botEnd *canvas.Rectangle
	readout, hintNew, hintUndo         *canvas.Rectangle
	status                             fyne.CanvasObject
	hintNewText, hintUndoText          *widget.RichText

	panel fyne.CanvasObject
}

// newFrame builds the frame around panel. texts are the four column labels.
func newFrame(texts [4]*widget.RichText, status fyne.CanvasObject, panel fyne.CanvasObject) (*frame, fyne.CanvasObject) {
	f := &frame{
		topBar:    block(colPeach),
		topCap:    rounded(colPeach, 44, 0, 0, 0),
		topJoin:   block(colPeach),
		topCorner: rounded(colCanvas, 22, 0, 0, 0), // the panel's inner corner, cut out of the join
		topEnd:    rounded(colPeach, 0, 12, 12, 0),
		brand:     caps(brandName, nameOrange, sizeBrand),

		blockTexts: texts,

		botCap:       rounded(colRed, 0, 0, 0, 34),
		botJoin:      block(colRed),
		botCorner:    rounded(colCanvas, 0, 0, 0, 18),
		botEnd:       rounded(colRed, 0, 11, 11, 0),
		readout:      block(colRed),
		hintNew:      block(colGold),
		hintUndo:     block(colLavender),
		status:       status,
		hintNewText:  caps(hintNewText, nameBlack, sizeHint),
		hintUndoText: caps(hintUndoText, nameBlack, sizeHint),
		panel:        panel,
	}
	objs := []fyne.CanvasObject{f.topBar, f.topCap, f.topJoin, f.topCorner, f.topEnd, f.brand}
	for i := range f.blocks {
		f.blocks[i] = block(blockColors[i])
		objs = append(objs, f.blocks[i])
	}
	for _, t := range texts {
		objs = append(objs, t)
	}
	objs = append(objs, f.botCap, f.botJoin, f.botCorner, f.readout, f.status,
		f.hintNew, f.hintNewText, f.hintUndo, f.hintUndoText, f.botEnd, f.panel)
	return f, container.New(f, objs...)
}

func (f *frame) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(420, 380) }

// panelRect is where the panel sits for a window of the given size.
func panelRect(size fyne.Size) (fyne.Position, fyne.Size) {
	top := frameMargin + topCapHeight + segGap + segGap
	bottom := frameMargin + botCapHeight + segGap + 6
	return fyne.NewPos(pillLine, top), fyne.NewSize(size.Width-pillLine-panelRight, size.Height-top-bottom)
}

func place(o fyne.CanvasObject, x, y, w, h float32) {
	o.Move(fyne.NewPos(x, y))
	o.Resize(fyne.NewSize(w, h))
}

func (f *frame) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	W, H, m, col := size.Width, size.Height, frameMargin, columnWidth

	// Top elbow: the peach cap curves into the bar across the top. The title
	// breaks the bar on the right; a short end cap closes it.
	brandMin := f.brand.MinSize()
	boxW := brandMin.Width + 12 // the title's box: its text plus 10 px each side
	barEnd := W - m - endCapWidth - boxW
	place(f.topCap, m, m, col, topCapHeight)
	place(f.topBar, m+col, m, barEnd-(m+col), topBarHeight)
	place(f.topJoin, m+col, m+topBarHeight, joinWidth, topCapHeight-topBarHeight)
	place(f.topCorner, m+col, m+topBarHeight+8, joinWidth, topCapHeight-topBarHeight-8)
	place(f.topEnd, W-m-endCapWidth, m, endCapWidth, topBarHeight)
	place(f.brand, barEnd+6, m+topBarHeight/2-4-brandMin.Height/2, brandMin.Width, brandMin.Height)

	// Left column: three fixed blocks and one that takes the rest, 4 px apart.
	colTop := m + topCapHeight + segGap
	colBottom := H - m - botCapHeight - segGap
	rest := colBottom - colTop - segGap*3
	for _, h := range blockHeights {
		rest -= h
	}
	y := colTop
	for i := range f.blocks {
		h := blockHeights[i]
		if h == 0 {
			h = max(rest, 0)
		}
		place(f.blocks[i], m, y, col, h)
		// The label sits bottom-right, 8 px from the right and 4 px up.
		t := f.blockTexts[i]
		th := t.MinSize().Height
		place(t, m, y+h-th, col-textPad, th)
		y += h + segGap
	}

	// Bottom elbow: the red cap curves into a bar that holds the status
	// readout, two key hints and an end cap.
	top := H - m - botCapHeight
	place(f.botCap, m, top, col, botCapHeight)
	place(f.botJoin, m+col, top, joinWidth, botCapHeight)
	place(f.botCorner, m+col, top, joinWidth, 18)
	barY := H - m - botBarHeight
	place(f.botEnd, W-m-endCapWidth, barY, endCapWidth, botBarHeight)
	right := W - m - endCapWidth - segGap
	for _, h := range []struct {
		bg   *canvas.Rectangle
		text *widget.RichText
	}{{f.hintUndo, f.hintUndoText}, {f.hintNew, f.hintNewText}} {
		tm := h.text.MinSize()
		w := tm.Width + 12
		place(h.bg, right-w, barY, w, botBarHeight)
		place(h.text, right-w+6, barY+(botBarHeight-tm.Height)/2-inkLift, tm.Width, tm.Height)
		right -= w + segGap
	}
	readoutX := m + col + joinWidth
	place(f.readout, readoutX, barY, right-readoutX, botBarHeight)
	sh := f.status.MinSize().Height
	place(f.status, readoutX+6, barY+(botBarHeight-sh)/2-inkLift, max(right-readoutX-12, 0), sh)

	pos, sz := panelRect(size)
	place(f.panel, pos.X, pos.Y, sz.Width, sz.Height)
}

// panelParts are the objects inside the panel, in tab order.
type panelParts struct {
	findPill    *canvas.Rectangle
	findText    *widget.RichText
	outline     *canvas.Rectangle
	search      fyne.CanvasObject
	listArea    fyne.CanvasObject
	divider     [5]*canvas.Rectangle
	previewHead *widget.RichText
	previewName fyne.CanvasObject
	previewBody fyne.CanvasObject
	buttons     fyne.CanvasObject
}

var dividerColors = [5]color.Color{colPeach, colGold, colLavender, colPeri, colPeach}

func newPanel(p *panelParts) fyne.CanvasObject {
	p.findPill = rounded(colOrange, searchHeight/2, 0, 0, searchHeight/2)
	p.findText = caps(findLabel, nameBlack, sizeCode)
	p.outline = rounded(color.Transparent, 0, searchHeight/2, searchHeight/2, 0)
	p.outline.StrokeColor, p.outline.StrokeWidth = colOrange, 2
	objs := []fyne.CanvasObject{p.findPill, p.findText, p.outline, p.search, p.listArea}
	for i := range p.divider {
		p.divider[i] = block(dividerColors[i])
		objs = append(objs, p.divider[i])
	}
	p.divider[0].TopLeftCornerRadius, p.divider[0].BottomLeftCornerRadius = dividerH/2, dividerH/2
	p.divider[4].TopRightCornerRadius, p.divider[4].BottomRightCornerRadius = dividerH/2, dividerH/2
	objs = append(objs, p.previewHead, p.previewName, p.previewBody, p.buttons)
	return container.New(p, objs...)
}

func (p *panelParts) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(200, 200) }

func (p *panelParts) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	W, H := size.Width, size.Height

	// Search: the FIND pill and its outlined field, 4 px apart and the same height.
	ft := p.findText.MinSize()
	findW := ft.Width + 16 // text inset on the left, and 12 px after it
	place(p.findPill, 0, 0, findW, searchHeight)
	place(p.findText, textInset-textPad, (searchHeight-ft.Height)/2-inkLift, ft.Width, ft.Height)
	fieldX := findW + segGap
	place(p.outline, fieldX, 0, W-fieldX, searchHeight)
	place(p.search, fieldX+textInset-textPad, -inkLift, W-fieldX-textInset-searchHeight/2+textPad, searchHeight)

	// Below the search: the list, then the divider, then the preview, with the
	// buttons at the bottom. The list takes listShare of what is left.
	const above, below, bar = 10, 10, 14 // search to list; buttons' top margin; list to divider
	listTop := searchHeight + above
	actionsH := copyHeight
	actionsTop := H - actionsH
	room := actionsTop - below - listTop - (bar + dividerH + below)
	listH := max(round(room*listShare), 0)
	place(p.listArea, 0, listTop, W, listH)

	divY := listTop + listH + bar
	fr := (W - segGap*4 - 34 - 22) / 4
	widths := [5]float32{2 * fr, 34, fr, fr, 22}
	x := float32(0)
	for i, w := range widths {
		place(p.divider[i], x, divY, w, dividerH)
		x += w + segGap
	}

	y := divY + dividerH + below
	hm := p.previewHead.MinSize()
	place(p.previewHead, textInset-textPad, y, W-textInset, hm.Height)
	y += hm.Height
	nm := p.previewName.MinSize()
	place(p.previewName, textInset-textPad, y, W-textInset, nm.Height)
	y += nm.Height
	bodyBottom := actionsTop - below
	place(p.previewBody, textInset-textPad, y, W-textInset+textPad, max(bodyBottom-y, 0))

	place(p.buttons, 0, actionsTop, W, actionsH)
}

// Button sizes. COPY is the primary action and is at least 112×40; it starts
// on the pill line.
const (
	copyWidth    float32 = 120
	copyHeight   float32 = 40
	buttonHeight float32 = 30
)

var buttonWidths = [4]float32{copyWidth, 58, 58, 70}

// buttonRowLayout lays out the four action pills side by side, segGap apart,
// all centered vertically on COPY.
type buttonRowLayout struct{}

func (buttonRowLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	var w float32
	for _, bw := range buttonWidths {
		w += bw + segGap
	}
	return fyne.NewSize(w-segGap, copyHeight)
}

func (buttonRowLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	x := float32(0)
	for i, o := range objs {
		h := buttonHeight
		if i == 0 {
			h = copyHeight
		}
		place(o, x, round((size.Height-h)/2), buttonWidths[i], h)
		x += buttonWidths[i] + segGap
	}
}

// pill is an action button on its color. The color is a rectangle behind the
// button, so each end can be rounded on its own; the button on top is flat, with
// a black label and a veil for hover and focus. A disabled pill drops to a dim
// version of its color, which stays clearly different from the enabled one.
type pill struct {
	btn  *widget.Button
	fill *canvas.Rectangle
	c    color.Color
	obj  fyne.CanvasObject
}

func newPill(label string, tapped func(), c color.Color, text float32, left, right float32) *pill {
	p := &pill{btn: widget.NewButton(label, tapped), c: c}
	p.btn.Alignment = widget.ButtonAlignTrailing
	p.fill = rounded(c, left, right, right, left)
	radius := max(left, right)
	p.obj = container.New(pillLayout{}, p.fill, container.NewThemeOverride(p.btn, pillTheme{NewTheme(), text, radius}))
	return p
}

// pillLayout fills the pill with its color, and lifts the button a little so its
// label looks centered (see inkLift).
type pillLayout struct{}

func (pillLayout) MinSize(objs []fyne.CanvasObject) fyne.Size { return objs[1].MinSize() }

func (pillLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	place(objs[0], 0, 0, size.Width, size.Height)
	place(objs[1], 0, -inkLift, size.Width, size.Height)
}

// dimmed is c at 28% over black, the fill of a disabled pill.
func dimmed(c color.Color) color.Color {
	r, g, b, _ := c.RGBA()
	d := func(v uint32) uint8 { return uint8(v * 28 / 100 >> 8) }
	return color.NRGBA{R: d(r), G: d(g), B: d(b), A: 0xFF}
}

func (p *pill) setEnabled(on bool) {
	if on {
		p.btn.Enable()
		p.fill.FillColor = p.c
	} else {
		p.btn.Disable()
		p.fill.FillColor = dimmed(p.c)
	}
	p.fill.Refresh()
}
