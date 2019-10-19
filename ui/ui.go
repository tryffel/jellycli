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
	"github.com/sirupsen/logrus"
	"tryffel.net/pkg/jellycli/player"
	"tryffel.net/pkg/jellycli/task"
	"tryffel.net/pkg/jellycli/ui/controller"
	"tryffel.net/pkg/jellycli/ui/widgets"
)

type Gui struct {
	task    *task.Task
	window  widgets.Window
	player  *player.Player
	content *controller.Content
}

func NewUi(player *player.Player, content *controller.Content) *Gui {
	u := &Gui{
		task: &task.Task{},

		player:  player,
		content: content,
	}
	u.window = widgets.NewWindow(u)
	u.task.Name = "Gui"
	u.task.SetLoop(u.loop)
	return u
}

func (gui *Gui) Start() error {
	err := gui.task.Start()
	if err != nil {
		return err
	}
	return gui.window.Run()
}

func (gui *Gui) Stop() {
	gui.window.Stop()
	_ = gui.task.Stop()
}

func (gui *Gui) loop() {
	for true {
		select {
		case <-gui.task.StopChan():
			break
		case state := <-gui.player.StateChannel():
			logrus.Info(state)
		case <-gui.content.SearchCompleteChan():
			logrus.Info("Got some results yay")
		}
	}
}

func (gui *Gui) Control(state player.State, volume int) {
	action := player.Action{
		State:   state,
		Type:    0,
		Volume:  volume,
		Artist:  "",
		Album:   "",
		Song:    "",
		AudioId: "",
	}
	gui.player.ActionChannel() <- action
}

func (gui *Gui) Search(q string) {
	go gui.content.Search(q)
}
