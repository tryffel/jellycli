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
	"github.com/sirupsen/logrus"
	"sync"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/controller"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/player"
	"tryffel.net/go/jellycli/util"
	"unicode/utf8"
)

const (
	btnPrevious = "|<<"
	btnPause    = " ||"
	btnStop     = "■"
	btnPlay     = ">"
	btnNext     = ">>|"
	btnForward  = ">>"
	btnBackward = " <<"
	btnQueue    = "☰"

	btnLessVolume = "\xF0\x9F\x94\x89"
	btnMoreVolume = ""

	btnStyleStart = "[white:red:b]"
	btnStyleStop  = "[-:-:-]"
)

func btn(button string) string {
	return btnStyleStart + button + btnStyleStop + " "
}

func effect(text string, e string) string {
	return fmt.Sprintf("[::%s]%s[::-]", e, text)
}

type Status struct {
	lock sync.RWMutex

	frame   *tview.Box
	layout  *tview.Grid
	details *tview.TextView

	btnPlay     *tview.Button
	btnPause    *tview.Button
	btnNext     *tview.Button
	btnPrevious *tview.Button
	btnForward  *tview.Button
	btnBackward *tview.Button
	btnStop     *tview.Button
	btnQueue    *tview.Button

	buttons   []*tview.Button
	shortCuts []string
	progress  ProgressBar
	volume    ProgressBar

	lastState player.State

	controlsFgColor tcell.Color
	controlsBgColor tcell.Color

	detailsFgColor tcell.Color
	detailsBgColor tcell.Color

	detailsMainColor tcell.Color
	detailsDimColor  tcell.Color

	state controller.Status
	song  *models.SongInfo

	actionCb func(state player.State, volume int)

	controller controller.MediaController
}

func newStatus(ctrl controller.MediaController) *Status {
	s := &Status{frame: tview.NewBox()}
	s.controller = ctrl

	s.controlsBgColor = config.ColorBackground
	s.controlsFgColor = config.ColorControls
	s.detailsMainColor = config.ColorPrimary
	s.detailsDimColor = config.ColorPrimaryDim
	s.detailsBgColor = config.ColorBackground

	s.frame.SetBackgroundColor(s.controlsBgColor)
	s.layout = tview.NewGrid()
	s.layout.SetBorder(true)

	s.layout.SetRows(-1, 1, 1)
	s.layout.SetColumns(-1, 5, -1)
	s.frame.SetBorder(true)
	s.frame.SetBorderAttributes(tcell.AttrBold)
	s.frame.SetBorderColor(s.controlsFgColor)

	s.details = tview.NewTextView()

	s.btnPlay = tview.NewButton(btnPlay)
	s.btnPlay.SetSelectedFunc(s.namedCbFunc(btnPlay))
	s.btnPause = tview.NewButton(btnPause)
	s.btnPause.SetSelectedFunc(s.namedCbFunc(btnPause))
	s.btnNext = tview.NewButton(btnNext)
	s.btnNext.SetSelectedFunc(s.namedCbFunc(btnNext))
	s.btnPrevious = tview.NewButton(btnPrevious)
	s.btnPrevious.SetSelectedFunc(s.namedCbFunc(btnPrevious))
	s.btnForward = tview.NewButton(btnForward)
	s.btnForward.SetSelectedFunc(s.namedCbFunc(btnForward))
	s.btnBackward = tview.NewButton(btnBackward)
	s.btnBackward.SetSelectedFunc(s.namedCbFunc(btnBackward))
	s.btnStop = tview.NewButton(btnStop)
	s.btnStop.SetSelectedFunc(s.namedCbFunc(btnStop))
	s.btnQueue = tview.NewButton(btnQueue)

	s.progress = NewProgressBar(40, 100)
	s.volume = NewProgressBar(10, 100)

	state := player.PlayingState{
		State:               player.Stop,
		PlayingType:         player.Playlist,
		Song:                "",
		Artist:              "",
		Album:               "",
		CurrentSongDuration: 0,
		CurrentSongPast:     0,
		PlaylistDuration:    0,
		PlaylistLeft:        0,
		Volume:              50,
	}

	s.state = controller.Status{
		PlayingState: state,
		Song:         nil,
		Album:        nil,
		Artist:       nil,
	}

	s.lastState = player.Stop

	s.buttons = []*tview.Button{
		s.btnPrevious, s.btnBackward, s.btnPlay, s.btnForward, s.btnNext,
	}

	s.shortCuts = []string{
		util.PackKeyBindingName(config.KeyBinds.Global.Previous, 5),
		util.PackKeyBindingName(config.KeyBinds.Global.Backward, 5),
		util.PackKeyBindingName(config.KeyBinds.Global.PlayPause, 5),
		util.PackKeyBindingName(config.KeyBinds.Global.Forward, 5),
		util.PackKeyBindingName(config.KeyBinds.Global.Next, 5),
	}

	for _, v := range s.buttons {
		v.SetBackgroundColor(s.controlsFgColor)
		v.SetLabelColor(s.controlsBgColor)
		v.SetBorder(false)
	}
	return s
}

