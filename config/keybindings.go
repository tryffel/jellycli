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

package config

import "github.com/gdamore/tcell"

var (
	KeyBinds = DefaultKeyBindings()
)

// GlobalBindings can have only one action since they override all others
type GlobalBindings struct {
	PlayPause  tcell.Key
	Stop       tcell.Key
	Next       tcell.Key
	Previous   tcell.Key
	Forward    tcell.Key
	Backward   tcell.Key
	VolumeUp   tcell.Key
	VolumeDown tcell.Key
	MuteUnmute tcell.Key
}

// NavigationBarBindings also override every other key
type NavigationBarBindings struct {
	Quit     tcell.Key
	Help     tcell.Key
	View     tcell.Key
	Search   tcell.Key
	Queue    tcell.Key
	History  tcell.Key
	Settings tcell.Key
	Dump     tcell.Key
}

// MovingBindings control moving cursor inside panel
type MovingBindings struct {
	Up       tcell.Key
	Down     tcell.Key
	Left     tcell.Key
	Right    tcell.Key
	UpAlt    tcell.Key
	DownAlt  tcell.Key
	LeftAlt  tcell.Key
	RightAlt tcell.Key
}

// PanelBindings moving between panels
type PanelBindings struct {
	MovingBindings
}

type KeyBindings struct {
	Global        GlobalBindings
	NavigationBar NavigationBarBindings
	Moving        MovingBindings
	Panel         PanelBindings
}

func DefaultKeyBindings() KeyBindings {
	k := KeyBindings{
		Global: GlobalBindings{
			PlayPause:  tcell.KeyF6,
			Stop:       tcell.KeyF5,
			Next:       tcell.KeyF7,
			Previous:   tcell.KeyF4,
			Forward:    0,
			Backward:   0,
			VolumeUp:   tcell.KeyF10,
			VolumeDown: tcell.KeyF9,
			MuteUnmute: tcell.KeyF11,
		},
		NavigationBar: NavigationBarBindings{
			Help:    tcell.KeyF1,
			Search:  tcell.KeyCtrlF,
			Queue:   tcell.KeyF2,
			History: tcell.KeyF3,
			Dump:    tcell.KeyCtrlW,
		},
		Moving: MovingBindings{
			Up:    tcell.KeyUp,
			Down:  tcell.KeyDown,
			Left:  tcell.KeyLeft,
			Right: tcell.KeyRight,
			//UpAlt:    tcell.key,
			//DownAlt:  0,
			//LeftAlt:  0,
			//RightAlt: 0,
		},
		Panel: PanelBindings{MovingBindings{
			Up:    tcell.KeyCtrlK,
			Down:  tcell.KeyCtrlJ,
			Left:  tcell.KeyCtrlH,
			Right: tcell.KeyCtrlL,
			//UpAlt:    0,
			//DownAlt:  0,
			//LeftAlt:  0,
			//RightAlt: 0,
		}},
	}
	return k
}
