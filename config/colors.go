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
	TextDisabled2        = tcell.Color247
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
	TextDisabled2            tcell.Color
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
		TextDisabled2:            TextDisabled2,
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
