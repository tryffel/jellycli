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
	"tryffel.net/pkg/jellycli/config"
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

	//■ ▶ ⏯

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

type status struct {
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

	buttons []*tview.Button

	progress progressBar

	controlsFgColor tcell.Color
	controlsBgColor tcell.Color

	detailsFgColor tcell.Color
	detailsBgColor tcell.Color

	detailsMainColor tcell.Color
	detailsDimColor  tcell.Color
	playing          bool
}

func newStatus() *status {
	s := &status{frame: tview.NewBox()}
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
	//s.frame.SetTitle("[red:]Status")

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

	s.buttons = []*tview.Button{
		s.btnPrevious, s.btnBackward, s.btnPlay, s.btnForward, s.btnNext,
	}

	for _, v := range s.buttons {
		v.SetBackgroundColor(s.controlsFgColor)
		v.SetLabelColor(s.controlsBgColor)
		v.SetBorder(false)
	}

	p := newProgressBar()
	p.value = 3
	s.progress = p

	return s
}

func (s *status) Draw(screen tcell.Screen) {
	s.frame.Draw(screen)
	x, y, w, _ := s.frame.GetInnerRect()

	p := components.NewProgressBar(60, 100)
	v := components.NewProgressBar(10, 100)
	t := p.Draw(30)

	tview.Print(screen, " "+components.SecToString(125)+" "+t+" "+components.SecToString(225)+" ", x+1, y-1, w, tview.AlignLeft, s.controlsFgColor)
	tview.Print(screen, "Volume "+v.Draw(65), x+75, y-1, w, tview.AlignLeft, s.controlsFgColor)

	btnY := y + 2
	btnX := x + 1

	for _, v := range s.buttons {
		v.SetRect(btnX, btnY, 3, 1)
		v.Draw(screen)
		btnX += 4
	}
	s.WriteStatus("Song", "Artist", "Album", 2018, screen, x+2, y)
}

func (s *status) GetRect() (int, int, int, int) {
	return s.frame.GetRect()
}

func (s *status) SetRect(x, y, width, height int) {
	s.frame.SetRect(x, y, width, height)
}

func (s *status) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return s.frame.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		//key := event.Key()

	})
}

func (s *status) Focus(delegate func(p tview.Primitive)) {
	s.frame.Focus(delegate)
}

func (s *status) Blur() {
	s.frame.Blur()
}

func (s *status) GetFocusable() tview.Focusable {
	return s.frame.GetFocusable()
}

func (s *status) WriteStatus(song, artist, album string, year int, screen tcell.Screen, x, y int) {
	xi := x
	w, _ := screen.Size()
	tview.Print(screen, effect(song, "b")+" - ", x, y, w, tview.AlignLeft, s.detailsMainColor)
	x += len(song) + 3
	tview.Print(screen, effect(artist, "b")+" ", x, y, w, tview.AlignLeft, s.detailsMainColor)
	x += len(artist) + 1
	x = xi + 4
	tview.Print(screen, album+" ", x, y+1, w, tview.AlignLeft, s.detailsDimColor)
	x += len(album) + 1
	tview.Print(screen, fmt.Sprintf("(%d)", year), x, y+1, w, tview.AlignLeft, s.detailsDimColor)
}

// Return cb func that includes name
func (s *status) namedCbFunc(name string) func() {
	return func() {
		s.buttonCb(name)
	}
}

// Button pressed cb with name of the button
func (s *status) buttonCb(name string) {
	logrus.Infof("Button %s was pressed!", name)
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
