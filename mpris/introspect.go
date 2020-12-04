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
	"github.com/godbus/dbus/introspect"
)

// IntrospectNode returns the root node of the library's introspection output.
func (m *MediaController) IntrospectNode() *introspect.Node {
	return &introspect.Node{
		Name: m.Name(),
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			introspect.Interface{
				Name: "org.mpris.MediaPlayer2",
				Properties: []introspect.Property{
					introspect.Property{
						Name:   "CanQuit",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanRaise",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "HasTrackList",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "Identity",
						Type:   "s",
						Access: "read",
					},
					introspect.Property{
						Name:   "SupportedUriSchemes",
						Type:   "as",
						Access: "read",
					},
					introspect.Property{
						Name:   "SupportedMimeTypes",
						Type:   "as",
						Access: "read",
					},
				},
				Methods: []introspect.Method{
					introspect.Method{
						Name: "Raise",
					},
					introspect.Method{
						Name: "Quit",
					},
				},
			},
			introspect.Interface{
				Name: "org.mpris.MediaPlayer2.Player",
				Properties: []introspect.Property{
					introspect.Property{
						Name:   "PlaybackStatus",
						Type:   "s",
						Access: "read",
					},
					introspect.Property{
						Name:   "LoopStatus",
						Type:   "s",
						Access: "readwrite",
					},
					introspect.Property{
						Name:   "Rate",
						Type:   "d",
						Access: "readwrite",
					},
					introspect.Property{
						Name:   "Shuffle",
						Type:   "b",
						Access: "readwrite",
					},
					introspect.Property{
						Name:   "Metadata",
						Type:   "a{sv}",
						Access: "read",
					},
					introspect.Property{
						Name:   "Volume",
						Type:   "d",
						Access: "readwrite",
					},
					introspect.Property{
						Name:   "Position",
						Type:   "x",
						Access: "read",
					},
					introspect.Property{
						Name:   "MinimumRate",
						Type:   "d",
						Access: "read",
					},
					introspect.Property{
						Name:   "MaximumRate",
						Type:   "d",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanGoNext",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanGoPrevious",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanPlay",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanSeek",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanControl",
						Type:   "b",
						Access: "read",
					},
				},
				Signals: []introspect.Signal{
					introspect.Signal{
						Name: "Seeked",
						Args: []introspect.Arg{
							introspect.Arg{
								Name: "Position",
								Type: "x",
							},
						},
					},
				},
				Methods: []introspect.Method{
					introspect.Method{
						Name: "Next",
					},
					introspect.Method{
						Name: "Previous",
					},
					introspect.Method{
						Name: "Pause",
					},
					introspect.Method{
						Name: "PlayPause",
					},
					introspect.Method{
						Name: "Stop",
					},
					introspect.Method{
						Name: "Play",
					},
					introspect.Method{
						Name: "Seek",
						Args: []introspect.Arg{
							introspect.Arg{
								Name:      "Offset",
								Type:      "x",
								Direction: "in",
							},
						},
					},
					introspect.Method{
						Name: "SetPosition",
						Args: []introspect.Arg{
							introspect.Arg{
								Name:      "TrackId",
								Type:      "o",
								Direction: "in",
							},
							introspect.Arg{
								Name:      "Position",
								Type:      "x",
								Direction: "in",
							},
						},
					},
				},
			},
			// TODO: This interface is not fully implemented.
			// introspect.Interface{
			// 	Name: "org.mpris.MediaPlayer2.TrackList",

			// },
		},
	}
}
