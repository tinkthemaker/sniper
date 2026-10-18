package ui

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/tinkthemaker/sniper/internal/snippets"
)

// editorEntry forwards the editor's shortcuts (Ctrl+S, Escape) because a focused
// widget.Entry would otherwise swallow them.
type editorEntry struct {
	widget.Entry
	ed *editor
}

func newEditorEntry(ed *editor, multiLine bool) *editorEntry {
	e := &editorEntry{ed: ed}
	e.MultiLine = multiLine
	if !multiLine {
		insetField(&e.Entry)
	}
	e.ExtendBaseWidget(e)
	return e
}

func (e *editorEntry) TypedShortcut(s fyne.Shortcut) {
	if e.ed.handleShortcut(s) {
		return
	}
	e.Entry.TypedShortcut(s)
}

func (e *editorEntry) TypedKey(k *fyne.KeyEvent) {
	if k.Name == fyne.KeyEscape {
		e.ed.cancel()
		return
	}
	e.Entry.TypedKey(k)
}

// editor is the modal New/Edit form.
type editor struct {
	m  *MainWindow
	id string // empty when creating

	dlg       *dialog.CustomDialog
	title     *editorEntry
	body      *editorEntry
	errLabel  *widget.Label
	cancelBtn *widget.Button

	initialTitle, initialBody string
	closed                    bool
}

func (m *MainWindow) newSnippet() { m.openEditor(nil) }

func (m *MainWindow) editSelected() {
	if s, ok := m.selected(); ok {
		m.openEditor(&s)
	}
}

func (m *MainWindow) openEditor(existing *snippets.Snippet) {
	ed := &editor{m: m}
	heading := editorTitleNew
	if existing != nil {
		ed.id, ed.initialTitle, ed.initialBody = existing.ID, existing.Title, existing.Body
		heading = editorTitleEdit
	}
	ed.title = newEditorEntry(ed, false)
	ed.title.SetText(ed.initialTitle)
	ed.title.OnSubmitted = func(string) { m.win.Canvas().Focus(ed.body) }
	ed.body = newEditorEntry(ed, true)
	ed.body.Wrapping = fyne.TextWrapWord
	ed.body.SetText(ed.initialBody)
	ed.errLabel = widget.NewLabel("")
	ed.errLabel.Importance = widget.DangerImportance
	ed.errLabel.Wrapping = fyne.TextWrapWord
	ed.errLabel.Hide()

	ed.cancelBtn = widget.NewButton(labelCancel, ed.cancel)
	save := widget.NewButton(labelSave, ed.save)
	save.Importance = widget.HighImportance

	// Field labels sit above their fields (the Windows text box pattern), so
	// both fields share one left edge and the full width. A label's inner
	// padding plus a 4 px gap puts its text 8 px from the field.
	stack := func(objs ...fyne.CanvasObject) *fyne.Container {
		return container.New(layout.NewCustomPaddedVBoxLayout(4), objs...)
	}
	top := stack(widget.NewLabelWithStyle(heading, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		stack(widget.NewLabel(labelTitle), container.NewThemeOverride(ed.title, fieldTheme{NewTheme()})))
	bodyLabel := widget.NewLabel(labelBody)
	body := inset(container.New(layout.NewBorderLayout(bodyLabel, nil, nil, nil), bodyLabel, container.NewThemeOverride(ed.body, bodyTheme{NewTheme()})), gapItem, 0, 0, 0)
	buttons := container.New(layout.NewCustomPaddedHBoxLayout(gapItem), layout.NewSpacer(), ed.cancelBtn, save)
	bottom := inset(stack(ed.errLabel, buttons), gapItem, 0, 0, 0)
	// Tab order follows child order, and NewBorder puts the center first, so
	// list the children in visual order: title, body, then the buttons.
	content := container.New(layout.NewBorderLayout(top, bottom, nil, nil), top, body, bottom)

	// The heading is drawn above as part of the content, so that it lines up
	// with everything else; the dialog's own heading is left empty.
	ed.dlg = dialog.NewCustomWithoutButtons("", onDialogGrid(content), m.win)
	size := m.win.Canvas().Size()
	ed.dlg.Resize(fyne.NewSize(max(size.Width-40, 420), max(size.Height-40, 380)))
	m.editor = ed
	m.present(ed.dlg, heading, "")
	ed.dlg.SetOnClosed(func() { m.editor = nil })
	if existing != nil {
		m.win.Canvas().Focus(ed.body)
	} else {
		m.win.Canvas().Focus(ed.title)
	}
}

func (ed *editor) handleShortcut(s fyne.Shortcut) bool {
	if s.ShortcutName() != shortcutSave.ShortcutName() {
		return false
	}
	ed.save()
	return true
}

func (ed *editor) dirty() bool {
	return ed.title.Text != ed.initialTitle || ed.body.Text != ed.initialBody
}

func (ed *editor) showError(text string) {
	ed.errLabel.SetText(text)
	ed.errLabel.Show()
}

// save validates, persists, then updates the window. A failure keeps the form open.
func (ed *editor) save() {
	if ed.closed {
		return
	}
	m := ed.m
	title, body := snippets.Normalize(ed.title.Text, ed.body.Text)
	if err := snippets.ValidateFields(title, body); err != nil {
		ed.showError(err.Error())
		var ve *snippets.ValidationError
		if errors.As(err, &ve) && ve.Field == "body" {
			m.win.Canvas().Focus(ed.body)
		} else {
			m.win.Canvas().Focus(ed.title)
		}
		return
	}

	var (
		next snippets.Library
		item snippets.Snippet
		err  error
	)
	if ed.id == "" {
		next, item, err = snippets.Add(m.lib, title, body)
	} else {
		next, item, err = snippets.Update(m.lib, ed.id, title, body)
	}
	if err != nil {
		ed.showError(err.Error())
		return
	}
	if err := m.deps.Store.Save(next); err != nil {
		ed.showError(saveFailedText(err))
		return
	}
	kind := undoEdit
	if ed.id == "" {
		kind = undoCreate
	}
	m.pushUndo(undoStep{kind: kind, prev: m.lib, id: item.ID, title: item.Title})
	m.lib = next
	ed.close()
	m.showSaved(item)
}

// cancel closes the form, asking first if anything changed.
func (ed *editor) cancel() {
	if ed.closed {
		return
	}
	if !ed.dirty() {
		ed.close()
		return
	}
	d := dialog.NewConfirm(discardTitle, discardMessage, func(discard bool) {
		if discard {
			ed.close()
		} else if !ed.closed {
			ed.m.win.Canvas().Focus(ed.body)
		}
	}, ed.m.win)
	d.SetConfirmText(labelDiscard)
	d.SetDismissText(labelKeepEditing)
	ed.m.present(d, discardTitle, discardMessage)
}

func (ed *editor) close() {
	ed.closed = true
	ed.dlg.Hide()
	ed.m.win.Canvas().Focus(ed.m.search)
}

// showSaved selects the saved snippet (clearing the query if it would hide it)
// and reports the save.
func (m *MainWindow) showSaved(item snippets.Snippet) {
	if !m.inResults(item.ID) {
		m.clearQuery()
	}
	m.refresh(item.ID)
	m.setStatus(savedStatus(item.Title))
}

func (m *MainWindow) inResults(id string) bool {
	for _, s := range snippets.Search(m.lib.Snippets, m.query) {
		if s.ID == id {
			return true
		}
	}
	return false
}
