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

package mpris

import (
	"fmt"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"github.com/godbus/dbus/prop"
	"os"
	"strings"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
)

const (
	basePath   = "/org/mpris/MediaPlayer2"
	baseObject = "org.mpris.MediaPlayer2"
)

func objectName(name string) string {
	return baseObject + "." + name
}

// MediaController manages connection to DBus.
// It contains a connection to the MPD server and the DBus connection.
type MediaController struct {
	dbus       *dbus.Conn
	props      *prop.Properties
	controller interfaces.MediaController
	name       string
}

// Close ends the connection.
func (m *MediaController) Close() error {
	return m.dbus.Close()
}

// Name returns the name of the instance.
func (m *MediaController) Name() string {
	return m.name
}

//NewController creates new Mpris controller and connects to DBus.
func NewController(controller interfaces.MediaController) (c *MediaController, err error) {
	c = &MediaController{
		name:       fmt.Sprintf("%s.%s.instance%d", baseObject, strings.ToLower(config.AppName), os.Getpid()),
		controller: controller,
	}
	if c.dbus, err = dbus.SessionBus(); err != nil {
		return nil, err
	}

	c.dbus.Export(c, basePath, baseObject)

	player := &Player{MediaController: c}
	c.dbus.Export(player, basePath, objectName("Player"))

	c.dbus.Export(introspect.NewIntrospectable(c.IntrospectNode()), basePath,
		"org.freedesktop.DBus.Introspectable")

	c.props = prop.New(c.dbus, basePath, map[string]map[string]*prop.Prop{
		baseObject:           c.properties(),
		objectName("Player"): player.properties(),
	})

	reply, err := c.dbus.RequestName(c.Name(), dbus.NameFlagReplaceExisting)

	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		return nil, fmt.Errorf("request dbus name: %v", err)
	}

	return
}

func (m *MediaController) properties() map[string]*prop.Prop {
	return map[string]*prop.Prop{
		"CanQuit":      newProp(false, false, true, nil),
		"CanRaise":     newProp(false, false, true, nil),
		"HasTrackList": newProp(false, false, true, nil),
		"Identity":     newProp(config.AppName, false, true, nil),
		// Empty because we can't add arbitary files in...
		"SupportedUriSchemes": newProp([]string{}, false, true, nil),
		"SupportedMimeTypes":  newProp([]string{}, false, true, nil),
	}
}

// Raise brings the media player's user interface to the front using any appropriate mechanism available.
// https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Method:Raise
func (m *MediaController) Raise() *dbus.Error { return nil }

// Quit causes the media player to stop running.
// https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Method:Quit
func (m *MediaController) Quit() *dbus.Error { return nil }

func newProp(value interface{}, write bool, emitValue bool, cb func(*prop.Change) *dbus.Error) *prop.Prop {
	var emitFlag prop.EmitType
	if emitValue {
		emitFlag = prop.EmitTrue
	} else {
		emitFlag = prop.EmitInvalidates
	}
	return &prop.Prop{
		Value:    value,
		Writable: write,
		Emit:     emitFlag,
		Callback: cb,
	}
}
