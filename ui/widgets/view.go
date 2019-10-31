/*
 * Copyright 2019 Tero Vierimaa
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package widgets

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/controller"
)

type ViewModal struct {
	list    *tview.List
	visible bool

	doneFunc func()
	viewFunc func(view controller.View)
}

func (v *ViewModal) SetDoneFunc(doneFunc func()) {
	v.doneFunc = doneFunc
}

func (v *ViewModal) Draw(screen tcell.Screen) {
	v.list.Draw(screen)
}

func (v *ViewModal) GetRect() (int, int, int, int) {
	return v.list.GetRect()
}

func (v *ViewModal) SetRect(x, y, width, height int) {
	v.list.SetRect(x, y, width, height)
}

func (v *ViewModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()
		if key == tcell.KeyEscape {
			v.doneFunc()
		} else {
			v.list.InputHandler()(event, setFocus)
		}
	}
}

func (v *ViewModal) Focus(delegate func(p tview.Primitive)) {
	v.list.SetBorderColor(config.ColorBorderFocus)
	v.list.Focus(delegate)
}

func (v *ViewModal) Blur() {
	v.list.SetBorderColor(config.ColorBorder)
	v.list.Blur()
}

func (v *ViewModal) GetFocusable() tview.Focusable {
	return v.list.GetFocusable()
}

func (v *ViewModal) View() tview.Primitive {
	return v
}

func (v *ViewModal) SetVisible(visible bool) {
	v.visible = visible
}

func (v *ViewModal) namedSelectFunc(view controller.View) func() {
	return func() {
		v.selectFunc(view)
	}
}

func (v *ViewModal) selectFunc(view controller.View) {
	if v.viewFunc != nil {
		v.viewFunc(view)
	}

	v.doneFunc()

}

func (v *ViewModal) SetViewFunc(viewFunc func(view controller.View)) {
	v.viewFunc = viewFunc
}

func NewViewModal() *ViewModal {
	v := &ViewModal{
		list: tview.NewList(),
	}

	v.list.SetBorder(true)
	v.list.SetBorderColor(config.ColorBorder)
	v.list.SetTitleColor(config.ColorPrimary)
	v.list.SetTitle("Go to")
	v.list.ShowSecondaryText(true)
	v.list.SetShortcutColor(config.ColorControls)
	v.list.SetBackgroundColor(config.ColorBackground)
	v.list.SetSelectedTextColor(config.ColorSecondary)
	v.list.SetMainTextColor(config.ColorPrimary)
	v.list.SetHighlightFullLine(true)
	v.list.SetBorderPadding(2, 2, 2, 2)

	v.list.AddItem("1. Latest Music", "", 0, v.namedSelectFunc(controller.ViewLatestMusic))
	v.list.AddItem("2. All Artists", "", 0, v.namedSelectFunc(controller.ViewAllArtists))
	v.list.AddItem("3. All Albums", "", 0, v.namedSelectFunc(controller.ViewAllAlbums))
	v.list.AddItem("4. All Songs", "", 0, v.namedSelectFunc(controller.ViewAllSongs))
	v.list.AddItem("5. Favorite Artists", "", 0, v.namedSelectFunc(controller.ViewFavoriteArtists))
	v.list.AddItem("6. Favorite Albums", "", 0, v.namedSelectFunc(controller.ViewFavoriteAlbums))
	v.list.AddItem("7. Favorite Songs", "", 0, v.namedSelectFunc(controller.ViewFavoriteSongs))
	v.list.AddItem("8. Playlists", "", 0, v.namedSelectFunc(controller.ViewPlaylists))

	return v
}
