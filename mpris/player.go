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
	"errors"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/prop"
	"github.com/sirupsen/logrus"
	"math"
	"time"
	"tryffel.net/go/jellycli/interfaces"
)

// This file implements a struct that satisfies the `org.mpris.MediaPlayer2.Player` interface.

// Player is a DBus object satisfying the `org.mpris.MediaPlayer2.Player` interface.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html
type Player struct {
	*MediaController
	lastState interfaces.AudioStatus
}

// TrackID is the Unique track identifier.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Simple-Type:Track_Id
type TrackID string

// PlaybackRate is a playback rate.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Simple-Type:Playback_Rate
type PlaybackRate float64

// TimeInUs is time in microseconds.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Simple-Type:Time_In_Us
type TimeInUs int64

// UsFromDuration returns the type from a time.Duration
func UsFromDuration(t time.Duration) TimeInUs {
	return TimeInUs(t / time.Microsecond)
}

// Duration returns the type in time.Duration
func (t TimeInUs) Duration() time.Duration { return time.Duration(t) * time.Microsecond }

// PlaybackStatus is a playback state.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Enum:Playback_Status
type PlaybackStatus string

// Defined PlaybackStatuses.
const (
	PlaybackStatusPlaying PlaybackStatus = "Playing"
	PlaybackStatusPaused  PlaybackStatus = "Paused"
	PlaybackStatusStopped PlaybackStatus = "Stopped"
)

// LoopStatus is a repeat / loop status.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Enum:Loop_Status
type LoopStatus = string

// Defined LoopStatuses
const (
	LoopStatusNone     LoopStatus = "None"
	LoopStatusTrack    LoopStatus = "Track"
	LoopStatusPlaylist LoopStatus = "Playlist"
)

//UpdateStatus updates status to dbus
func (p *Player) UpdateStatus(state interfaces.AudioStatus) {
	p.lastState = state
	var playStatus PlaybackStatus
	switch state.State {
	case interfaces.AudioStatePlaying:
		playStatus = PlaybackStatusPlaying
	case interfaces.AudioStatePaused:
		playStatus = PlaybackStatusPaused
	case interfaces.AudioStateStopped:
		playStatus = PlaybackStatusStopped
	}
	object := objectName("Player")

	var pos int64 = 0
	var data = MetadataMap{}

	if state.Song != nil {
		pos = int64(state.SongPast.MicroSeconds())
		data = mapFromStatus(state)

	}
	if err := p.props.Set(object, "Metadata", dbus.MakeVariant(data)); err != nil {
		logrus.Error(err)
		return
	}
	if err := p.props.Set(object, "Position", dbus.MakeVariant(pos)); err != nil {
		logrus.Error(err)
		return
	}

	if err := p.props.Set(object, "PlaybackStatus", dbus.MakeVariant(playStatus)); err != nil {
		logrus.Error(err)
		return
	}
}

func notImplemented(c *prop.Change) *dbus.Error {
	return dbus.MakeFailedError(errors.New("Not implemented"))
}

// OnLoopStatus handles LoopStatus change.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Property:LoopStatus
func (p *Player) OnLoopStatus(c *prop.Change) *dbus.Error {
	loop := LoopStatus(c.Value.(string))
	logrus.Debugf("LoopStatus changed to %v\n", loop)

	return nil
}

// OnVolume handles volume changes.
func (p *Player) OnVolume(c *prop.Change) *dbus.Error {
	val := int(c.Value.(float64) * 100)
	logrus.Debugf("Volume changed to %v\n", val)
	if val < 0 {
		val = 0
	}
	//return transform(p.mpd.SetVolume(val))
	volume := interfaces.AudioVolume(val)
	p.controller.SetVolume(volume)
	return nil
}

// OnShuffle handles Shuffle change.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Property:Shuffle
func (p *Player) OnShuffle(c *prop.Change) *dbus.Error {
	logrus.Debugf("Shuffle changed to %v\n", c.Value.(bool))
	//return transform(p.mpd.Random(c.Value.(bool)))
	return nil
}

func (p *Player) properties() map[string]*prop.Prop {
	return map[string]*prop.Prop{
		"PlaybackStatus": newProp(PlaybackStatusPlaying, true, true, nil),
		"LoopStatus":     newProp(LoopStatusTrack, true, true, p.OnLoopStatus),
		"Rate":           newProp(1.0, true, true, notImplemented),
		"Shuffle":        newProp(false, true, true, p.OnShuffle),
		"Metadata":       newProp(mapFromStatus(p.lastState), true, true, nil),
		"Volume":         newProp(math.Max(0, float64(80)/100.0), true, true, p.OnVolume),
		"Position": &prop.Prop{
			Value:    UsFromDuration(0),
			Writable: true,
			Emit:     prop.EmitTrue,
			Callback: nil,
		},
		"MinimumRate":   newProp(1.0, false, true, nil),
		"MaximumRate":   newProp(1.0, false, true, nil),
		"CanGoNext":     newProp(true, false, true, nil),
		"CanGoPrevious": newProp(true, false, true, nil),
		"CanPlay":       newProp(true, false, true, nil),
		"CanPause":      newProp(true, false, true, nil),
		"CanSeek":       newProp(false, true, true, nil),
		"CanControl":    newProp(true, false, true, nil),
	}
}

// Next skips to the next track in the tracklist.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Next
func (p *Player) Next() *dbus.Error {
	p.controller.Next()
	return nil
}

// Previous skips to the previous track in the tracklist.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Previous
func (p *Player) Previous() *dbus.Error {
	return nil
}

// Pause pauses playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Pause
func (p *Player) Pause() *dbus.Error {
	p.controller.Pause()
	return nil
}

// Play starts or resumes playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Play
func (p *Player) Play() *dbus.Error {
	p.controller.Continue()
	return nil
}

// Stop stops playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Stop
func (p *Player) Stop() *dbus.Error {
	p.controller.StopMedia()
	return nil
}

// PlayPause toggles playback.
// If playback is already paused, resumes playback.
// If playback is stopped, starts playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:PlayPause
func (p *Player) PlayPause() *dbus.Error {
	if p.lastState.State == interfaces.AudioStatePlaying {
		p.controller.Pause()
	} else if p.lastState.State == interfaces.AudioStatePaused {
		p.controller.Continue()
	}
	return nil
}

// Seek seeks forward in the current track by the specified number of microseconds.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Seek
func (p *Player) Seek(x TimeInUs) *dbus.Error {
	return nil
}

// SetPosition sets the current track position in microseconds.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:SetPosition
func (p *Player) SetPosition(o TrackID, x TimeInUs) *dbus.Error {
	return nil
}
