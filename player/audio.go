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
	streamer        beep.StreamSeekCloser
	ctrl            *beep.Ctrl
	volume          *effects.Volume
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
	}
	return a
}

// Notify upstream channel that stream has completed
func (a *audio) streamCompletedCB() {
	logrus.Info("stream completed")
	if a.streamer != nil {
		err := a.streamer.Close()
		if err != nil {
			logrus.Error("failed to close stream: %v", err)
		}
	}

	if a.streamCompleted == nil {
		return
	}
	a.streamCompleted <- true
}

func (a *audio) newStream(streamer beep.StreamSeekCloser) {
	a.streamer = streamer
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

	a.newStream(streamer)
	return nil
}

func (a *audio) playStream() error {
	if a.streamer == nil {
		return errors.New("no stream available")
	}
	// Play stream first, then call callback
	a.ctrl.Streamer = beep.Seq(a.streamer, beep.Callback(a.streamCompletedCB))
	a.volume.Streamer = a.ctrl
	logrus.Debug("Speaker set new play")
	speaker.Play(a.volume)

	a.pause(false)
	return nil
}

func (a *audio) timePast() time.Duration {
	if a.streamer == nil {
		return 0
	}
	speaker.Lock()
	defer speaker.Unlock()
	left := a.streamer.Position() / config.AudioSamplingRate

	return time.Second * time.Duration(left)
}

func (a *audio) pause(state bool) {
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
	decibels := volumeToDb(percent)
	logrus.Infof("Set volume to %f Db", decibels)
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
		a.volume.Volume = float64(decibels)
	}
}

func (a *audio) stop() {
	a.pause(true)
	//if a.streamers != nil {
	//	a.streamers.
	//}

}
