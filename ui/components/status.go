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

package components

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"tryffel.net/pkg/jellycli/player"
)

const ProgressBarName = "progressbar"

type StatusBar struct {
	component
	playing       bool
	song          string
	artist        string
	album         string
	updatePending bool
	pauseFunc     func(bool)
	currentStatus string
}

func NewStatusBar(pauseFunc func(bool)) *StatusBar {
	p := &StatusBar{}
	p.name = ProgressBarName
	p.Title = "Status"
	p.Editable = false
	p.Frame = true
	p.Scaling = scalingMax
	p.SizeMin = Point{X: 30, Y: 2}
	p.SizeMax = Point{X: 60, Y: 3}
	p.initialized = true
	p.updatePending = true
	p.updateFunc = p.refresh
	p.pauseFunc = pauseFunc
	return p
}

func (p *StatusBar) AssignKeyBindings(gui *gocui.Gui) error {
	if err := gui.SetKeybinding("", gocui.KeySpace, gocui.ModNone, p.onSpace); err != nil {
		return err
	}
	return nil
}

func (p *StatusBar) onSpace(g *gocui.Gui, v *gocui.View) error {
	p.playing = !p.playing
	p.updatePending = true
	p.pause()
	return nil
}

func (p *StatusBar) refresh() {
	//p.view.Clear()

	//p.view.Write([]byte(status))
	//p.updatePending = false

}

func (p *StatusBar) pause() {
	if p.pauseFunc != nil {
		p.pauseFunc(!p.playing)
	}
}

func (p *StatusBar) Update(state *player.PlayingState) {
	logrus.Debug("Update progress bar view")
	p.view.Clear()
	status := ""
	if state.State == player.Play {
		status += "⏮ ⏸ ⏯ " + state.Song
	} else {
		status += "⏮ ▶ ⏯ " + state.Song
	}
	status += fmt.Sprintf("%d / %d sec", state.CurrentSongPast, state.CurrentSongDuration)
	p.view.Write([]byte(status))
	p.updatePending = false

	p.currentStatus = status

}
