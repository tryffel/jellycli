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

package player

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/sirupsen/logrus"
	"time"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
)

type audioFormat string

const (
	audioFormatMp3  audioFormat = "mp3"
	audioFormatFlac audioFormat = "flac"
)

// Audio manages playing song and implements interfaces.Player
type Audio struct {
	status interfaces.AudioStatus

	// todo: we need multiple streamers to allow seamlessly running next song
	streamer beep.StreamSeekCloser

	// ctrl allows pause
	ctrl *beep.Ctrl
	// volume
	volume *effects.Volume
	// mixer allows adding multiple streams sequentially
	mixer *beep.Mixer

	songCompleteFunc func()

	statusCallbacks []func(status interfaces.AudioStatus)
}

// initialize new player. This also initializes faiface.Speaker, which should be initialized only once.
func newAudio() *Audio {
	a := &Audio{
		ctrl: &beep.Ctrl{
			Streamer: nil,
			Paused:   false,
		},
		volume: &effects.Volume{
			Streamer: nil,
			Base:     config.AudioVolumeLogBase,
			Volume:   (config.AudioMinVolumedB + config.AudioMaxVolumedB) / 2,
			Silent:   false,
		},
		mixer:           &beep.Mixer{},
		statusCallbacks: make([]func(status interfaces.AudioStatus), 0),
	}
	a.ctrl.Streamer = a.mixer
	a.ctrl.Paused = false
	a.volume.Streamer = a.ctrl
	a.volume.Silent = false
	a.status.Volume = 50
	return a
}

func initAudio() error {
	err := speaker.Init(config.AudioSamplingRate, config.AudioSamplingRate/1000*
		int(config.AudioBufferPeriod.Seconds()*1000))
	if err != nil {
		return fmt.Errorf("init speaker: %v", err)
	}
	return nil
}

func (a *Audio) getStatus() interfaces.AudioStatus {
	speaker.Lock()
	defer speaker.Unlock()
	return a.status
}

// PlayPause toggles pause.
func (a *Audio) PlayPause() {
	speaker.Lock()
	if a.ctrl == nil {
		return
	}
	state := !a.ctrl.Paused
	if state {
		logrus.Info("Pause")
	} else {
		logrus.Info("Continue")
	}
	a.ctrl.Paused = state
	a.status.Paused = state
	a.status.Action = interfaces.AudioActionPlayPause
	speaker.Unlock()
	go a.flushStatus()
}

// Pause pauses audio. If audio is already paused, do nothing.
func (a *Audio) Pause() {
	logrus.Info("Pause audio")
	speaker.Lock()
	if a.ctrl == nil {
		return
	}
	a.ctrl.Paused = true
	a.status.Paused = true
	a.status.Action = interfaces.AudioActionPlayPause
	speaker.Unlock()
	go a.flushStatus()
}

// Continue continues paused audio. If audio is already playing, do nothing.
func (a *Audio) Continue() {
	logrus.Info("Continue audio")
	speaker.Lock()
	if a.ctrl == nil {
		return
	}
	a.ctrl.Paused = false
	a.status.Paused = false
	a.status.Action = interfaces.AudioActionPlayPause
	speaker.Unlock()
	go a.flushStatus()
}

// StopMedia stops music. If there is no audio to play, do nothing.
func (a *Audio) StopMedia() {
	speaker.Lock()
	a.status.State = interfaces.AudioStateStopped
	a.status.Action = interfaces.AudioActionStop
	speaker.Unlock()
	speaker.Clear()

	speaker.Lock()
	err := a.closeOldStream()
	speaker.Unlock()
	if err != nil {
		logrus.Errorf("stop: %v", err)
	}
	go a.flushStatus()
}

// Next plays next track. If there's no next song to play, do nothing.
func (a *Audio) Next() {
	speaker.Lock()
	a.status.Action = interfaces.AudioActionNext
	speaker.Unlock()
	go a.flushStatus()
}

// Previous plays previous track. If previous track does not exist, do nothing.
func (a *Audio) Previous() {
	speaker.Lock()
	a.status.Action = interfaces.AudioActionPrevious
	speaker.Unlock()
	go a.flushStatus()
}

// Seek seeks given ticks. If there is no audio, do nothing.
func (a *Audio) Seek(ticks interfaces.AudioTick) {
}

// AddStatusCallback adds a callback that gets called every time audio status is changed, or after certain time.
func (a *Audio) AddStatusCallback(cb func(status interfaces.AudioStatus)) {
	a.statusCallbacks = append(a.statusCallbacks, cb)
}

