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

package player

import (
	"errors"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/sirupsen/logrus"
	"io"
	"sync"
	"time"
	"tryffel.net/pkg/jellycli/config"
)

type Format int

const (
	FormatMp3  = 1
	FormatFlac = 2
)

// Transform volume to db
var volumeToDbA = float32(config.AudioMaxVolumeDb-config.AudioMinVolumeDb) /
	(config.AudioMaxVolume - config.AudioMinVolume)
var volumeToDbB = float32(config.AudioMinVolumeDb - config.AudioMinVolume)

func volumeToDb(volume int) float32 {
	return volumeToDbA*float32(volume) + volumeToDbB
}

//Initialize speaker
func initAudio() error {
	err := speaker.Init(config.AudioSamplingRate, config.AudioSamplingRate/1000*
		int(config.AudioBufferPeriod.Seconds()*1000))
	if err != nil {
		return fmt.Errorf("speaker initialization failed: %v", err)
	}
	return nil
}

type audio struct {
	lock            sync.RWMutex
	streamer        beep.StreamSeekCloser
	ctrl            *beep.Ctrl
	volume          *effects.Volume
	mixer           *beep.Mixer
	streamCompleted chan bool
}

func newAudio(streamDoneChan chan bool) *audio {
	a := &audio{
		streamer:        nil,
		streamCompleted: streamDoneChan,
		ctrl: &beep.Ctrl{
			Streamer: nil,
			Paused:   false,
		},
		volume: &effects.Volume{
			Streamer: nil,
			Base:     config.AudioVolumeLogBase,
			Volume:   (config.AudioMinVolumeDb + config.AudioMaxVolumeDb) / 2,
			Silent:   false,
		},
		mixer: &beep.Mixer{},
	}
	a.ctrl.Streamer = a.mixer
	a.volume.Streamer = a.ctrl
	return a
}

// Notify upstream channel that stream has completed
func (a *audio) streamCompletedCB() {
	logrus.Info("stream completed")
	a.lock.Lock()
	if a.streamer != nil {
		err := a.streamer.Err()
		if err != nil {
			logrus.Error("Streamer returner error: ", err.Error())
		}
		err = a.streamer.Close()
		if err != nil {
			logrus.Error("failed to close stream: %v", err)
		}
	}
	a.lock.Unlock()

	if a.streamCompleted == nil {
		return
	}
	a.streamCompleted <- true
}

func (a *audio) playNewStream(streamer beep.StreamSeekCloser, play bool) error {
	logrus.Debug("Setting new streamer")
	if streamer == nil {
		return fmt.Errorf("empty streamer")
	}
	var err error
	a.lock.Lock()
	speaker.Clear()
	speaker.Lock()
	old := a.streamer
	a.mixer.Clear()
	a.streamer = streamer
	a.mixer.Add(streamer)
	a.ctrl.Paused = !play
	speaker.Unlock()
	a.lock.Unlock()
	if old != nil {
		err := old.Close()
		if err != nil {
			err = fmt.Errorf("failed to close old stream: %v", err)
		}
	} else {
	}
	speaker.Play(a.volume)
	return err
}

func (a *audio) newFileStream(reader io.ReadCloser, format Format) error {
	var streamer beep.StreamSeekCloser
	var err error
	switch format {
	case FormatMp3:
		streamer, _, err = mp3.Decode(reader)
	case FormatFlac:
		var f beep.Format
		streamer, f, err = flac.Decode(reader)
		logrus.Info("Song samplerate: ", f.SampleRate)
	default:
		err = errors.New("unknown audio format")
	}
	if err != nil {
		return fmt.Errorf("failed to initialize stream: %v", err)
	}

	return a.playNewStream(streamer, true)
}

func (a *audio) timePast() time.Duration {
	a.lock.RLock()
	defer a.lock.RUnlock()
	if a.streamer == nil {
		return 0
	}
	speaker.Lock()
	defer speaker.Unlock()
	left := a.streamer.Position() / config.AudioSamplingRate

	return time.Second * time.Duration(left)
}

func (a *audio) pause(state bool) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.ctrl == nil {
		return
	}
	if state {
		logrus.Info("Pause")
	} else {
		logrus.Info("Continue")
	}
	speaker.Lock()
	a.ctrl.Paused = state
	speaker.Unlock()
}

func (a *audio) setVolume(percent int) {
	decibels := float64(volumeToDb(percent))
	logrus.Debugf("Set volume to %d %s -> %.2f Db", percent, "%", decibels)
	a.lock.Lock()
	defer a.lock.Unlock()
	speaker.Lock()
	defer speaker.Unlock()
	if decibels <= config.AudioMinVolumeDb {
		a.volume.Silent = true
		a.volume.Volume = config.AudioMinVolumeDb
	} else if decibels >= config.AudioMaxVolumeDb {
		a.volume.Volume = config.AudioMaxVolumeDb
		a.volume.Silent = false
	} else {
		a.volume.Silent = false
		a.volume.Volume = decibels
	}
}

func (a *audio) stop() {
	a.pause(true)
	a.lock.Lock()
	a.mixer.Clear()
	a.lock.Unlock()
}
