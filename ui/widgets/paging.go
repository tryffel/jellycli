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
	"fmt"
	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
	"tryffel.net/go/jellycli/config"
)

// PageSelector shows current page and buttons for next and previous page. SelectFunc can be nil,
// in which case buttons do nothing.
type PageSelector struct {
	*cview.Box
	Next       *button
	Previous   *button
	PageNum    int
	TotalPages int

	SelectFunc func(page int)
	visible    bool
}

func NewPageSelector(selectPage func(int)) *PageSelector {
	p := &PageSelector{
		Box:        cview.NewBox(),
		Next:       newButton(" > "),
		Previous:   newButton(" < "),
		SelectFunc: selectPage,
	}

	p.Box.SetBackgroundColor(config.Color.Background)
	p.PageNum = 1
	p.Next.SetSelectedFunc(p.next)
	p.Previous.SetSelectedFunc(p.previous)
	return p
}

// SetPage sets current page
func (p *PageSelector) SetPage(n int) {
	p.PageNum = n
}

// SetTotalPages sets number of pages
func (p *PageSelector) SetTotalPages(n int) {
	p.TotalPages = n
}

func (p *PageSelector) next() {
	if p.PageNum < p.TotalPages-1 && p.SelectFunc != nil {
		p.SelectFunc(p.PageNum + 1)
	}
}

func (p *PageSelector) previous() {
	if p.PageNum > 0 && p.SelectFunc != nil {
		p.SelectFunc(p.PageNum - 1)
	}
}

func (p *PageSelector) Draw(screen tcell.Screen) {
	p.Box.Draw(screen)
	if p.visible {
		x, y, _, _ := p.GetRect()
		p.Next.Draw(screen)

		cview.Print(screen, fmt.Sprintf("%d / %d", p.PageNum+1, p.TotalPages),
			x+4, y, 9, cview.AlignCenter, config.Color.Text)
		p.Previous.Draw(screen)
	}
}

func (p *PageSelector) SetRect(x, y, width, height int) {
	if height < 1 || width < 18 {
		p.visible = false
	} else {
		p.visible = true
		p.Previous.SetRect(x, y, 3, 1)
		p.Next.SetRect(x+14, y, 3, 1)
	}
	p.Box.SetRect(x, y, width, height)
}
