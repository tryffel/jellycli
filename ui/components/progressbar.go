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

package components

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
	progress := int(float32(currentValue) / float32(p.maximumValue) * 100)
	var splits int
	if progress == 0 {
		splits = 0
	} else {
		splits = p.splits * progress / 100
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
