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
	"strings"
)

const ContentView = "content"

//Content Provides unified view for artist, playlists, albums and songs
type Content struct {
	component
	count    int
	playFunc func(int)
}

func NewContentView(playFunc func(int)) *Content {
	a := &Content{
		count: 0,
	}
	a.name = ContentView
	a.Title = "Browser"
	a.Editable = false
	a.Frame = true
	a.Scaling = scalingMax
	a.SizeMin = Point{X: 60, Y: 10}
	a.SizeMax = Point{X: 100, Y: 20}
	a.initialized = true
	a.Highlight = true
	a.playFunc = playFunc
	return a
}

func (c *Content) AssignKeyBindings(gui *gocui.Gui) error {
	if err := gui.SetKeybinding(c.name, gocui.KeyCtrlJ, gocui.ModNone, c.scrollDown); err != nil {
		return err
	}
	if err := gui.SetKeybinding(c.name, gocui.KeyCtrlJ, gocui.ModNone, c.scrollUp); err != nil {
		return err
	}
	if err := gui.SetKeybinding(c.name, gocui.MouseLeft, gocui.ModNone, c.activate); err != nil {
		return err
	}
	if err := gui.SetKeybinding(c.name, gocui.KeyEnter, gocui.ModNone, c.play); err != nil {
		return err
	}
	return nil
}

func (c *Content) scrollDown(gui *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()
	return v.SetCursor(cx, cy+1)
}

func (c *Content) scrollUp(gui *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()
	return v.SetCursor(cx, cy-1)
}

func (c *Content) activate(gui *gocui.Gui, v *gocui.View) error {
	_, err := gui.SetCurrentView(c.name)
	return err
}

func (c *Content) SetText(text []string) error {
	c.view.Clear()
	_, err := c.view.Write([]byte(strings.Join(text, "\n")))
	return err
}

func (c *Content) play(gui *gocui.Gui, v *gocui.View) error {
	_, index := v.Cursor()
	if c.playFunc != nil {
		c.playFunc(index)
	}

	return nil
}
