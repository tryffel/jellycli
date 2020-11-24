package widgets

import (
	"fmt"
	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
	"sync"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/twidgets"
)

// single item in SearchTopList. Item represents
// differen keys: album,artist,playlist etc.
type searchListItem struct {
	*cview.TextView
	items []models.Item
	index int
	name  string
}

func newSearchListItem(name string, index int, items []models.Item) *searchListItem {
	item := &searchListItem{
		TextView: cview.NewTextView(),
		items:    items,
		index:    index,
		name:     name,
	}
	item.TextView.SetDynamicColors(true)
	item.SetBorder(false)
	item.SetBackgroundColor(config.Color.Background)
	item.SetBorderPadding(0, 0, 1, 1)
	item.SetTextColor(config.Color.Text)
	return item
}

func (s *searchListItem) SetSelected(selected twidgets.Selection) {
	switch selected {
	case twidgets.Selected:
		s.SetBackgroundColor(config.Color.BackgroundSelected)
		s.SetTextColor(config.Color.TextSelected)
	case twidgets.Blurred:
		s.SetBackgroundColor(config.Color.TextDisabled)
	case twidgets.Deselected:
		s.SetBackgroundColor(config.Color.Background)
		s.SetTextColor(config.Color.Text)
	}
}

func (s *searchListItem) Draw(screen tcell.Screen) {
	_, _, _, h := s.GetRect()
	// title
	h -= 1

	var text string

	if h <= 1 {
		text = fmt.Sprintf("[yellow]%ss[-] (%d)\n", s.name, len(s.items))
	} else {
		text = fmt.Sprintf("[yellow]%ss[-]\n", s.name)
		for i, v := range s.items {
			if i > 0 {
				text += "\n"
			}
			if i > h-2 {
				text += fmt.Sprintf("%d more", len(s.items)-i)
				break
			}
			text += " * " + v.GetName()
		}
	}

	s.TextView.SetText(text)
	s.TextView.Draw(screen)
}

func (s *searchListItem) heightHint() int {
	return len(s.items) + 1
}

type searchBox struct {
	*cview.InputField
	lock  *sync.Mutex
	label string

	previousQuery string
	searchFunc    func(string)
	blurFunc      func(key tcell.Key)
}

func (s *searchBox) SetBlurFunc(f func(key tcell.Key)) {
	s.blurFunc = f
}

func newSearchBox(label string, searchFunc func(string)) *searchBox {
	s := &searchBox{
		InputField: cview.NewInputField(),
		label:      label,
		searchFunc: searchFunc,
	}

	colors := config.Color

	s.InputField.SetBackgroundColor(colors.Background)
	s.InputField.SetLabelColor(colors.TextSecondary)
	s.InputField.SetFieldTextColor(colors.Text)
	s.InputField.SetFieldBackgroundColor(colors.Background)
	s.InputField.SetPlaceholderTextColor(colors.TextDisabled)

	s.InputField.SetPlaceholder("John Cage")
	s.InputField.SetLabel(label)
	s.InputField.SetDoneFunc(s.done)
	s.InputField.SetInputCapture(s.inputCapture)
	return s
}

func (s *searchBox) Blur() {
	s.InputField.SetFieldBackgroundColor(config.Color.Background)
	s.InputField.SetFieldTextColor(config.Color.Text)
	s.InputField.SetPlaceholderTextColor(config.Color.TextDisabled)
}

func (s *searchBox) Focus(delegate func(p cview.Primitive)) {
	s.InputField.SetFieldBackgroundColor(config.Color.BackgroundSelected)
	s.InputField.SetFieldTextColor(config.Color.TextSelected)
	s.InputField.SetPlaceholderTextColor(config.Color.TextDisabled2)
	s.InputField.Focus(delegate)
}

func (s *searchBox) done(key tcell.Key) {
	if key == tcell.KeyEnter {
		query := s.InputField.GetText()
		s.previousQuery = query
		if s.blurFunc != nil {
			s.blurFunc(key)
		}
		if s.searchFunc != nil {
			s.searchFunc(query)
		}
	} else if key == tcell.KeyEsc {
		s.InputField.SetText(s.previousQuery)
	} else if key == tcell.KeyTab {
		if s.blurFunc != nil {
			s.blurFunc(key)
		}
	} else if key == tcell.KeyBacktab {
		if s.blurFunc != nil {
			s.blurFunc(key)
		}
	}
}

func (s *searchBox) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	key := event.Key()

	switch key {
	case tcell.KeyUp:
		return tcell.NewEventKey(tcell.KeyBacktab, event.Rune(), event.Modifiers())
	case tcell.KeyDown:
		return tcell.NewEventKey(tcell.KeyTab, event.Rune(), event.Modifiers())
	}
	return event
}

