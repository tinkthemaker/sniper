package ui

import (
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/tinkthemaker/sniper/internal/clipboard"
	"github.com/tinkthemaker/sniper/internal/snippets"
)

// Storage is the subset of *store.Store the UI uses. It is an interface only so
// UI tests can wrap a real store and inject Save failures.
type Storage interface {
	Dir() string
	Save(lib snippets.Library) error
	RestoreBackup() (snippets.Library, error)
	QuarantinePrimary(now time.Time) (string, error)
}

// Deps are the collaborators the UI needs, injected so tests can fake them.
type Deps struct {
	App        fyne.App
	Store      Storage
	Clipboard  clipboard.Writer
	OpenFolder func(path string) error
	Now        func() time.Time
}

// MainWindow owns the library and everything shown for it.
type MainWindow struct {
	deps Deps
	win  fyne.Window

	lib        snippets.Library
	query      string
	results    []snippets.Snippet
	selectedID string

	syncing bool // true while we drive list.Select, so OnSelected ignores it

	lastTapID string    // the row last clicked, to recognize a double-click
	lastTapAt time.Time // when it was clicked

	undo     []undoStep     // library changes saved this session, newest last
	undoItem *fyne.MenuItem // Edit → Undo

	editor  *editor        // the open editor, if any
	dialogs []*shownDialog // dialogs open over the window, oldest first

	rows         map[fyne.CanvasObject]*rowRefs
	countSeg     *widget.TextSegment
	count        *widget.RichText
	search       *searchEntry
	list         *resultList
	emptyLabel   *widget.Label
	previewHead  *widget.TextSegment
	previewLabel *widget.RichText
	previewTitle *widget.Label
	previewBody  *widget.Label
	copyBtn      *widget.Button
	newBtn       *widget.Button
	editBtn      *widget.Button
	deleteBtn    *widget.Button
	pills        []*pill // the four action pills, in order
	frame        *frame
	parts        *panelParts
	status       *widget.Label
}

// ShowMain creates and shows the main window for lib.
func ShowMain(d Deps, lib snippets.Library) *MainWindow {
	m := &MainWindow{deps: d, lib: lib, rows: map[fyne.CanvasObject]*rowRefs{}}
	m.win = d.App.NewWindow(appTitle)
	m.win.SetMaster()      // closing the main window quits the app
	m.win.SetPadded(false) // the layout sets every margin itself (chrome.go)
	m.build()
	m.win.SetContent(m.layout())
	m.win.Resize(fyne.NewSize(520, 600))
	m.registerShortcuts()
	m.win.SetMainMenu(m.buildMenu())
	m.refresh("")
	m.win.Show()
	m.win.Canvas().Focus(m.search)
	return m
}

func (m *MainWindow) Window() fyne.Window         { return m.win }
func (m *MainWindow) Library() snippets.Library   { return m.lib }
func (m *MainWindow) Results() []snippets.Snippet { return m.results }
func (m *MainWindow) SelectedID() string          { return m.selectedID }
func (m *MainWindow) Status() string              { return m.status.Text }

func (m *MainWindow) build() {
	m.search = newSearchEntry(m)
	m.search.TextStyle = fyne.TextStyle{Bold: true}
	m.search.OnChanged = func(q string) {
		m.query = q
		m.refresh("")
	}
	m.list = newResultList(m)
	m.list.HideSeparators = true
	m.emptyLabel = widget.NewLabel("")
	m.emptyLabel.Alignment = fyne.TextAlignCenter
	m.emptyLabel.Wrapping = fyne.TextWrapWord

	m.previewHead = capsSeg(previewLabel, nameLavender, sizeCode)
	m.previewLabel = widget.NewRichText(m.previewHead)
	m.previewTitle = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	m.previewTitle.Truncation = fyne.TextTruncateEllipsis
	m.previewBody = widget.NewLabel("")
	m.previewBody.Selectable = true
	m.previewBody.Wrapping = fyne.TextWrapWord

	// COPY is rounded on its left end and DELETE on its right, like the ends of a bar.
	copyP := newPill(btnCopy, m.copySelected, colOrange, 20, copyHeight/2, 0)
	newP := newPill(btnNew, m.newSnippet, colPeri, 15, 0, 0)
	editP := newPill(btnEdit, m.editSelected, colLavender, 15, 0, 0)
	deleteP := newPill(btnDelete, m.deleteSelected, colRed, 15, 0, buttonHeight/2)
	m.pills = []*pill{copyP, newP, editP, deleteP}
	m.copyBtn, m.newBtn, m.editBtn, m.deleteBtn = copyP.btn, newP.btn, editP.btn, deleteP.btn

	m.countSeg = capsSeg("", nameBlack, sizeCode)
	m.countSeg.Style.Alignment = fyne.TextAlignTrailing
	m.count = widget.NewRichText(m.countSeg)
	m.status = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	m.status.Truncation = fyne.TextTruncateEllipsis
}

// trailing is a column label: black capitals, aligned to the block's right edge.
func trailing(text string) *widget.RichText {
	seg := capsSeg(text, nameBlack, sizeCode)
	seg.Style.Alignment = fyne.TextAlignTrailing
	return widget.NewRichText(seg)
}

