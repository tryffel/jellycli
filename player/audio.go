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

//Initialize speaker
func initAudio() error {
	err := speaker.Init(config.AudioSamplingRate, config.AudioSamplingRate/1000*int(config.AudioBufferPeriod.Seconds()*1000))
	if err != nil {
		return fmt.Errorf("speaker initialization failed: %v", err)
	}
	return nil
}

type audio struct {
	streamer        beep.StreamSeekCloser
	ctrl            *beep.Ctrl
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
	}
	return a
}

// Notify upstream channel that stream has completed
func (a *audio) streamCompletedCB() {
	logrus.Info("stream completed")
	if a.streamer != nil {
		//err := a.streamer.Close()
		//if err != nil {
		//logrus.Error("failed to close stream: %v", err)
		//}
	}

	if a.streamCompleted == nil {
		return
	}
	a.streamCompleted <- true
}

func (a *audio) newStream(reader io.ReadCloser, format Format) error {
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

	a.streamer = streamer
	return nil
}

func (a *audio) playStream() error {
	if a.streamer == nil {
		return errors.New("no stream available")
	}
	// Play stream first, then call callback
	a.ctrl.Streamer = beep.Seq(a.streamer, beep.Callback(a.streamCompletedCB))
	logrus.Debug("Speaker set new play")
	speaker.Play(a.ctrl)
	logrus.Debug("Speaker got new play")
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

func (a *audio) pause() error {
	if a.ctrl == nil || a.streamer == nil {
		return errors.New("no active stream")
	}
	logrus.Info("Toggle speaker pause")
	speaker.Lock()
	a.ctrl.Paused = !a.ctrl.Paused
	speaker.Unlock()
	logrus.Debug("Speaker pause toggled")
	return nil
}
