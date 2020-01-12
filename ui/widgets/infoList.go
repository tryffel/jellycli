/*
 * Copyright 2020 Tero Vierimaa
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
	"github.com/rivo/tview"
	"tryffel.net/pkg/jellycli/config"
)

//InfoList contains two widgets: text mode info on top and below that there's
// list primitive
type InfoList struct {
	*tview.Grid
	info *tview.TextView
	list tview.Primitive
}

//NewInfoList creates new infolist with given ListView.
func NewInfoList(listView tview.Primitive) *InfoList {
	i := &InfoList{
		Grid: tview.NewGrid(),
		info: tview.NewTextView(),
		list: listView,
	}

	if i.list == nil {
		i.list = tview.NewList()
	}
	i.info.SetBorder(true)
	i.info.SetBorderColor(config.ColorBorder)

	config.DebugGridBorders(i.Grid)
	i.SetBackgroundColor(config.ColorBackground)
	i.SetBorder(true)
	i.SetBorderColor(config.ColorBorder)
	i.SetTitle("Artist")
	i.SetTitleColor(config.ColorBorder)

	i.Grid.SetRows(-1, -3)
	i.Grid.SetColumns(-1, -1)

	i.Grid.AddItem(i.info, 0, 0, 1, 2, 4, 15, false)
	i.Grid.AddItem(i.list, 1, 0, 1, 2, 4, 15, true)
	return i
}

func (i *InfoList) SetInfoText(text string) {
	i.info.SetText(text)
	i.info.Highlight()
}
