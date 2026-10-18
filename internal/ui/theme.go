package ui

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// The palette is the one in docs/design/lcars-mockup-v3.html: a black canvas,
// seven pastel block colors, and text in tan or a warm white. Text that sits on
// a colored block is always black; TestThemeContrastFloors keeps every pairing
// at 4.5:1 or better.
var (
	colCanvas   = rgb(0x00, 0x00, 0x00) // window background, inputs, menus
	colSurface  = rgb(0x1B, 0x15, 0x1B) // dialogs: one step above the canvas so they stand off it
	colRaised   = rgb(0x33, 0x27, 0x36) // quiet buttons on dialogs
	colSelect   = rgb(0x5B, 0x4A, 0x6B) // selected text in an entry
	colHover    = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x1C}
	colFocus    = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x38}
	colScroll   = color.NRGBA{R: 0xFF, G: 0x99, B: 0x66, A: 0x99}
	colOrange   = rgb(0xFF, 0x99, 0x00) // action and focus
	colPeach    = rgb(0xFF, 0x99, 0x66) // frame
	colGold     = rgb(0xFF, 0xCC, 0x66)
	colTan      = rgb(0xFF, 0xCC, 0x99) // list text
	colLavender = rgb(0xCC, 0x99, 0xCC)
	colPeri     = rgb(0x99, 0x99, 0xFF)
	colRed      = rgb(0xCC, 0x66, 0x66)
	colInk      = rgb(0xF1, 0xE3, 0xD3) // running text
	colDim      = rgb(0x9A, 0x8A, 0x9A) // placeholders
	colOff      = rgb(0x7A, 0x6B, 0x7A) // disabled text
)

func rgb(r, g, b uint8) color.Color { return color.NRGBA{R: r, G: g, B: b, A: 0xFF} }

// Theme color and size names the main window's custom text uses. They exist so
// a RichText segment can pick a palette color or a size by name.
const (
	nameBlack    fyne.ThemeColorName = "sniperBlack"
	nameOrange   fyne.ThemeColorName = "sniperOrange"
	nameTan      fyne.ThemeColorName = "sniperTan"
	nameLavender fyne.ThemeColorName = "sniperLavender"
	namePeri     fyne.ThemeColorName = "sniperPeri"

	sizeRow   fyne.ThemeSizeName = "sniperRow"   // list titles
	sizeCode  fyne.ThemeSizeName = "sniperCode"  // record codes, block labels, the FIND pill
	sizeTitle fyne.ThemeSizeName = "sniperTitle" // the preview title
	sizeBrand fyne.ThemeSizeName = "sniperBrand" // SNIPER in the top bar
	sizeHint  fyne.ThemeSizeName = "sniperHint"  // the bottom bar
)

// lcarsTheme is always dark: the finish is black, whatever the OS setting.
type lcarsTheme struct {
	base fyne.Theme
	font func(fyne.TextStyle) fyne.Resource
}

// NewTheme returns the LCARS-style theme, with the embedded Antonio and Jost faces.
func NewTheme() fyne.Theme {
	return &lcarsTheme{base: theme.DefaultTheme(), font: lcarsFont}
}

func (t *lcarsTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case nameBlack, theme.ColorNameForegroundOnPrimary, theme.ColorNameForegroundOnError,
		theme.ColorNameForegroundOnWarning, theme.ColorNameForegroundOnSuccess:
		return colCanvas
	case nameOrange:
		return colOrange
	case nameTan:
		return colTan
	case nameLavender:
		return colLavender
	case namePeri:
		return colPeri
	case theme.ColorNameBackground, theme.ColorNameInputBackground, theme.ColorNameMenuBackground:
		return colCanvas
	case theme.ColorNameForeground:
		return colInk
	case theme.ColorNameDisabled:
		return colOff
	case theme.ColorNamePlaceHolder:
		return colDim
	case theme.ColorNamePrimary, theme.ColorNameHyperlink, theme.ColorNameWarning, theme.ColorNameSuccess:
		return colOrange
	case theme.ColorNameButton:
		return colRaised
	case theme.ColorNameDisabledButton, theme.ColorNameHeaderBackground:
		return colSurface
	case theme.ColorNameHover:
		return colHover
	case theme.ColorNamePressed, theme.ColorNameSelection:
		return colSelect
	case theme.ColorNameFocus:
		return colFocus
	case theme.ColorNameOverlayBackground:
		return colSurface
	case theme.ColorNameInputBorder, theme.ColorNameScrollBar:
		return colScroll
	case theme.ColorNameSeparator:
		return colRaised
	case theme.ColorNameScrollBarBackground:
		return color.Transparent
	case theme.ColorNameShadow:
		return color.NRGBA{A: 0x66}
	case theme.ColorNameError:
		return colRed
	}
	return t.base.Color(name, theme.VariantDark)
}

func (t *lcarsTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14
	case theme.SizeNameCaptionText:
		return 12
	case sizeRow:
		return 17
	case sizeCode:
		return 13
	case sizeTitle:
		return 22
	case sizeBrand:
		return 34
	case sizeHint:
		return 12
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameInnerPadding:
		return 4 // widget text inset; pillLine + 8 = textLine (chrome.go)
	case theme.SizeNameInputBorder, theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameButtonRadius:
		return 100 // clamped to half the height: a full pill
	case theme.SizeNameInputRadius:
		return 6 // Fyne's own fields (the file dialog's name box); ours use fieldTheme
	case theme.SizeNameSelectionRadius:
		return rowHeight / 2 // a list row's hover and selection fills are pills
	case theme.SizeNameCardRadius, theme.SizeNamePopupRadius, theme.SizeNameMenuRadius, theme.SizeNameDialogRadius:
		return 4
	case theme.SizeNameScrollBar:
		return 10
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameScrollBarRadius:
		return 0
	}
	return t.base.Size(name)
}

