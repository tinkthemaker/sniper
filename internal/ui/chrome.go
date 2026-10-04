package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// strip lays content on a raised iron surface, like the header and key strips.
func strip(content fyne.CanvasObject) fyne.CanvasObject {
	return container.NewStack(canvas.NewRectangle(ironSurface), container.NewPadded(content))
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

// newRow builds a two-line row: a blood-red bar that shows when selected, a
// bold title, and a dim one-line excerpt.
func (m *MainWindow) newRow() fyne.CanvasObject {
	refs := &rowRefs{bar: canvas.NewRectangle(color.Transparent)}
	refs.bar.SetMinSize(fyne.NewSize(3, 1))
	refs.title = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	refs.title.Truncation = fyne.TextTruncateEllipsis
	refs.excerpt = microText("", theme.ColorNamePlaceHolder, false)
	refs.rich = widget.NewRichText(refs.excerpt)
	refs.rich.Truncation = fyne.TextTruncateEllipsis
	body := container.NewVBox(refs.title, refs.rich)
	row := container.NewThemeOverride(container.NewBorder(nil, nil, refs.bar, nil, body), denseTheme{NewTheme()})
	m.rows[row] = refs
	return row
}
