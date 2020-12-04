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

// Package ui implements graphical user interface in terminal. It uses widgets layout.
package ui

import (
	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
	"tryffel.net/go/jellycli/config"
	player2 "tryffel.net/go/jellycli/player"
	"tryffel.net/go/jellycli/task"
	"tryffel.net/go/jellycli/ui/widgets"
)

type Gui struct {
	task.Task
	window widgets.Window
	player *player2.Player
}

func NewUi(player *player2.Player) *Gui {
	u := &Gui{
		player: player,
	}
	bindDefaultTheme()
	u.window = widgets.NewWindow(player, player, player)
	u.Name = "Gui"
	u.SetLoop(u.loop)
	return u
}

func (gui *Gui) Start() error {
	err := gui.Task.Start()
	if err != nil {
		return err
	}
	return gui.window.Run()
}

func (gui *Gui) Stop() error {
	gui.window.Stop()
	return gui.Task.Stop()
}

func (gui *Gui) loop() {
	//gui.window.InitBrowser(gui.controller.GetDefault())

	for true {
		select {
		case <-gui.StopChan():
			break
		}
	}
}

func bindDefaultTheme() {

	colors := config.Color

	theme := cview.Theme{
		TitleColor:                  tcell.ColorWhite,
		BorderColor:                 colors.Border,
		GraphicsColor:               tcell.ColorWhite,
		PrimaryTextColor:            colors.Text,
		SecondaryTextColor:          colors.TextSecondary,
		TertiaryTextColor:           colors.Modal.Text,
		InverseTextColor:            0,
		ContrastSecondaryTextColor:  0,
		PrimitiveBackgroundColor:    colors.Background,
		ContrastBackgroundColor:     colors.Background,
		MoreContrastBackgroundColor: colors.Status.ProgressBar,
		ContextMenuPaddingTop:       0,
		ContextMenuPaddingBottom:    0,
		ContextMenuPaddingLeft:      1,
		ContextMenuPaddingRight:     1,

		ScrollBarColor: tcell.ColorWhite,
	}

	cview.Styles = theme
}
