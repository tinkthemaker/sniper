package clipboard

import "fyne.io/fyne/v2"

// Writer copies text to the system clipboard.
type Writer interface{ SetText(text string) }

type writer struct{ cb fyne.Clipboard }

// New wraps a Fyne clipboard (app.Clipboard()) so text gets platform line endings.
func New(cb fyne.Clipboard) Writer { return writer{cb: cb} }

func (w writer) SetText(text string) { w.cb.SetContent(ToPlatform(text)) }
