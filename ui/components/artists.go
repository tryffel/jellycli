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

import "github.com/jroimartin/gocui"

const ArtistsView = "artists"

type Artists struct {
	component
	count int
}

func NewArtistsView() *Artists {
	a := &Artists{
		count: 0,
	}
	a.name = ArtistsView
	a.Title = "Artists"
	a.Editable = false
	a.Frame = true
	a.Scaling = scalingMax
	a.SizeMin = Point{X: 40, Y: 10}
	a.SizeMax = Point{X: 60, Y: 20}
	a.initialized = true
	return a
}

func (a *Artists) AssignKeyBindings(gui *gocui.Gui) error {
	return nil
}
