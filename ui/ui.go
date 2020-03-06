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

package ui

import (
	"tryffel.net/go/jellycli/controller"
	"tryffel.net/go/jellycli/task"
	"tryffel.net/go/jellycli/ui/widgets"
)

type Gui struct {
	task.Task
	window     widgets.Window
	controller *controller.Content
}

func NewUi(controller *controller.Content) *Gui {
	u := &Gui{
		controller: controller,
	}
	u.window = widgets.NewWindow(controller)
	u.Name = "Gui"
	u.SetLoop(u.loop)
	return u
}

func (gui *Gui) Start() error {
	err := gui.Task.Start()
	if err != nil {
		return err
	}
	return gui.window.Run()
}

func (gui *Gui) Stop() error {
	gui.window.Stop()
	return gui.Task.Stop()
}

func (gui *Gui) loop() {
	gui.window.InitBrowser(gui.controller.GetDefault())

	for true {
		select {
		case <-gui.StopChan():
			break
		}
	}
}
