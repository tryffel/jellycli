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
	"tryffel.net/pkg/jellycli/player"
)

const ProgressBarName = "progressbar"

type StatusBar struct {
	component
	playing       player.State
	volume        int
	song          string
	artist        string
	album         string
	updatePending bool
	ctrlFunc      func(state player.State, volume int)
	currentStatus string
	progressBar   *progressBar
	volumeBar     *progressBar
}

func NewStatusBar(ctrlFunc func(state player.State, volume int)) *StatusBar {
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
	p.ctrlFunc = ctrlFunc

	p.progressBar = newProgressBar(15, 100)
	p.volumeBar = newProgressBar(10, 100)
	p.volume = 50
	p.playing = player.Stop
	return p
}

func (p *StatusBar) AssignKeyBindings(gui *gocui.Gui) error {
	if err := gui.SetKeybinding("", gocui.KeyCtrlSpace, gocui.ModNone, p.onSpace); err != nil {
		return err
	}

	if err := gui.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, p.onPlus); err != nil {
		return err
	}
	if err := gui.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, p.onMinus); err != nil {
		return err
	}
	return nil
}

func (p *StatusBar) onSpace(g *gocui.Gui, v *gocui.View) error {
	p.updatePending = true
	p.pause()
	return nil
}

func (p *StatusBar) onPlus(g *gocui.Gui, v *gocui.View) error {
	p.setVolume(5)
	return nil
}

func (p *StatusBar) onMinus(g *gocui.Gui, v *gocui.View) error {
	p.setVolume(-5)
	return nil
}

func (p *StatusBar) refresh() {
	//p.view.Clear()

	//p.view.Write([]byte(status))
	//p.updatePending = false

}

func (p *StatusBar) pause() {
	if p.playing == player.Play {
		p.ctrlFunc(player.Pause, p.volume)
	} else if p.playing == player.Pause || p.playing == player.Stop {
		p.ctrlFunc(player.Play, p.volume)
	}
}

func (p *StatusBar) setVolume(amount int) {
	volume := p.volume + amount
	p.ctrlFunc(p.playing, volume)
}

func (p *StatusBar) Update(state *player.PlayingState) {
	p.view.Clear()
	p.playing = state.State
	p.volume = state.Volume
	if state.State == player.Play {
		p.progressBar.SetMaximum(state.CurrentSongDuration)
	}

	status := ""
	status += SecToString(state.CurrentSongPast) + " "
	status += p.progressBar.Draw(state.CurrentSongPast)
	status += " "
	status += SecToString(state.CurrentSongDuration)

	status += " Volume: " + p.volumeBar.Draw(state.Volume)

	status += "\n          "
	if state.State == player.Play {
		status += "⏮ ■ ⏸ ⏯ " + state.Song
	} else {
		status += "⏮ ■ ▶ ⏯ " + state.Song
	}

	p.view.Write([]byte(status))
	p.updatePending = false

	p.currentStatus = status
}

// Print seconds as formatted time:
// 50, 1:50,
// 0:05, 1.05, 1:05:05
func SecToString(sec int) string {
	if sec < 60 {
		return fmt.Sprintf("0:%02d", sec)
	}
	minutes := sec / 60
	if sec < 3600 {
		return fmt.Sprintf("%d:%02d", minutes, sec%60)
	} else {
		hours := sec / 3600
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes-60*hours, sec%3600%60)
	}
}
