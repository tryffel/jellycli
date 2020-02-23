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
