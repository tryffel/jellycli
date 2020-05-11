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

import (
	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/twidgets"
)

// a button that implements widgets.Selectable
type button struct {
	*cview.Button
}

func (b *button) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		override := event
		key := event.Key()
		r := event.Rune()

		if key == tcell.KeyTAB || key == tcell.KeyDown || r == 'j' {
			override = tcell.NewEventKey(tcell.KeyTAB, 'j', tcell.ModNone)
		} else if key == tcell.KeyUp || r == 'k' {
			override = tcell.NewEventKey(tcell.KeyBacktab, 'k', tcell.ModShift)
		}

		if override == event {
			b.Button.InputHandler()(event, setFocus)
		} else {
			b.Button.InputHandler()(override, setFocus)
		}
	}
}

func (b *button) Focus(delegate func(p cview.Primitive)) {
	b.Button.Focus(delegate)
}

func (b *button) GetFocusable() cview.Focusable {
	return b.Button.GetFocusable()
}

func (b *button) SetBlurFunc(blur func(key tcell.Key)) {
	b.Button.SetBlurFunc(blur)
}

func newButton(label string) *button {
	return &button{
		Button: cview.NewButton(label),
	}
}

// a dropdown that implements widgets.Selectable
type dropDown struct {
	*cview.DropDown
	blurFunc   func(key tcell.Key)
	isOpen     bool
	isSelected bool
}

func (d *dropDown) SetBlurFunc(f func(key tcell.Key)) {
	d.blurFunc = f
}

func newDropDown(text string) *dropDown {
	d := &dropDown{
		DropDown: cview.NewDropDown(),
	}
	d.SetDoneFunc(d.done)
	d.SetInputCapture(d.inputCapture)

	d.SetLabelColor(config.Color.ButtonLabel)
	d.SetBackgroundColor(config.Color.ButtonBackground)
	d.SetFieldBackgroundColor(config.Color.Background)
	d.SetFieldTextColor(config.Color.Text)
	d.SetBorder(false)
	d.SetBorderPadding(0, 0, 1, 2)

	d.SetLabel(text)
	return d
}

func (d *dropDown) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		override := event
		key := event.Key()
		r := event.Rune()

		if key == tcell.KeyEnter {
			override = tcell.NewEventKey(tcell.KeyEnter, 'j', tcell.ModNone)
			d.isOpen = true

		} else if key == tcell.KeyTAB || key == tcell.KeyDown || r == 'j' {
			override = tcell.NewEventKey(tcell.KeyTAB, 'j', tcell.ModNone)
		} else if key == tcell.KeyUp || r == 'k' {
			override = tcell.NewEventKey(tcell.KeyBacktab, 'k', tcell.ModShift)
		}

		if override == event {
			d.DropDown.InputHandler()(event, setFocus)
		} else {
			d.DropDown.InputHandler()(override, setFocus)
		}
	}
}

func (d *dropDown) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	key := event.Key()
	r := event.Rune()
	d.isSelected = true

	if d.isOpen {
		if key == tcell.KeyDown || r == 'j' {
			return tcell.NewEventKey(tcell.KeyDown, 'j', tcell.ModNone)
		} else if key == tcell.KeyUp || r == 'k' {
			return tcell.NewEventKey(tcell.KeyUp, 'k', tcell.ModShift)
		}
	} else {
		if key == tcell.KeyEnter {
			d.isOpen = true
			return tcell.NewEventKey(tcell.KeyEnter, 'j', tcell.ModNone)
		} else if key == tcell.KeyTAB || key == tcell.KeyDown || r == 'j' {
			return tcell.NewEventKey(tcell.KeyTAB, 'j', tcell.ModNone)
		} else if key == tcell.KeyUp || r == 'k' {
			return tcell.NewEventKey(tcell.KeyBacktab, 'k', tcell.ModShift)
		}
	}
	return event
}

func (d *dropDown) done(key tcell.Key) {
	d.SetLabelColor(config.Color.ButtonLabel)
	d.SetBackgroundColor(config.Color.ButtonBackground)
	d.isOpen = false
	d.isSelected = false
	if d.blurFunc != nil {
		d.blurFunc(key)
	}
}

func (d *dropDown) Focus(delegate func(p cview.Primitive)) {
	d.SetLabelColor(config.Color.ButtonLabelSelected)
	d.SetBackgroundColor(config.Color.ButtonBackgroundSelected)
	d.DropDown.Focus(delegate)
	d.isSelected = true
}

func newScrollList(selectFunc func(index int)) *twidgets.ScrollList {
	s := twidgets.NewScrollList(selectFunc)
	s.SetBackgroundColor(config.Color.Background)
	s.SetBorder(true)
	s.SetBorderColor(config.Color.Border)
	return s
}
