package ui

import (
	"image/color"
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Slayer finish: blood red on black iron, with ember as the hot end. The values
// are derived the same way as in the Screech tool (warm grays pulled toward the
// accent hue). Blood red is only 3.3:1 on the canvas, so it never colors small
// text; text that needs red uses redText, and the action color is ember.
var (
	ironCanvas  = rgb(0x0A, 0x0A, 0x0A) // window background
	ironPanel   = rgb(0x10, 0x09, 0x08) // inputs, menus, dialogs
	ironSurface = rgb(0x21, 0x12, 0x11) // strips and buttons
	ironSelect  = rgb(0x35, 0x1A, 0x19) // selected row
	ironHover   = rgb(0x2A, 0x16, 0x14)
	ironEdge    = rgb(0x51, 0x0A, 0x09) // metal edge: borders and rules
	ironRule    = rgb(0x3A, 0x13, 0x11) // quiet separators

	textBright = rgb(0xF5, 0xEA, 0xEA)
	textMid    = rgb(0xC0, 0xAE, 0xAC)
	textDim    = rgb(0xA3, 0x8A, 0x89)

	bloodRed = rgb(0xC2, 0x1A, 0x16) // structure: bars, edges (never small text)
	redText  = rgb(0xE8, 0x43, 0x3C) // red that must be read: errors, danger fills
	ember    = rgb(0xFF, 0x6A, 0x18) // action and focus
)

func rgb(r, g, b uint8) color.Color { return color.NRGBA{R: r, G: g, B: b, A: 0xFF} }

// slayerTheme is always dark: the finish is black iron, whatever the OS setting.
type slayerTheme struct {
	base fyne.Theme
	font func(fyne.TextStyle) fyne.Resource
}

// NewTheme returns the Slayer theme. Text uses Consolas from the Windows fonts
// folder when it is installed and Fyne's default fonts otherwise.
func NewTheme() fyne.Theme {
	return &slayerTheme{base: theme.DefaultTheme(), font: systemFonts()}
}

func (t *slayerTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return ironCanvas
	case theme.ColorNameForeground:
		return textBright
	case theme.ColorNameDisabled:
		return rgb(0x6B, 0x5A, 0x59)
	case theme.ColorNamePlaceHolder:
		return textDim
	case theme.ColorNamePrimary, theme.ColorNameHyperlink:
		return ember
	case theme.ColorNameForegroundOnPrimary, theme.ColorNameForegroundOnError,
		theme.ColorNameForegroundOnWarning, theme.ColorNameForegroundOnSuccess:
		return ironCanvas
	case theme.ColorNameButton, theme.ColorNameHeaderBackground:
		return ironSurface
	case theme.ColorNameDisabledButton:
		return rgb(0x14, 0x0D, 0x0C)
	case theme.ColorNameHover:
		return ironHover
	case theme.ColorNamePressed, theme.ColorNameSelection:
		return ironSelect
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0xFF, G: 0x6A, B: 0x18, A: 0x66}
	case theme.ColorNameInputBackground, theme.ColorNameMenuBackground:
		return ironPanel
	case theme.ColorNameOverlayBackground:
		return ironSurface // dialogs sit one step above the window
	case theme.ColorNameInputBorder:
		return ironEdge
	case theme.ColorNameSeparator:
		return ironRule
	case theme.ColorNameScrollBar:
		return ironEdge
	case theme.ColorNameScrollBarBackground:
		return color.Transparent
	case theme.ColorNameShadow:
		return color.NRGBA{A: 0x66}
	case theme.ColorNameError:
		return redText
	case theme.ColorNameWarning, theme.ColorNameSuccess:
		return ember
	}
	return t.base.Color(name, theme.VariantDark)
}

func (t *slayerTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameInnerPadding:
		return 6
	case theme.SizeNameInputBorder, theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInputRadius, theme.SizeNameButtonRadius, theme.SizeNameSelectionRadius,
		theme.SizeNameCardRadius, theme.SizeNamePopupRadius, theme.SizeNameMenuRadius,
		theme.SizeNameDialogRadius:
		return 2
	case theme.SizeNameScrollBar:
		return 10
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameScrollBarRadius:
		return 0
	case theme.SizeNameSplitThickness:
		return 1
	}
	return t.base.Size(name)
}

func (t *slayerTheme) Font(s fyne.TextStyle) fyne.Resource {
	if t.font != nil {
		if r := t.font(s); r != nil {
			return r
		}
	}
	return t.base.Font(s)
}

// denseTheme tightens padding for list rows so about five fit in the list.
type denseTheme struct{ fyne.Theme }

func (d denseTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameInnerPadding || name == theme.SizeNamePadding {
		return 2
	}
	return d.Theme.Size(name)
}

func (t *slayerTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return t.base.Icon(name) }

// systemFonts loads Consolas (regular, bold, italic, bold italic) from the
// Windows fonts folder. It returns nil on other systems, or if any file is missing.
func systemFonts() func(fyne.TextStyle) fyne.Resource {
	if runtime.GOOS != "windows" {
		return nil
	}
	dir := filepath.Join(os.Getenv("SystemRoot"), "Fonts")
	load := func(file string) fyne.Resource {
		data, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			return nil
		}
		return fyne.NewStaticResource(file, data)
	}
	regular, bold := load("consola.ttf"), load("consolab.ttf")
	italic, boldItalic := load("consolai.ttf"), load("consolaz.ttf")
	if regular == nil || bold == nil || italic == nil || boldItalic == nil {
		return nil
	}
	return func(s fyne.TextStyle) fyne.Resource {
		switch {
		case s.Symbol:
			return nil // keep Fyne's symbol font
		case s.Bold && s.Italic:
			return boldItalic
		case s.Bold:
			return bold
		case s.Italic:
			return italic
		}
		return regular
	}
}
