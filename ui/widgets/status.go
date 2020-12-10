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

package widgets

import (
	"fmt"
	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
	"sync"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
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

	// yellow heart, utf8. Not visible on all editors.
	charFavorite = "💛"
	btnShuffle   = "Shuffle"

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

	frame   *cview.Box
	layout  *cview.Grid
	details *cview.TextView

	btnPlay     *cview.Button
	btnPause    *cview.Button
	btnNext     *cview.Button
	btnPrevious *cview.Button
	btnForward  *cview.Button
	btnBackward *cview.Button
	btnStop     *cview.Button
	btnShuffle  *cview.Button

	buttons   []*cview.Button
	shortCuts []string
	progress  ProgressBar
	volume    ProgressBar

	state interfaces.AudioStatus

	detailsMainColor tcell.Color

	song *models.SongInfo

	actionCb func(state interfaces.AudioStatus)

	player interfaces.Player
}

func (s *Status) MouseHandler() func(action cview.MouseAction, event *tcell.EventMouse, setFocus func(p cview.Primitive)) (consumed bool, capture cview.Primitive) {
	return s.layout.MouseHandler()
}

func newStatus(ctrl interfaces.Player) *Status {
	s := &Status{frame: cview.NewBox()}
	s.player = ctrl

	colors := config.Color.Status
	s.detailsMainColor = colors.Text

	s.frame.SetBackgroundColor(colors.Background)
	s.layout = cview.NewGrid()
	s.layout.SetBorder(true)

	s.layout.SetRows(-1, 1, 1)
	s.layout.SetColumns(-1, 5, -1)
	s.frame.SetBorder(true)
	s.frame.SetBorderAttributes(tcell.AttrBold)
	s.frame.SetBorderColor(colors.Border)

	s.details = cview.NewTextView()

	// button inputs are captured globally in widgets.Window.
	// Status panel buttons act as labels only.

	s.btnPlay = cview.NewButton(btnPlay)
	s.btnPause = cview.NewButton(btnPause)
	s.btnNext = cview.NewButton(btnNext)
	s.btnPrevious = cview.NewButton(btnPrevious)
	s.btnForward = cview.NewButton(btnForward)
	s.btnBackward = cview.NewButton(btnBackward)
	s.btnStop = cview.NewButton(btnStop)
	s.btnShuffle = cview.NewButton(btnShuffle)

	s.progress = NewProgressBar(40, 100)
	s.volume = NewProgressBar(10, 100)

	state := interfaces.AudioStatus{
		State:    interfaces.AudioStateStopped,
		Song:     nil,
		Artist:   nil,
		Album:    nil,
		SongPast: 0,
		Volume:   50,
		Muted:    false,
	}

	s.state = state

	s.buttons = []*cview.Button{
		s.btnPrevious, s.btnBackward, s.btnStop, s.btnPlay, s.btnForward, s.btnNext,
	}

	s.shortCuts = []string{
		util.PackKeyBindingName(config.KeyBinds.Global.Previous, 5),
		util.PackKeyBindingName(config.KeyBinds.Global.Backward, 5),
		util.PackKeyBindingName(config.KeyBinds.Global.Stop, 5),
		util.PackKeyBindingName(config.KeyBinds.Global.PlayPause, 5),
		util.PackKeyBindingName(config.KeyBinds.Global.Forward, 5),
		util.PackKeyBindingName(config.KeyBinds.Global.Next, 5),
		util.PackKeyBindingName(config.KeyBinds.Global.Shuffle, 10),
	}

	for _, v := range append(s.buttons, s.btnShuffle) {
		v.SetBackgroundColor(colors.ButtonBackground)
		v.SetLabelColor(colors.ButtonLabel)
	}

	s.btnShuffle.SetBackgroundColor(colors.Background)
	return s
}

