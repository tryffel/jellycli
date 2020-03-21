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

package modal

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"tryffel.net/go/jellycli/config"
)

type Message struct {
	*tview.TextView
	visible bool
	closeCb func()

	okBtn *tview.Button
}

func NewMessage() *Message {
	m := &Message{
		TextView: tview.NewTextView(),
		visible:  false,
		closeCb:  nil,
		okBtn:    tview.NewButton("Close"),
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

func (m *Message) View() tview.Primitive {
	return m
}

func (m *Message) SetVisible(visible bool) {
	m.visible = visible
}

func (m *Message) Focus(delegate func(p tview.Primitive)) {
	m.TextView.SetBorderColor(config.Color.BorderFocus)
	m.TextView.Focus(delegate)
}

func (m *Message) Blur() {
	m.TextView.SetBorderColor(config.Color.Border)
	m.TextView.Blur()
}

func (m *Message) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()
		if key == tcell.KeyEscape {
			m.closeCb()
		}
		m.TextView.InputHandler()(event, setFocus)
	}
}
