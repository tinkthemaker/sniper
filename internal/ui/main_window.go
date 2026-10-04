package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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

	editor  *editor        // the open editor, if any
	dialogs []*shownDialog // dialogs open over the window, oldest first

	rows         map[fyne.CanvasObject]*rowRefs
	countSeg     *widget.TextSegment
	count        *widget.RichText
	search       *searchEntry
	list         *resultList
	emptyLabel   *widget.Label
	previewTitle *widget.Label
	previewBody  *widget.Label
	copyBtn      *widget.Button
	newBtn       *widget.Button
	editBtn      *widget.Button
	deleteBtn    *widget.Button
	status       *widget.Label
}

// ShowMain creates and shows the main window for lib.
func ShowMain(d Deps, lib snippets.Library) *MainWindow {
	m := &MainWindow{deps: d, lib: lib, rows: map[fyne.CanvasObject]*rowRefs{}}
	m.win = d.App.NewWindow(appTitle)
	m.win.SetMaster() // closing the main window quits the app
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
	m.search.OnChanged = func(q string) {
		m.query = q
		m.refresh("")
	}
	m.list = newResultList(m)
	m.emptyLabel = widget.NewLabel("")
	m.emptyLabel.Alignment = fyne.TextAlignCenter
	m.emptyLabel.Wrapping = fyne.TextWrapWord

	m.previewTitle = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	m.previewTitle.Truncation = fyne.TextTruncateEllipsis
	m.previewBody = widget.NewLabel("")
	m.previewBody.Selectable = true
	m.previewBody.Wrapping = fyne.TextWrapWord

	m.copyBtn = widget.NewButton(labelCopy, m.copySelected)
	m.copyBtn.Importance = widget.HighImportance
	m.countSeg = microText("", theme.ColorNamePlaceHolder, false)
	m.countSeg.Style.Alignment = fyne.TextAlignTrailing
	m.count = widget.NewRichText(m.countSeg)
	m.newBtn = widget.NewButton(labelNew, m.newSnippet)
	m.editBtn = widget.NewButton(labelEdit, m.editSelected)
	m.deleteBtn = widget.NewButton(labelDelete, m.deleteSelected)
	m.newBtn.Importance, m.editBtn.Importance, m.deleteBtn.Importance = widget.LowImportance, widget.LowImportance, widget.LowImportance
	m.status = widget.NewLabel("")
	m.status.Truncation = fyne.TextTruncateEllipsis
}

func (m *MainWindow) layout() fyne.CanvasObject {
	listArea := container.NewStack(m.list, container.NewCenter(m.emptyLabel))
	previewHead := container.NewVBox(widget.NewRichText(microText(previewLabel, theme.ColorNamePlaceHolder, true)), m.previewTitle)
	preview := container.NewBorder(previewHead, nil, nil, nil, container.NewVScroll(m.previewBody))
	split := container.NewVSplit(listArea, preview)
	split.Offset = 0.55
	brand := widget.NewRichText(microText(brandName, theme.ColorNameError, true), microText("  "+brandSub, theme.ColorNamePlaceHolder, false))
	header := strip(container.NewStack(brand, m.count))
	buttons := container.NewHBox(m.copyBtn, m.newBtn, m.editBtn, m.deleteBtn)
	top := container.NewVBox(header, edgeRule(), container.NewPadded(m.search))
	bottom := container.NewVBox(edgeRule(), buttons, m.status, strip(keyHints()))
	return container.NewBorder(top, bottom, nil, nil, split)
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
	if m.query == "" {
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
		m.previewTitle.SetText(sel.Title)
		m.previewBody.SetText(sel.Body)
	} else {
		m.previewTitle.SetText("")
		m.previewBody.SetText(previewEmptyText)
	}
	for _, b := range []*widget.Button{m.copyBtn, m.editBtn, m.deleteBtn} {
		if has {
			b.Enable()
		} else {
			b.Disable()
		}
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
	refs.title.SetText(s.Title)
	refs.excerpt.Text = snippets.Excerpt(s.Body, 80)
	refs.rich.Refresh()
	refs.bar.FillColor = color.Color(color.Transparent)
	if s.ID == m.selectedID {
		refs.bar.FillColor = bloodRed
	}
	refs.bar.Refresh()
}

func (m *MainWindow) onListSelected(i widget.ListItemID) {
	if m.syncing || i >= len(m.results) {
		return
	}
	m.selectedID = m.results[i].ID
	m.syncView()
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
