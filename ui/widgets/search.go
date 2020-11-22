package widgets

import (
	"fmt"
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

// SearchTopList shows overall result of different keys, where user
// can click any key and see actual results.
type SearchTopList struct {
	*twidgets.Banner
	*previous

	name *cview.TextView

	list        *twidgets.ScrollList
	listFocused bool
	selectFunc  func(itemType models.ItemType)

	results map[models.ItemType]*searchListItem

	prevBtn  *button
	prevFunc func()
}

func NewSearchTopList() *SearchTopList {
	stp := &SearchTopList{
		Banner:      twidgets.NewBanner(),
		previous:    &previous{},
		name:        cview.NewTextView(),
		listFocused: false,
		selectFunc:  nil,
		prevBtn:     newButton("Back"),
		prevFunc:    nil,
		results:     map[models.ItemType]*searchListItem{},
	}

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
	selectables := []twidgets.Selectable{stp.prevBtn, stp.list}

	for _, v := range btns {
		v.SetBackgroundColor(config.Color.ButtonBackground)
		v.SetLabelColor(config.Color.ButtonLabel)
		v.SetBackgroundColorActivated(config.Color.ButtonBackgroundSelected)
		v.SetLabelColorActivated(config.Color.ButtonLabelSelected)
	}

	stp.prevBtn.SetSelectedFunc(stp.goBack)
	stp.Banner.Selectable = selectables

	stp.Grid.SetRows(1, 1, 1, 1, -1)
	stp.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	stp.Grid.SetMinSize(1, 6)
	stp.Grid.SetBackgroundColor(config.Color.Background)
	stp.name.SetBackgroundColor(config.Color.Background)
	stp.name.SetTextColor(config.Color.Text)
	stp.name.SetText("Search results")

	stp.list.Grid.SetColumns(1, -1)

	stp.Grid.AddItem(stp.prevBtn, 0, 0, 1, 1, 1, 5, false)
	stp.Grid.AddItem(stp.name, 0, 2, 2, 6, 1, 10, false)
	stp.Grid.AddItem(stp.list, 4, 0, 1, 8, 6, 20, false)
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
