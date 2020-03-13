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

import (
	"github.com/gdamore/tcell"
	"tryffel.net/go/twidgets"
)

const (
	colorBackground      = tcell.Color234
	colorModalBackground = tcell.Color236
	colorText            = tcell.Color252
	colorShortcut        = tcell.Color214
	TextSecondary        = tcell.Color179
	TextDisabled         = tcell.Color241
)

var Color = defaultColors()

type AppColor struct {
	Background               tcell.Color
	Border                   tcell.Color
	BorderFocus              tcell.Color
	ButtonBackground         tcell.Color
	ButtonBackgroundSelected tcell.Color
	ButtonLabel              tcell.Color
	ButtonLabelSelected      tcell.Color
	Text                     tcell.Color
	TextSecondary            tcell.Color
	TextDisabled             tcell.Color
	BackgroundSelected       tcell.Color
	TextSelected             tcell.Color
	TextSongPlaying          tcell.Color
	NavBar                   ColorNavBar
	Status                   ColorStatus
	Modal                    ColorModal
}

func defaultColors() AppColor {
	return AppColor{
		Background:               colorBackground,
		Border:                   tcell.Color246,
		BorderFocus:              tcell.Color253,
		ButtonBackground:         tcell.Color241,
		ButtonBackgroundSelected: tcell.Color23,
		ButtonLabel:              tcell.Color254,
		ButtonLabelSelected:      tcell.Color253,
		Text:                     colorText,
		TextSecondary:            TextSecondary,
		TextDisabled:             TextDisabled,
		BackgroundSelected:       tcell.Color23,
		TextSelected:             tcell.Color252,
		TextSongPlaying:          colorShortcut,
		NavBar:                   defaultColorNavBar(),
		Status:                   defaultColorStatus(),
		Modal:                    defaultColorModal(),
	}
}

type ColorModal struct {
	Background tcell.Color
	Text       tcell.Color
	Headers    tcell.Color
}

func defaultColorModal() ColorModal {
	return ColorModal{
		Background: colorModalBackground,
		Text:       tcell.Color252,
		Headers:    tcell.Color228,
	}
}

type ColorNavBar struct {
	Background       tcell.Color
	Text             tcell.Color
	ButtonBackground tcell.Color
	Shortcut         tcell.Color
}

func (c *ColorNavBar) ToWidgetsNavBar() *twidgets.NavBarColors {
	return &twidgets.NavBarColors{
		Background:            c.Background,
		BackgroundFocus:       c.Background,
		ButtonBackground:      c.ButtonBackground,
		ButtonBackgroundFocus: c.ButtonBackground,
		Text:                  c.Text,
		TextFocus:             c.Text,
		Shortcut:              c.Shortcut,
		ShortcutFocus:         c.Shortcut,
	}
}

func defaultColorNavBar() ColorNavBar {
	return ColorNavBar{
		Background:       colorBackground,
		Text:             colorText,
		ButtonBackground: colorBackground,
		Shortcut:         colorShortcut,
	}
}

type ColorStatus struct {
	Background       tcell.Color
	Border           tcell.Color
	ProgressBar      tcell.Color
	Text             tcell.Color
	ButtonBackground tcell.Color
	ButtonLabel      tcell.Color
	Shortcuts        tcell.Color
	TextPrimary      tcell.Color
	TextSecondary    tcell.Color
}

func defaultColorStatus() ColorStatus {
	return ColorStatus{
		Background:       colorBackground,
		Border:           tcell.Color245,
		ProgressBar:      tcell.Color245,
		Text:             colorText,
		ButtonBackground: tcell.Color240,
		ButtonLabel:      colorText,
		Shortcuts:        colorShortcut,
		TextPrimary:      colorText,
		TextSecondary:    tcell.Color26,
	}
}
