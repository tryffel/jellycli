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
			Search:  tcell.KeyF2,
			Queue:   tcell.KeyF3,
			History: tcell.KeyF4,
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
