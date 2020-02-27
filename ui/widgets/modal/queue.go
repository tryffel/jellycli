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
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
)

type QueueMode int

const (
	QueueModeQueue QueueMode = iota
	QueueModeHistory
)

// Queue show list of songs played or songs to be played
type Queue struct {
	text      *tview.TextView
	mode      QueueMode
	visible   bool
	closeFunc func()
}

func NewQueue(mode QueueMode) *Queue {
	q := &Queue{
		text: tview.NewTextView(),
		mode: mode,
	}
	if mode == QueueModeQueue {
		q.text.SetTitle("Queue")
	} else if mode == QueueModeHistory {
		q.text.SetTitle("History")
	}

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

func (q *Queue) SetData(items []*models.SongInfo, duration int) {
	var text string
	if q.mode == QueueModeQueue {
		text = fmt.Sprintf("Total duration: %s, songs: %d\n\n", util.SecToString(duration), len(items))
		for i, v := range items {
			text += fmt.Sprintf("%d. %s - %s (%s) %s\n",
				i+1, v.Name, v.Album, v.Artist, util.SecToString(v.Duration))
		}
	} else if q.mode == QueueModeHistory {
		text = fmt.Sprintf("Total history: %s, songs: %d\n\n", util.SecToString(duration), len(items))
		for i, item := range items {
			text += fmt.Sprintf("%d. %s - %s (%s) %s\n",
				i+1, item.Name, item.Album, item.Artist, util.SecToString(item.Duration))
		}
	}
	q.text.SetText(text)
}

func (q *Queue) SetDoneFunc(doneFunc func()) {
	q.closeFunc = doneFunc
}

func (q *Queue) View() tview.Primitive {
	a := &q
	return *a
}

func (q *Queue) SetVisible(visible bool) {
	q.visible = visible
}

func (q *Queue) Draw(screen tcell.Screen) {
	if q.visible {
		q.text.Draw(screen)
	}
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
		if q.visible && key == tcell.KeyEscape {
			q.closeFunc()
		}
	})
}

func (q *Queue) Focus(delegate func(p tview.Primitive)) {
	q.text.SetBorderColor(config.ColorBorderFocus)
	q.text.Focus(delegate)
}

func (q *Queue) Blur() {
	q.text.SetBorderColor(config.ColorBorder)
	q.text.Blur()
}

func (q *Queue) GetFocusable() tview.Focusable {
	return q.text.GetFocusable()
}