// SetVolume sets volume to given level.
func (a *Audio) SetVolume(volume interfaces.AudioVolume) {
	decibels := float64(volumeTodB(int(volume)))
	logrus.Debugf("Set volume to %d %s -> %.2f Db", volume, "%", decibels)
	speaker.Lock()

	// settings volume to 0 does not mute audio, set silent to true
	if decibels <= config.AudioMinVolumedB {
		a.volume.Silent = true
		a.volume.Volume = config.AudioMinVolumedB
		a.status.Volume = interfaces.AudioVolumeMin
	} else if decibels >= config.AudioMaxVolumedB {
		a.volume.Volume = config.AudioMaxVolumedB
		a.volume.Silent = false
		a.status.Volume = interfaces.AudioVolumeMax
	} else {
		a.volume.Silent = false
		a.volume.Volume = decibels
		a.status.Volume = volume
	}
	a.status.Action = interfaces.AudioActionSetVolume
	speaker.Unlock()
	go a.flushStatus()
}

// SetMute mutes and un-mutes audio
func (a *Audio) SetMute(muted bool) {

	if muted {
		logrus.Info("Mute audio")
	} else {
		logrus.Info("Unmute audio")
	}
	speaker.Lock()
	if a.ctrl == nil {
		return
	}
	a.ctrl.Paused = false
	a.volume.Silent = muted
	a.status.Muted = muted
	speaker.Unlock()
	go a.flushStatus()
}

func (a *Audio) streamCompleted() {
	logrus.Debug("audio stream complete")
	err := a.closeOldStream()
	if err != nil {
		logrus.Errorf("complete stream: %v", err)
	}
	if a.songCompleteFunc != nil {
		a.songCompleteFunc()
	}
}

func (a *Audio) closeOldStream() error {
	// don't use locking here, since speaker calls streamCompleted, which calls this to close reader
	var err error
	var streamErr error
	if a.streamer != nil {
		streamErr = a.streamer.Err()
		if streamErr != nil {
			streamErr = fmt.Errorf("streamer error: %v", streamErr)
		}
		err = a.streamer.Close()
		if err != nil {
			err = fmt.Errorf("close streamer: %v", err)
		} else {
			logrus.Debug("closed old streamer")
		}
		a.streamer = nil
	} else {
		err = fmt.Errorf("audio stream completed but streamer is nil")
	}
	if err == nil && a.streamer == nil {
		return nil
	}
	if err != nil && streamErr != nil {
		return fmt.Errorf("%v, %v", err, streamErr)
	}
	if err != nil {
		return err
	}
	if streamErr != nil {
		return streamErr
	}
	return nil
}

// gather latest status and flush it to callbacks
func (a *Audio) updateStatus() {
	past := a.getPastTicks()
	speaker.Lock()
	a.status.SongPast = past
	a.status.Action = interfaces.AudioActionTimeUpdate
	speaker.Unlock()
	a.flushStatus()
}

func (a *Audio) flushStatus() {
	speaker.Lock()
	status := a.status
	speaker.Unlock()
	for _, v := range a.statusCallbacks {
		v(status)
	}
}

// play song from io reader. Only song/album/artist/imageurl are used from status.
func (a *Audio) playSongFromReader(metadata songMetadata) error {
	// decode
	var streamer beep.StreamSeekCloser
	var err error
	switch metadata.format {
	case audioFormatMp3:
		streamer, _, err = mp3.Decode(metadata.reader)
	case audioFormatFlac:
		streamer, _, err = flac.Decode(metadata.reader)
	default:
		return fmt.Errorf("unknown audio format: %s", metadata.format)
	}
	if err != nil {
		return fmt.Errorf("decode audio stream: %v", err)
	}

	// play
	logrus.Debug("Setting new streamer")
	if streamer == nil {
		return fmt.Errorf("empty streamer")
	}
	stream := beep.Seq(streamer, beep.Callback(a.streamCompleted))
	speaker.Clear()
	speaker.Lock()
	old := a.streamer
	a.mixer.Clear()
	a.streamer = streamer
	a.mixer.Add(stream)
	speaker.Unlock()
	if old != nil {
		err := old.Close()
		if err != nil {
			err = fmt.Errorf("failed to close old stream: %v", err)
		}
	} else {
		logrus.Warning("No old streamer to close")
	}
	speaker.Play(a.volume)
	speaker.Lock()

	a.status.Song = metadata.song
	a.status.Album = metadata.album
	a.status.Artist = metadata.artist
	a.status.AlbumImageUrl = metadata.albumImageUrl
	a.status.State = interfaces.AudioStatePlaying
	a.status.Action = interfaces.AudioActionPlay
	speaker.Unlock()
	a.flushStatus()
	return err
}

// linear scaling with a & b coefficients
var volumeTodBA = float32(config.AudioMaxVolumedB-config.AudioMinVolumedB) /
	(config.AudioMaxVolume - config.AudioMinVolume)
var volumeTodBB = float32(config.AudioMinVolumedB - config.AudioMinVolume)

// Transform volume to db
func volumeTodB(volume int) float32 {
	return volumeTodBA*float32(volume) + volumeTodBB
}

// how many ticks current track has played
func (a *Audio) getPastTicks() interfaces.AudioTick {
	speaker.Lock()
	defer speaker.Unlock()
	if a.streamer == nil {
		return 0
	}
	left := a.streamer.Position() / config.AudioSamplingRate
	return interfaces.AudioTick((time.Second * time.Duration(left)).Milliseconds())
}