func (t *lcarsTheme) Font(s fyne.TextStyle) fyne.Resource {
	if t.font != nil {
		if r := t.font(s); r != nil {
			return r
		}
	}
	return t.base.Font(s)
}

func (t *lcarsTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return t.base.Icon(name) }

// denseTheme removes the gap the list adds between rows, so the row pitch is
// the row height. It leaves the inner padding alone, because that is what keeps
// row text where it is. Wrap only the list in it. (The list's own items ignore
// a theme override, so their pill-shaped hover and selection fills come from
// the app theme's selection radius instead.)
type denseTheme struct{ fyne.Theme }

func (d denseTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNamePadding {
		return 0
	}
	return d.Theme.Size(name)
}

// pillTheme dresses one main-window button: black label, flat fill, a veil for
// hover and focus, and the given text size. The colored fill is drawn behind
// the button by pill(), so each end can be rounded on its own. The inner
// padding is the gap between the label and the button's right end.
type pillTheme struct {
	fyne.Theme
	text   float32
	radius float32
}

func (p pillTheme) Color(name fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameForeground:
		return colCanvas
	case theme.ColorNameButton, theme.ColorNameDisabledButton:
		return color.Transparent
	}
	return p.Theme.Color(name, v)
}

func (p pillTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return p.text
	case theme.SizeNameInnerPadding:
		return pillPad
	case theme.SizeNameButtonRadius:
		return p.radius
	}
	return p.Theme.Size(name)
}

// searchTheme strips the search entry down to its text: the outline is drawn by
// the frame (so the left end can be square), so the entry has no fill, border
// or focus ring of its own.
type searchTheme struct{ fyne.Theme }

func (s searchTheme) Color(name fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameInputBackground:
		return color.Transparent
	case theme.ColorNameForeground:
		return colTan
	case theme.ColorNameScrollBar:
		return color.Transparent // a one-line field never scrolls up or down
	}
	return s.Theme.Color(name, v)
}

func (s searchTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInputBorder:
		return 0
	case theme.SizeNameText:
		return 16
	}
	return s.Theme.Size(name)
}

// fieldTheme is for the editor's one-line field: a pill, with its text pulled
// in from the curve. Fyne insets entry text by the inner padding (4) plus the
// border, which leaves the text close to a 14 px curve, and raising the inner
// padding would also make the field taller. An entry that has an icon insets
// its text by InlineIcon + LineSpacing instead, so the field gets a blank icon
// (insetField) and this theme sizes it: InlineIcon plus the 4 px line spacing
// = fieldInset on the left, with a spacer of the same width on the right. The
// height does not change.
type fieldTheme struct{ fyne.Theme }

func (f fieldTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInputRadius:
		return fieldRadius
	case theme.SizeNameInlineIcon:
		return fieldInset - f.Theme.Size(theme.SizeNameLineSpacing)
	}
	return f.Theme.Size(name)
}

// bodyTheme is for the editor's multi-line field, which fills the room it is
// given, so a bigger inner padding costs no height anywhere that matters. It
// moves the first line down and in from the corner as well, which the icon
// trick cannot do.
type bodyTheme struct{ fyne.Theme }

func (f bodyTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInputRadius:
		return fieldRadius
	case theme.SizeNameInnerPadding:
		return bodyPad
	}
	return f.Theme.Size(name)
}

// labelTheme sets the text size and color of one label.
type labelTheme struct {
	fyne.Theme
	text float32
	fg   color.Color
}

func (s labelTheme) Color(name fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameForeground && s.fg != nil {
		return s.fg
	}
	return s.Theme.Color(name, v)
}

func (s labelTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return s.text
	}
	return s.Theme.Size(name)
}

// Two faces, both free under the SIL Open Font License (fonts/OFL.txt for
// Jost, fonts/OFL-Antonio.txt for Antonio) and embedded so every PC renders the
// same. Antonio, a tall condensed gothic, carries everything bold: titles,
// labels, buttons and headings. Jost, a Futura-spirited geometric sans, carries
// running text; Medium (500) because the 400 Book weight is thin on black.
//
//go:embed fonts/Antonio-SemiBold.ttf
var antonioSemiBold []byte

//go:embed fonts/Jost-500-Medium.ttf
var jostMedium []byte

//go:embed fonts/Jost-500-MediumItalic.ttf
var jostMediumItalic []byte

//go:embed fonts/Jost-600-SemiItalic.ttf
var jostSemiItalic []byte

var (
	fontRegular    = fyne.NewStaticResource("Jost-500-Medium.ttf", jostMedium)
	fontItalic     = fyne.NewStaticResource("Jost-500-MediumItalic.ttf", jostMediumItalic)
	fontBold       = fyne.NewStaticResource("Antonio-SemiBold.ttf", antonioSemiBold)
	fontBoldItalic = fyne.NewStaticResource("Jost-600-SemiItalic.ttf", jostSemiItalic)
)

// lcarsFont picks the font file for a text style. Symbol and monospace text keep
// Fyne's own fonts (nil means "use the base theme").
func lcarsFont(s fyne.TextStyle) fyne.Resource {
	switch {
	case s.Symbol, s.Monospace:
		return nil
	case s.Bold && s.Italic:
		return fontBoldItalic
	case s.Bold:
		return fontBold
	case s.Italic:
		return fontItalic
	}
	return fontRegular
}
