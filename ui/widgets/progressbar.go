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

const (
	fullBlock     = "█"
	threeQuarters = "▊"
	twoQuarters   = "▌"
	oneQuarter    = "▎"

	startChar = "┫"
	stopChar  = "┣"
	empty     = "╍"
)

type ProgressBar interface {
	SetWidth(w int)
	SetMaximum(m int)
	Draw(val int) string
}

type progressBar struct {
	maximumValue int
	width        int
	splits       int
}

//NewProgressBar creates new progress bar with given character width
func NewProgressBar(width int, maxValue int) ProgressBar {
	p := &progressBar{}
	p.SetWidth(width)
	p.SetMaximum(maxValue)
	return p
}

//SetWidth sets new total width for progress bar
func (p *progressBar) SetWidth(width int) {
	p.width = width - 2
	// Four splits per character
	p.splits = 4 * width
}

func (p *progressBar) SetMaximum(max int) {
	p.maximumValue = max
}

func (p *progressBar) Draw(currentValue int) string {
	text := startChar

	// Progress as percent
	progress := int(float32(currentValue) / float32(p.maximumValue) * 1000)
	var splits int
	if progress == 0 {
		splits = 0
	} else {
		splits = p.splits * progress / 1000
	}

	filled := 0
	blocks := 0

	if splits > 0 {
		blocks = splits / 4
		for i := 0; i < blocks; i++ {
			filled += 1
			text += fullBlock
		}
		switch splits % 4 {
		case 0:
			break
		case 1:
			filled += 1
			text += oneQuarter
		case 2:
			filled += 1
			text += twoQuarters
		case 3:
			filled += 1
			text += threeQuarters
		}
	}

	for i := 0; i < p.width-filled+2; i++ {
		text += empty
	}
	text += stopChar
	return text
}