func (s *Status) Draw(screen tcell.Screen) {
	s.frame.Draw(screen)
	x, y, w, _ := s.frame.GetInnerRect()

	songPast := util.SecToString(s.state.SongPast.Seconds())
	songPast = " " + songPast + " "
	var songDuration = " 0:00 "
	if s.state.Song != nil {
		songDuration = util.SecToString(s.state.Song.Duration)
		songDuration = " " + songDuration + " "
	}

	volume := " Volume " + s.volume.Draw(int(s.state.Volume))
	topRowFree := w - len(songPast) - len(songDuration) - utf8.RuneCountInString(volume) - 5

	showShuffleBtn := false
	showShuffleSmall := false
	if w > 100 {
		showShuffleBtn = true
		topRowFree -= 6
	} else if w > 60 {
		showShuffleSmall = true
		topRowFree -= 3
	}

	s.progress.SetWidth(topRowFree * 10 / 11)

	s.lock.RLock()
	defer s.lock.RUnlock()

	progressBar := s.progress.Draw(s.state.SongPast.Seconds())
	progress := songPast + progressBar + songDuration
	progressLen := utf8.RuneCountInString(progress)
	topX := x + 1
	colors := config.Color.Status

	cview.Print(screen, progress, topX, y-1, progressLen+5, cview.AlignLeft, colors.ProgressBar)
	topX += progressLen + progressLen/10

	volumeLen := utf8.RuneCountInString(volume)
	volumeX := x + w - volumeLen - 1
	if s.state.Muted {
		cview.Print(screen, volume, volumeX, y-1, w, cview.AlignLeft, colors.VolumeMuted)
	} else {
		cview.Print(screen, volume, volumeX, y-1, w, cview.AlignLeft, colors.ProgressBar)
	}

	cview.Print(screen, util.PackKeyBindingName(config.KeyBinds.Global.VolumeDown, 5),
		volumeX+7, y, topX+16, cview.AlignLeft, colors.Shortcuts)
	cview.Print(screen, util.PackKeyBindingName(config.KeyBinds.Global.VolumeUp, 5),
		volumeX+18, y, topX+1, cview.AlignLeft, colors.Shortcuts)

	btnY := y + 1
	btnX := x + 1

	if w > 40 {
		for i, v := range s.buttons {
			cview.Print(screen, s.shortCuts[i], btnX, btnY-1, 4, cview.AlignLeft, colors.Shortcuts)
			v.SetRect(btnX, btnY, 3, 1)
			v.Draw(screen)
			btnX += 5
		}
	}
	if showShuffleBtn {
		s.btnShuffle.SetLabel("Shuffle")
		shuffleX := x + w - volumeLen - 11
		// draw two empty characters around the button the separate it from status box.
		cview.Print(screen, "         ", shuffleX, btnY-2, 9, cview.AlignLeft, colors.Shortcuts)
		s.btnShuffle.SetRect(shuffleX+1, btnY-2, 7, 1)
		s.btnShuffle.Draw(screen)
	} else if showShuffleSmall {
		s.btnShuffle.SetLabel("S")
		shuffleX := x + w - volumeLen - 4
		// draw two empty characters around the button the separate it from status box.
		cview.Print(screen, "   ", shuffleX, btnY-2, 3, cview.AlignLeft, colors.Shortcuts)
		s.btnShuffle.SetRect(shuffleX+1, btnY-2, 1, 1)
		s.btnShuffle.Draw(screen)
	}
	s.WriteStatus(screen, x+30, y)
}

func (s *Status) GetRect() (int, int, int, int) {
	return s.frame.GetRect()
}

func (s *Status) SetRect(x, y, width, height int) {
	s.frame.SetRect(x, y, width, height)
}

func (s *Status) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return s.frame.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		//key := event.Key()

	})
}

func (s *Status) Focus(delegate func(p cview.Primitive)) {
	s.frame.Focus(delegate)
}

func (s *Status) Blur() {
	s.frame.Blur()
}

func (s *Status) GetFocusable() cview.Focusable {
	return s.frame.GetFocusable()
}

func (s *Status) WriteStatus(screen tcell.Screen, x, y int) {
	if s.state.State != interfaces.AudioStateStopped &&
		(s.state.Song != nil && s.state.Album != nil && s.state.Artist != nil) {
		xi := x
		x += 2
		w, _ := screen.Size()
		if s.state.Song.Favorite {
			cview.Print(screen, charFavorite, x, y, 2, cview.AlignLeft, config.Color.TextSelected)
			x += 3
		}

		cview.Print(screen, effect(s.state.Song.Name, "b")+" - ", x, y, w, cview.AlignLeft, s.detailsMainColor)
		x += len(s.state.Song.Name) + 3
		cview.Print(screen, effect(s.state.Artist.Name, "b")+" ", x, y, w, cview.AlignLeft, s.detailsMainColor)
		x += len(s.state.Artist.Name) + 1
		x = xi + 4
		cview.Print(screen, s.state.Album.Name+" ", x, y+1, w, cview.AlignLeft, s.detailsMainColor)
		x += len(s.state.Album.Name) + 1
		cview.Print(screen, fmt.Sprintf("(%d)", s.state.Album.Year), x, y+1, w, cview.AlignLeft, s.detailsMainColor)
	}
}

func (s *Status) UpdateState(state interfaces.AudioStatus, song *models.SongInfo) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.song = song
	if state.Song != nil {
		s.progress.SetMaximum(state.Song.Duration)
	}
	s.state = state
	s.DrawButtons()
}

func (s *Status) DrawButtons() {
	if s.state.Paused || s.state.State == interfaces.AudioStateStopped {
		s.btnPlay.SetLabel(btnPlay)
	} else {
		s.btnPlay.SetLabel(btnPause)
	}

	if s.state.Shuffle {
		s.btnShuffle.SetBackgroundColor(config.Color.BackgroundSelected)

	} else {
		s.btnShuffle.SetBackgroundColor(config.Color.Background)
	}
}
