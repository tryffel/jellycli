/*
 * Copyright 2020 Tero Vierimaa
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
