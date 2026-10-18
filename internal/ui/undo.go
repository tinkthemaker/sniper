package ui

import "github.com/tinkthemaker/sniper/internal/snippets"

// maxUndo is how many library changes undo can step back through.
const maxUndo = 20

type undoKind int

const (
	undoCreate undoKind = iota
	undoEdit
	undoDelete
	undoImport
)

// undoStep is one saved change: prev is the library as it was before it.
// Libraries are immutable values, so keeping them is cheap.
type undoStep struct {
	kind  undoKind
	prev  snippets.Library
	id    string // the snippet created, edited or deleted
	title string // its title at the time of the change
	n     int    // snippets imported
}

// pushUndo records a change that has just been saved successfully.
func (m *MainWindow) pushUndo(s undoStep) {
	m.undo = append(m.undo, s)
	if len(m.undo) > maxUndo {
		m.undo = append([]undoStep(nil), m.undo[len(m.undo)-maxUndo:]...)
	}
	m.syncUndoMenu()
}

// syncUndoMenu enables the Edit → Undo item only while there is something to undo.
func (m *MainWindow) syncUndoMenu() {
	if m.undoItem == nil {
		return
	}
	m.undoItem.Disabled = len(m.undo) == 0
	if mm := m.win.MainMenu(); mm != nil {
		mm.Refresh()
	}
}

// undoLast reverts the most recent change through Store.Save like any other
// change. If the save fails the stack and the library are left as they were.
func (m *MainWindow) undoLast() {
	if len(m.undo) == 0 {
		m.setStatus(nothingToUndoStatus)
		return
	}
	step := m.undo[len(m.undo)-1]
	if err := m.deps.Store.Save(step.prev); err != nil {
		m.showError(undoFailedText(err))
		return
	}
	// Where the selection sits now, for the delete selection rules below.
	at := m.indexOf(m.selectedID)
	keep := m.selectedID
	m.undo = m.undo[:len(m.undo)-1]
	m.lib = step.prev
	m.syncUndoMenu()

	switch step.kind {
	case undoEdit, undoDelete:
		if !m.inResults(step.id) {
			m.clearQuery()
		}
		m.refresh(step.id)
	default: // create, import: the snippets are gone, so pick as after a delete
		if m.inResults(keep) {
			m.refresh(keep)
		} else {
			m.refresh(m.pickAfterRemoval(at))
		}
	}
	m.setStatus(step.status())
}

func (s undoStep) status() string {
	switch s.kind {
	case undoCreate:
		return undidCreateStatus(s.title)
	case undoEdit:
		return undidEditStatus(s.title)
	case undoDelete:
		return undidDeleteStatus(s.title)
	}
	return undidImportStatus(s.n)
}

// pickAfterRemoval returns the result now at index at, else the last one, else
// none. It is the selection rule after a snippet at index at is removed.
func (m *MainWindow) pickAfterRemoval(at int) string {
	remaining := snippets.Search(m.lib.Snippets, m.query)
	switch {
	case at >= 0 && at < len(remaining):
		return remaining[at].ID
	case len(remaining) > 0:
		return remaining[len(remaining)-1].ID
	}
	return ""
}
