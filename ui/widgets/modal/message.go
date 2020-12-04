/*
 * Jellycli is a terminal music player for Jellyfin.
 * Copyright (C) 2020 Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package modal

import (
	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
	"tryffel.net/go/jellycli/config"
)

type Message struct {
	*cview.TextView
	visible bool
	closeCb func()

	okBtn *cview.Button
}

func NewMessage() *Message {
	m := &Message{
		TextView: cview.NewTextView(),
		visible:  false,
		closeCb:  nil,
		okBtn:    cview.NewButton("Close"),
	}

	colors := config.Color.Modal
	m.SetBackgroundColor(colors.Background)
	m.SetBorder(true)
	m.SetTitle("Info")
	m.SetBorderColor(config.Color.Border)
	m.SetTitleColor(config.Color.TextSecondary)
	m.SetTextColor(colors.Text)
	m.SetBorderPadding(0, 1, 2, 2)

	return m
}

func (m *Message) SetDoneFunc(doneFunc func()) {
	m.closeCb = doneFunc
	m.okBtn.SetSelectedFunc(doneFunc)
}

func (m *Message) View() cview.Primitive {
	return m
}

func (m *Message) SetVisible(visible bool) {
	m.visible = visible
}

func (m *Message) Focus(delegate func(p cview.Primitive)) {
	m.TextView.SetBorderColor(config.Color.BorderFocus)
	m.TextView.Focus(delegate)
}

func (m *Message) Blur() {
	m.TextView.SetBorderColor(config.Color.Border)
	m.TextView.Blur()
}

func (m *Message) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		key := event.Key()
		if key == tcell.KeyEscape {
			m.closeCb()
		}
		m.TextView.InputHandler()(event, setFocus)
	}
}
