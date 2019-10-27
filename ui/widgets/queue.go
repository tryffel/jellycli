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
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/models"
)

type Queue struct {
	text      *tview.TextView
	closeFunc func()
}

func NewQueue() *Queue {
	q := &Queue{
		text: tview.NewTextView(),
	}
	q.text.SetTitle("Queue")
	q.text.SetBackgroundColor(config.ColorBackground)
	q.text.SetBorderColor(config.ColorBorder)
	q.text.SetTextColor(config.ColorPrimary)
	q.text.SetTitleColor(config.ColorPrimary)
	q.text.SetBorderPadding(2, 2, 2, 2)
	q.text.SetWordWrap(true)
	q.text.SetScrollable(true)
	q.text.SetBorder(true)
	return q
}

func (q *Queue) setData(items []*models.SongInfo, duration int) {
	q.text.Clear()
	text := fmt.Sprintf("Total duration: %s, songs: %d\n\n", SecToString(duration), len(items))
	for i, v := range items {
		text += fmt.Sprintf("%d. %s - %s (%s) %s\n", i+1, v.Name, v.Album, v.Artist, SecToString(v.Duration))
	}
	q.text.SetText(text)
}

func (q *Queue) SetDoneFunc(doneFunc func()) {
	q.closeFunc = doneFunc
}

func (q *Queue) View() tview.Primitive {
	return q
}

func (q *Queue) Draw(screen tcell.Screen) {
	q.text.Draw(screen)
}

func (q *Queue) GetRect() (int, int, int, int) {
	return q.text.GetRect()
}

func (q *Queue) SetRect(x, y, width, height int) {
	q.text.SetRect(x, y, width, height)
}

func (q *Queue) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return q.text.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()
		if key == tcell.KeyEscape {
			q.closeFunc()
		}
	})
}

func (q *Queue) Focus(delegate func(p tview.Primitive)) {
	q.text.Focus(delegate)
}

func (q *Queue) Blur() {
	q.text.Blur()
}

func (q *Queue) GetFocusable() tview.Focusable {
	return q.text.GetFocusable()
}
