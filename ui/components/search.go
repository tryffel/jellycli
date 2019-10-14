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

import (
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
)

const SearchBarView = "searchbar"

var SearchBarSize = Point{12, 1}

type SearchBar struct {
	component
	lastQuery    string
	searchFunc   func(string) error
	previousView string
}

func (s *SearchBar) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	v.EditWrite(ch)
}

func NewSearchBar(searchFunc func(string) error) *SearchBar {
	s := &SearchBar{}
	s.name = SearchBarView
	s.SizeMin = Point{40, 2}
	s.SizeMax = Point{60, 2}
	s.SizeCurrent = s.SizeMax
	s.Title = "Search"
	s.Frame = true
	s.Editable = true
	s.Scaling = scalingMax
	s.initialized = true
	s.searchFunc = searchFunc
	s.Highlight = true
	s.SelFgColor = gocui.ColorYellow
	s.SelBgColor = gocui.ColorDefault
	return s
}

func (s *SearchBar) AssignKeyBindings(g *gocui.Gui) error {
	if err := g.SetKeybinding(s.name, gocui.KeyEnter, gocui.ModNone, s.onEnter); err != nil {
		return err
	}
	if err := g.SetKeybinding(s.name, gocui.MouseLeft, gocui.ModNone, s.activate); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlF, gocui.ModNone, s.activate); err != nil {
		return err
	}
	if err := g.SetKeybinding(s.name, gocui.KeyEsc, gocui.ModNone, s.onEscape); err != nil {
		return err
	}

	return nil
}

func (s *SearchBar) activate(g *gocui.Gui, v *gocui.View) error {
	view := g.CurrentView()
	if view != nil {
		s.previousView = view.Name()
	}
	s.Highlight = true

	g.SetCurrentView(s.name)
	//s.view.SetCursor(0,0)
	return nil
}

func (s *SearchBar) onEnter(g *gocui.Gui, v *gocui.View) error {
	clause, _ := s.view.Line(0)
	err := s.searchFunc(clause)
	if err != nil {
		s.view.Clear()
		logrus.Error("Error in search: ", err)
	}

	s.Highlight = false

	if s.previousView != "" {
		_, err = g.SetCurrentView(s.previousView)
		return err
	}
	return nil
}

func (s *SearchBar) onEscape(g *gocui.Gui, v *gocui.View) error {
	s.Highlight = false

	if s.previousView != "" {
		_, err := g.SetCurrentView(s.previousView)
		return err
	}
	return nil
}
