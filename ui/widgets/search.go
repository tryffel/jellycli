package widgets

import (
	"fmt"
	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
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
	item.SetBorder(false)
	item.SetBackgroundColor(config.Color.Background)
	item.SetBorderPadding(0, 0, 1, 1)
	item.SetTextColor(config.Color.Text)

	text := fmt.Sprintf("%s: total %d", name, len(items))
	item.TextView.SetText(text)
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

type searchBox struct {
	*cview.InputField
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

	return s
}

func (s *searchBox) Blur() {
	s.InputField.SetFieldBackgroundColor(config.Color.Background)
	s.InputField.SetFieldTextColor(config.Color.Text)
}

func (s *searchBox) Focus(delegate func(p cview.Primitive)) {
	s.InputField.SetFieldBackgroundColor(config.Color.BackgroundSelected)
	s.InputField.SetFieldTextColor(config.Color.TextSelected)
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

// SearchTopList shows overall result of different keys, where user
// can click any key and see actual results.
type SearchTopList struct {
	*twidgets.Banner
	*previous

	searchInput *searchBox

	list        *twidgets.ScrollList
	listFocused bool
	selectFunc  func(itemType models.ItemType)

	results map[models.ItemType]*searchListItem

	prevBtn  *button
	prevFunc func()
}

func NewSearchTopList(searchFunc func(string)) *SearchTopList {
	stp := &SearchTopList{
		Banner:      twidgets.NewBanner(),
		previous:    &previous{},
		listFocused: false,
		selectFunc:  nil,
		prevBtn:     newButton("Back"),
		prevFunc:    nil,
		results:     map[models.ItemType]*searchListItem{},
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

	stp.Grid.SetRows(1, 1, 1, -1)
	stp.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	stp.Grid.SetMinSize(1, 6)
	stp.Grid.SetBackgroundColor(config.Color.Background)
	stp.list.Grid.SetColumns(1, -1)

	stp.Grid.AddItem(stp.prevBtn, 0, 0, 1, 1, 1, 5, false)
	stp.Grid.AddItem(stp.searchInput, 1, 2, 2, 6, 1, 10, false)
	stp.Grid.AddItem(stp.list, 3, 0, 1, 8, 6, 20, false)
	stp.listFocused = false
	return stp
}

func (s *SearchTopList) selectItem(index int) {

}

func (s *SearchTopList) Clear() {
	if len(s.results) > 0 {
		s.results = map[models.ItemType]*searchListItem{}
	}
	s.list.Clear()
}

func (s *SearchTopList) addItems(itemType models.ItemType, items []models.Item) {
	if s.results[itemType] == nil {
		list := newSearchListItem(string(itemType), len(s.results), items)
		s.results[itemType] = list
		s.list.AddItem(list)
	} else {
		// TODO: replace existing
	}
}
