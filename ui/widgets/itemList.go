package widgets

import (
	"gitlab.com/tslocum/cview"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/twidgets"
)

type itemList struct {
	*twidgets.Banner
	*previous
	list        *twidgets.ScrollList
	listFocused bool

	description *cview.TextView
	prevBtn     *button

	prevFunc func()
}

func newItemList(listSelectfunc func(index int)) *itemList {
	itemList := &itemList{
		Banner:      twidgets.NewBanner(),
		previous:    &previous{},
		list:        twidgets.NewScrollList(listSelectfunc),
		description: cview.NewTextView(),
		prevBtn:     newButton("Back"),
		prevFunc:    nil,
	}

	itemList.SetBorder(true)
	itemList.SetBackgroundColor(config.Color.Background)
	itemList.Grid.SetBackgroundColor(config.Color.Background)
	itemList.list.SetBackgroundColor(config.Color.Background)

	itemList.description.SetDynamicColors(true)

	itemList.list.SetBorder(true)
	itemList.list.Grid.SetColumns(1, -1)

	itemList.SetBorder(true)
	itemList.prevBtn.SetSelectedFunc(itemList.goBack)
	return itemList
}

// init context menu list. Context menu list has to contain at least one item
// before calling this.
func (i *itemList) initContextMenuList() {
	i.list.ContextMenuList().SetBorder(true)
	i.list.ContextMenuList().SetBackgroundColor(config.Color.Background)
	i.list.ContextMenuList().SetBorderColor(config.Color.BorderFocus)
	i.list.ContextMenuList().SetSelectedBackgroundColor(config.Color.BackgroundSelected)
	i.list.ContextMenuList().SetMainTextColor(config.Color.Text)
	i.list.ContextMenuList().SetSelectedTextColor(config.Color.TextSelected)
}