// SearchTopList shows overall result of different keys, where user
// can click any key and see actual results.
type SearchTopList struct {
	*twidgets.Banner
	*previous

	searchInput *searchBox

	list          *twidgets.ScrollList
	listFocused   bool
	selectFunc    func(itemType models.ItemType)
	showMediafunc func(itemType models.ItemType, items []models.Item, query string)

	results []*searchListItem

	prevBtn  *button
	prevFunc func()
}

func NewSearchTopList(searchFunc func(string), selectMediaFunc func(itemType models.ItemType, items []models.Item, query string)) *SearchTopList {
	stp := &SearchTopList{
		Banner:        twidgets.NewBanner(),
		previous:      &previous{},
		listFocused:   false,
		selectFunc:    nil,
		prevBtn:       newButton("Back"),
		prevFunc:      nil,
		results:       []*searchListItem{},
		showMediafunc: selectMediaFunc,
	}

	stp.searchInput = newSearchBox("Search: ", searchFunc)
	stp.list = twidgets.NewScrollList(stp.selectItem)

	stp.SetBorder(true)
	stp.SetBorderColor(config.Color.Border)
	stp.SetBackgroundColor(config.Color.Background)
	stp.list.SetBackgroundColor(config.Color.Background)
	stp.list.SetBorder(true)
	stp.list.SetBorderColor(config.Color.Border)
	stp.list.Grid.SetColumns(-1, 5)
	stp.SetBorderColor(config.Color.Border)

	btns := []*button{stp.prevBtn}
	selectables := []twidgets.Selectable{stp.prevBtn, stp.searchInput, stp.list}

	for _, v := range btns {
		v.SetBackgroundColor(config.Color.ButtonBackground)
		v.SetLabelColor(config.Color.ButtonLabel)
		v.SetBackgroundColorActivated(config.Color.ButtonBackgroundSelected)
		v.SetLabelColorActivated(config.Color.ButtonLabelSelected)
	}

	stp.prevBtn.SetSelectedFunc(stp.goBack)
	stp.Banner.Selectable = selectables
	stp.list.ItemHeight = 6

	stp.Grid.SetRows(1, 1, -1)
	stp.Grid.SetColumns(6, 4, -1)
	stp.Grid.SetMinSize(1, 6)
	stp.Grid.SetBackgroundColor(config.Color.Background)
	stp.list.Grid.SetColumns(1, -1)

	stp.Grid.AddItem(stp.prevBtn, 0, 0, 1, 1, 1, 6, false)
	stp.Grid.AddItem(stp.searchInput, 0, 2, 1, 1, 1, 15, false)
	stp.Grid.AddItem(stp.list, 2, 0, 1, 3, 6, 28, false)
	stp.listFocused = false
	return stp
}

func (s *SearchTopList) selectItem(index int) {
	if index >= len(s.results) {
		return
	}

	if s.showMediafunc == nil {
		return
	}

	items := s.results[index].items
	var itemType models.ItemType

	if len(items) == 0 {
		itemType = models.ItemType("")
	} else {
		itemType = items[0].GetType()
	}

	s.showMediafunc(itemType, items, s.searchInput.GetText())
}

func (s *SearchTopList) ClearResults() {
	if len(s.results) > 0 {
		s.results = []*searchListItem{}
	}

	s.list.Clear()
}

func (s *SearchTopList) Clear() {
	s.ClearResults()
	s.searchInput.SetText("")
}

func (s *SearchTopList) addItems(itemType models.ItemType, items []models.Item) {
	list := newSearchListItem(string(itemType), len(s.results), items)
	s.results = append(s.results, list)
	s.list.AddItem(list)
}

func (s *SearchTopList) SetRect(x, y, w, h int) {
	_, _, _, oldH := s.GetRect()
	s.Banner.SetRect(x, y, w, h)
	if oldH != h {
		s.ResultsReady()
	}
}

// ResultsReady sets results layout
func (s *SearchTopList) ResultsReady() {
	// set item height so that items fit the window.

	if len(s.results) == 0 {
		return
	}

	_, _, _, h := s.GetRect()
	minHeight := 2
	maxHeight := h/len(s.results) - 3

	avgSizeHint := 0
	maxSizeHint := 0
	minSizeHint := 100

	for _, v := range s.results {
		hint := v.heightHint()
		maxSizeHint = max(maxSizeHint, hint)
		minSizeHint = min(minSizeHint, hint+4)
	}

	avgSizeHint = (minSizeHint + maxSizeHint) / 2

	minHeight = max(minHeight, minSizeHint)
	maxHeight = min(maxHeight, maxSizeHint)

	itemHeight := limit(avgSizeHint, minHeight, maxHeight)
	s.list.ItemHeight = itemHeight
}
