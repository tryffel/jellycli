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
	controller interfaces.Player
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
func NewController(controller interfaces.Player) (c *MediaController, err error) {
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