// layout builds the LCARS frame (lcars.go) around the panel.
func (m *MainWindow) layout() fyne.CanvasObject {
	parts := &panelParts{
		search:      container.NewThemeOverride(m.search, searchTheme{NewTheme()}),
		listArea:    container.NewThemeOverride(container.NewStack(m.list, container.NewCenter(m.emptyLabel)), denseTheme{NewTheme()}),
		previewHead: m.previewLabel,
		previewName: container.NewThemeOverride(m.previewTitle, labelTheme{NewTheme(), 22, colOrange}),
		previewBody: container.NewVScroll(m.previewBody),
		buttons: container.New(buttonRowLayout{},
			m.pills[0].obj, m.pills[1].obj, m.pills[2].obj, m.pills[3].obj),
	}
	status := container.NewThemeOverride(m.status, labelTheme{NewTheme(), 13, colCanvas})
	texts := [4]*widget.RichText{m.count, trailing(searchBlockLabel), trailing(previewLabel), trailing(versionText(m.deps.App.Metadata().Version))}
	var root fyne.CanvasObject
	m.parts = parts
	m.frame, root = newFrame(texts, status, newPanel(parts))
	return root
}

// refresh recomputes the results for the current query and selects preferID if
// it is among them, otherwise the first result.
func (m *MainWindow) refresh(preferID string) {
	m.results = snippets.Search(m.lib.Snippets, m.query)
	m.selectedID = ""
	if i := m.indexOf(preferID); i >= 0 {
		m.selectedID = m.results[i].ID
	} else if len(m.results) > 0 {
		m.selectedID = m.results[0].ID
	}
	m.list.Refresh()
	m.syncView()
}

func (m *MainWindow) updateCount() {
	// A query of only spaces matches everything (search splits on whitespace),
	// so it counts as no query here too.
	if strings.TrimSpace(m.query) == "" {
		m.countSeg.Text = countAllText(len(m.lib.Snippets))
	} else {
		m.countSeg.Text = countMatchText(len(m.results), len(m.lib.Snippets))
	}
	m.count.Refresh()
}

// syncView pushes the selection to the list, preview, buttons and empty state.
func (m *MainWindow) syncView() {
	m.syncing = true
	if i := m.indexOf(m.selectedID); i >= 0 {
		m.list.Select(i)
		m.list.Highlight(i)
		m.list.ScrollTo(i)
	} else {
		m.list.UnselectAll()
	}
	m.syncing = false
	m.list.Refresh() // the selection bar is drawn by the rows themselves
	m.updateCount()

	sel, has := m.selected()
	if has {
		m.previewHead.Text = previewRecordLabel(recordCode(sel.ID))
		m.previewTitle.SetText(strings.ToUpper(sel.Title))
		m.previewBody.SetText(sel.Body)
	} else {
		m.previewHead.Text = previewLabel
		m.previewTitle.SetText("")
		m.previewBody.SetText(previewEmptyText)
	}
	m.previewLabel.Refresh()
	for _, i := range []int{0, 2, 3} { // COPY, EDIT and DELETE need a selection; NEW never does
		m.pills[i].setEnabled(has)
	}

	switch {
	case len(m.lib.Snippets) == 0:
		m.emptyLabel.SetText(emptyLibraryText)
	case len(m.results) == 0:
		m.emptyLabel.SetText(noMatchesText(m.query))
	}
	if len(m.results) == 0 {
		m.emptyLabel.Show()
		m.list.Hide()
	} else {
		m.emptyLabel.Hide()
		m.list.Show()
	}
}

func (m *MainWindow) updateRow(i widget.ListItemID, o fyne.CanvasObject) {
	if i >= len(m.results) {
		return
	}
	refs := m.rows[o]
	if refs == nil {
		return
	}
	s := m.results[i]
	refs.show(s.ID, s.Title, s.ID == m.selectedID)
}

func (m *MainWindow) onListSelected(i widget.ListItemID) {
	if m.syncing || i >= len(m.results) {
		return
	}
	m.selectedID = m.results[i].ID
	m.syncView()
}

// tapRow handles a click on the row showing snippet id: it selects the row at
// once, and copies it too when this is the second click on it within
// doubleClickWindow. The row is found by ID, never by index, so it is right
// after filtering and scrolling. It does nothing while a dialog is open.
func (m *MainWindow) tapRow(id string) {
	if m.modalOpen() || m.indexOf(id) < 0 {
		return
	}
	m.win.Canvas().Focus(m.list)
	if id != m.selectedID {
		m.selectedID = id
		m.syncView()
	}
	now := m.deps.Now()
	if id == m.lastTapID && now.Sub(m.lastTapAt) <= doubleClickWindow {
		m.lastTapID = ""
		m.copySelected()
		return
	}
	m.lastTapID, m.lastTapAt = id, now
}

func (m *MainWindow) indexOf(id string) int {
	if id == "" {
		return -1
	}
	for i := range m.results {
		if m.results[i].ID == id {
			return i
		}
	}
	return -1
}

func (m *MainWindow) selected() (snippets.Snippet, bool) {
	if i := m.indexOf(m.selectedID); i >= 0 {
		return m.results[i], true
	}
	return snippets.Snippet{}, false
}

// moveSelection moves the selection by delta, clamped at both ends (no wrap).
func (m *MainWindow) moveSelection(delta int) {
	if len(m.results) == 0 {
		return
	}
	i := min(max(m.indexOf(m.selectedID)+delta, 0), len(m.results)-1)
	m.selectedID = m.results[i].ID
	m.syncView()
}

func (m *MainWindow) copySelected() {
	s, ok := m.selected()
	if !ok {
		return
	}
	m.deps.Clipboard.SetText(s.Body)
	m.setStatus(copiedStatus(s.Title))
}

func (m *MainWindow) setStatus(text string) { m.status.SetText(text) }

func (m *MainWindow) clearQuery() {
	if m.search.Text != "" {
		m.search.SetText("")
	}
}

// focusSearch focuses the search box and selects all of its text.
func (m *MainWindow) focusSearch() {
	m.win.Canvas().Focus(m.search)
	m.search.TypedShortcut(&fyne.ShortcutSelectAll{})
}
