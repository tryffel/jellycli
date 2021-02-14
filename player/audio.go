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

package player

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
	"github.com/sirupsen/logrus"
	"io"
	"time"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
)

type audioFormat string

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

	currentSampleRate int
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

	a.currentSampleRate = config.AudioSamplingRate
	return a
}

func initAudio() error {
	err := speaker.Init(config.AudioSamplingRate, config.AudioSamplingRate/1000*
		int(config.AudioBufferPeriod.Milliseconds()))
	if err != nil {
		return fmt.Errorf("init speaker: %v", err)
	}
	return nil
}

func (a *Audio) SetShuffle(shuffle bool) {
	if shuffle {
		logrus.Info("Enable shuffle")
	} else {
		logrus.Info("Disable shuffle")
	}

	speaker.Lock()
	defer speaker.Unlock()
	a.status.Shuffle = shuffle
	a.status.Action = interfaces.AudioActionShuffleChanged
	go a.flushStatus()
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
	logrus.Infof("Stop audio")
	speaker.Lock()
	a.status.State = interfaces.AudioStateStopped
	a.status.Action = interfaces.AudioActionStop
	a.ctrl.Paused = false
	a.status.Paused = false
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
	logrus.Info("Next song")
	speaker.Lock()
	a.status.Action = interfaces.AudioActionNext
	speaker.Unlock()
	go a.flushStatus()
}

// Previous plays previous track. If previous track does not exist, do nothing.
func (a *Audio) Previous() {
	logrus.Info("Previous song")
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

func (a *Audio) ToggleMute() {
	logrus.Info("Toggle mute")
	speaker.Lock()
	muted := a.status.Muted
	speaker.Unlock()
	a.SetMute(!muted)
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
			if streamErr != io.EOF {
				logrus.Errorf("streamer error: %v", streamErr)
			} else {
				logrus.Warning("got streamer error EOF")
				err = nil
			}
		}
		err = a.streamer.Close()
		if err != nil {
			if err == io.EOF {
				// pass
			} else {
				err = fmt.Errorf("close streamer: %v", err)
			}
		} else {
			logrus.Debug("closed old streamer")
		}
		a.streamer = nil
	} else {
		err = fmt.Errorf("audio stream completed but streamer is nil")
	}
	return err
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
	var songFormat beep.Format
	var streamer beep.StreamSeekCloser
	var err error
	switch metadata.format {
	case interfaces.AudioFormatMp3:
		streamer, songFormat, err = mp3.Decode(metadata.reader)
	case interfaces.AudioFormatFlac:
		streamer, songFormat, err = flac.Decode(metadata.reader)
	case interfaces.AudioFormatWav:
		streamer, songFormat, err = wav.Decode(metadata.reader)
	case interfaces.AudioFormatOgg:
		streamer, songFormat, err = vorbis.Decode(metadata.reader)
	default:
		return fmt.Errorf("unknown audio format: %s", metadata.format)
	}
	if err != nil {
		return fmt.Errorf("decode audio stream: %v", err)
	}

	logrus.Debugf("Song %s samplerate: %d Hz", metadata.song.Name, songFormat.SampleRate.N(time.Second))
	sampleRate := songFormat.SampleRate.N(time.Second)
	if a.currentSampleRate != sampleRate {
		logrus.Debugf("Set samplerate to %d kHz", sampleRate/1000)
		err = speaker.Init(songFormat.SampleRate, sampleRate/1000*
			int(config.AudioBufferPeriod.Seconds()*1000))
		if err != nil {
			logrus.Errorf("Update sample rate (%d -> %d): %v", a.currentSampleRate, sampleRate, err)
		} else {
			a.currentSampleRate = sampleRate
		}
	}
	logrus.Debug("Setting new streamer from ", metadata.format.String())
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
