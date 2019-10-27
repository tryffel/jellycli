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
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/controller"
	"tryffel.net/pkg/jellycli/models"
	"tryffel.net/pkg/jellycli/player"
	"tryffel.net/pkg/jellycli/ui/components"
)

const (
	btnPrevious = "|<<"
	btnPause    = "||"
	btnStop     = "■"
	btnPlay     = ">"
	btnNext     = ">>|"
	btnForward  = ">>"
	btnBackward = "<<"
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

	buttons  []*tview.Button
	progress components.ProgressBar
	volume   components.ProgressBar

	controlsFgColor tcell.Color
	controlsBgColor tcell.Color

	detailsFgColor tcell.Color
	detailsBgColor tcell.Color

	detailsMainColor tcell.Color
	detailsDimColor  tcell.Color

	state player.PlayingState
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

	s.progress = components.NewProgressBar(40, 100)
	s.volume = components.NewProgressBar(10, 100)

	s.state = player.PlayingState{
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

	s.buttons = []*tview.Button{
		s.btnPrevious, s.btnBackward, s.btnPlay, s.btnForward, s.btnNext,
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

	s.lock.RLock()
	defer s.lock.RUnlock()

	progress := fmt.Sprintf(" %s %s %s ",
		components.SecToString(s.state.CurrentSongPast), s.progress.Draw(s.state.CurrentSongPast),
		components.SecToString(s.state.CurrentSongDuration))
	volume := fmt.Sprintf(" Volume %s ", s.volume.Draw(s.state.Volume))

	tview.Print(screen, progress, x+1, y-1, w, tview.AlignLeft, config.ColorControls)
	tview.Print(screen, volume, x+75, y-1, w, tview.AlignLeft, config.ColorControls)

	btnY := y + 2
	btnX := x + 1

	for _, v := range s.buttons {
		v.SetRect(btnX, btnY, 3, 1)
		v.Draw(screen)
		btnX += 4
	}
	s.WriteStatus(screen, x+2, y)
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
	tview.Print(screen, effect(s.state.Song, "b")+" - ", x, y, w, tview.AlignLeft, s.detailsMainColor)
	x += len(s.state.Song) + 3
	tview.Print(screen, effect(s.state.Artist, "b")+" ", x, y, w, tview.AlignLeft, s.detailsMainColor)
	x += len(s.state.Artist) + 1
	x = xi + 4
	tview.Print(screen, s.state.Album+" ", x, y+1, w, tview.AlignLeft, s.detailsDimColor)
	x += len(s.state.Album) + 1
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

func (s *Status) UpdateState(state player.PlayingState, song *models.SongInfo) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.state = state
	s.song = song
	s.progress.SetMaximum(s.state.CurrentSongDuration)
}

func (s *Status) stateCb(state player.PlayingState) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.state = state
	s.progress.SetMaximum(s.state.CurrentSongDuration)
}

type progressBar struct {
	box   *tview.TextView
	bar   components.ProgressBar
	value int
}

func (p *progressBar) GetRect() (int, int, int, int) {
	return p.box.GetRect()
}

func (p *progressBar) SetRect(x, y, width, height int) {
	p.box.SetRect(x, y, width, height)
}

func (p *progressBar) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return p.box.InputHandler()
}

func (p *progressBar) Focus(delegate func(p tview.Primitive)) {
	p.box.Focus(delegate)
}

func (p *progressBar) Blur() {
	p.box.Blur()
}

func (p *progressBar) GetFocusable() tview.Focusable {
	return p.box.GetFocusable()

}

func newProgressBar() progressBar {
	p := progressBar{
		box: tview.NewTextView(),
		bar: components.NewProgressBar(10, 10),
	}
	return p
}

func (p *progressBar) Draw(screen tcell.Screen) {
	p.box.Draw(screen)
	x, y, w, _ := p.box.GetInnerRect()
	tview.Print(screen, p.bar.Draw(10), x+1, y+1, w, tview.AlignLeft, tcell.ColorGreen)

}