func (s *Status) Draw(screen tcell.Screen) {
	// TODO: Make drawing responsive
	s.frame.Draw(screen)
	x, y, w, _ := s.frame.GetInnerRect()

	songPast := util.SecToString(s.state.CurrentSongPast)
	songPast = " " + songPast + " "
	songDuration := util.SecToString(s.state.CurrentSongDuration)
	songDuration = " " + songDuration + " "

	volume := " Volume " + s.volume.Draw(s.state.Volume)
	topRowFree := w - len(songPast) - len(songDuration) - utf8.RuneCountInString(volume) - 5

	s.progress.SetWidth(topRowFree * 10 / 11)

	s.lock.RLock()
	defer s.lock.RUnlock()

	progressBar := s.progress.Draw(s.state.CurrentSongPast)

	progress := songPast + progressBar + songDuration
	progressLen := utf8.RuneCountInString(progress)

	topX := x + 1

	tview.Print(screen, progress, topX, y-1, progressLen+5, tview.AlignLeft, config.ColorControls)
	topX += progressLen + progressLen/10
	tview.Print(screen, volume, topX, y-1, w, tview.AlignLeft, config.ColorControls)

	tview.Print(screen, util.PackKeyBindingName(config.KeyBinds.Global.VolumeDown, 5),
		topX+7, y, topX+16, tview.AlignLeft, config.ColorControls)
	tview.Print(screen, util.PackKeyBindingName(config.KeyBinds.Global.VolumeUp, 5),
		topX+18, y, topX+1, tview.AlignLeft, config.ColorControls)

	btnY := y + 1
	btnX := x + 1

	for i, v := range s.buttons {
		tview.Print(screen, s.shortCuts[i], btnX, btnY-1, 4, tview.AlignLeft, config.ColorControls)

		v.SetRect(btnX, btnY, 3, 1)
		v.Draw(screen)
		btnX += 5
	}
	s.WriteStatus(screen, x+27, y)
}

func (s *Status) GetRect() (int, int, int, int) {
	return s.frame.GetRect()
}

func (s *Status) SetRect(x, y, width, height int) {
	s.frame.SetRect(x, y, width, height)
}

func (s *Status) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return s.frame.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		//key := event.Key()

	})
}

func (s *Status) Focus(delegate func(p tview.Primitive)) {
	s.frame.Focus(delegate)
}

func (s *Status) Blur() {
	s.frame.Blur()
}

func (s *Status) GetFocusable() tview.Focusable {
	return s.frame.GetFocusable()
}

func (s *Status) WriteStatus(screen tcell.Screen, x, y int) {
	xi := x
	w, _ := screen.Size()
	tview.Print(screen, effect(s.state.PlayingState.Song, "b")+" - ", x, y, w, tview.AlignLeft, s.detailsMainColor)
	x += len(s.state.PlayingState.Song) + 3
	tview.Print(screen, effect(s.state.PlayingState.Artist, "b")+" ", x, y, w, tview.AlignLeft, s.detailsMainColor)
	x += len(s.state.PlayingState.Artist) + 1
	x = xi + 4
	tview.Print(screen, s.state.PlayingState.Album+" ", x, y+1, w, tview.AlignLeft, s.detailsDimColor)
	x += len(s.state.PlayingState.Album) + 1
	tview.Print(screen, fmt.Sprintf("(%d)", s.state.Year), x, y+1, w, tview.AlignLeft, s.detailsDimColor)
}

// Return cb func that includes name
func (s *Status) namedCbFunc(name string) func() {
	return func() {
		s.buttonCb(name)
	}
}

// Button pressed cb with name of the button
func (s *Status) buttonCb(name string) {
	logrus.Infof("Button %s was pressed!", name)
	if s.actionCb == nil {
		return
	}
	switch name {
	case btnPlay:
		s.actionCb(player.Play, -1)
	case btnPause:
		s.actionCb(player.Pause, -1)
	}
}

func (s *Status) UpdateState(state controller.Status, song *models.SongInfo) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.state = state
	s.song = song
	s.progress.SetMaximum(s.state.CurrentSongDuration)

	if s.lastState != state.State {
		s.DrawButtons()
		s.lastState = state.State
	}

}

func (s *Status) DrawButtons() {
	if s.state.State == player.Play {
		s.btnPlay.SetLabel(btnPause)
	} else if s.state.State == player.Pause {
		s.btnPlay.SetLabel(btnPlay)
	}
}
