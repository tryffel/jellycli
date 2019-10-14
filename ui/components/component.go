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
	"errors"
	"fmt"
	"github.com/jroimartin/gocui"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// Component can be drawn on gui with given position
type Component interface {
	Draw(gui *gocui.Gui, rect Rectangle) (ScalingResult, error)
	CurrentSize() (int, int)
	MaxSize() (int, int)
	MinSize() (int, int)
	Name() string
	AssignKeyBindings(gui *gocui.Gui) error
}

type ScalingResult int

const (
	// Success
	ScalingSuccess ScalingResult = 0
	// Didn't use all available space
	ScaleFailedSmaller ScalingResult = 1
	// Component requires bigger space then given
	ScaledFailedBigger ScalingResult = 2
)

func (s *ScalingResult) Success() bool {
	return *s == ScalingSuccess
}

func (s *ScalingResult) Failed() bool {
	return !s.Success()
}

type scalePreference int

const (
	// scalingDisabled in disabling automatic scaling
	scalingDisabled scalePreference = 0
	// scalingMin prefers minimum set size
	scalingMin scalePreference = 1
	// scalingMax preferes maximum allowed size
	scalingMax scalePreference = 2
)

// Base component functionality. This doesn't implement Component interface and must be embedded
type component struct {
	Title       string
	Frame       bool
	Editable    bool
	Wrap        bool
	BgColor     gocui.Attribute
	FgColor     gocui.Attribute
	SelBgColor  gocui.Attribute
	SelFgColor  gocui.Attribute
	Highlight   bool
	SizeMin     Point
	SizeCurrent Point
	SizeMax     Point
	Scaling     scalePreference
	Rectangle
	name        string
	view        *gocui.View
	initialized bool
	updateFunc  func()
}

func (c *component) Apply(view *gocui.View) {
	view.Title = c.Title
	view.Frame = c.Frame
	view.Editable = c.Editable
	view.Wrap = c.Wrap
	view.SelBgColor = gocui.ColorGreen
	view.SelFgColor = gocui.ColorWhite

	view.BgColor = setColor(c.BgColor, gocui.ColorDefault)
	view.FgColor = setColor(c.FgColor, gocui.ColorDefault)
	view.SelBgColor = setColor(c.SelBgColor, gocui.ColorDefault)
	view.SelFgColor = setColor(c.SelFgColor, gocui.ColorDefault)

	view.Highlight = c.Highlight
}

func (c *component) CurrentSize() (int, int) {
	return c.SizeCurrent.List()
}

func (c *component) MaxSize() (int, int) {
	return c.SizeMax.List()
}

func (c *component) MinSize() (int, int) {
	return c.SizeMin.List()
}

func (c *component) Name() string {
	return c.name
}

func (c *component) Draw(gui *gocui.Gui, rect Rectangle) (ScalingResult, error) {
	x, y := rect.Size()
	var err error = nil
	if c.fitsToboundaries(x, y) {
		err = c.drawComponent(gui, rect)
	}
	if err != nil {
		return ScaledFailedBigger, err
	}
	if c.updateFunc != nil {
		c.updateFunc()
	}
	return ScalingSuccess, nil
}

// getSize returns size with given max size
// Maximum size is always hard size and no component can override it
func (c *component) getSize(maxX, maxY int) Point {
	if c.Scaling == scalingDisabled {
		p := Point{}
		p.X = min(c.SizeCurrent.X, maxX)
		p.Y = min(c.SizeCurrent.Y, maxY)
		return p
	}
	if c.Scaling == scalingMin {
		return c.SizeMin
	}
	if c.Scaling == scalingMax {
		p := Point{}
		p.X = min(c.SizeMax.X, maxX)
		p.Y = min(c.SizeMax.Y, maxY)
		return p
	} else {
		return Point{-1, -1}
	}
}

// outerCorner returns opposite corner for first corner and max size
func (c *component) outerCorner(x, y, maxX, maxY int) Point {
	p := Point{}
	size := c.getSize(maxX, maxY)
	p.X = x + size.X
	p.Y = y + size.Y
	return p
}

// check if component fits given coordinate boundaries
func (c *component) fitsToboundaries(maxX, maxY int) bool {
	maxSize := c.getSize(maxX, maxY)
	if maxSize.X > maxX || maxSize.Y > maxY {
		return false
	}
	return true
}

// Draw component to gui
func (c *component) drawComponent(gui *gocui.Gui, coords Rectangle) error {
	if !c.initialized {
		fmt.Printf("Error! Component '%s' not initialized properly!", c.name)
		return errors.New("component not initialized")
	}

	/* If gui size is 0 pixels, draw components as 0 as well. This is to just introduce component to gui even though
	it's not visible yet.
	*/
	var x, y int
	guiX, guiY := gui.Size()
	if guiX == 0 && guiY == 0 {
		coords.Set(0)
	} else {
		x, y := coords.Size()
		p := c.getSize(x, y)
		coords.Limit(p.X, p.Y)
		coords.Sanitize()
	}

	view, err := gui.SetView(c.name, coords.X0, coords.Y0, coords.X1, coords.Y1)
	if err != nil {
		return err
	}
	c.view = view
	x, y = coords.Size()
	c.SizeCurrent = Point{x, y}
	c.Rectangle = coords
	c.Apply(view)
	return err
}

func NewComponent(name string) component {
	return component{
		Title:    name,
		Frame:    false,
		Editable: false,
		Wrap:     true,
	}
}

func clearComponent(g *gocui.Gui, v *gocui.View) error {
	v.Clear()
	return nil
}

//Set non zero value or default value
func setColor(color gocui.Attribute, defaultVal gocui.Attribute) gocui.Attribute {
	if color != 0 {
		return color
	}
	return defaultVal
}
