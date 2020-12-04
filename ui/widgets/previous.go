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

import "gitlab.com/tslocum/cview"

// Previous can give last primitive
type Previous interface {
	cview.Primitive
	// Back gives last primitive
	Back() Previous
	// SetLast sets last primitive
	SetLast(p Previous)

	//SetBackCallback sets callback function that gets called when user clicks 'back'
	SetBackCallback(cb func(p Previous))
}

type previous struct {
	last     Previous
	callback func(p Previous)
}

func (p *previous) Back() Previous {
	return p.last
}

func (p *previous) SetLast(primitive Previous) {
	p.last = primitive
}

func (p *previous) SetBackCallback(cb func(p Previous)) {
	p.callback = cb
}

// call back callback if it's set
func (p *previous) goBack() {
	if p.callback != nil && p.last != nil {
		p.callback(p.last)

	}

}
