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
	"github.com/sirupsen/logrus"
	"testing"
	"tryffel.net/go/jellycli/interfaces"
)

func TestAudio_PlayPause(t *testing.T) {
	logrus.SetLevel(logrus.WarnLevel)
	audio := newAudio()

	wantPaused := func() {
		if !audio.ctrl.Paused {
			t.Errorf("expect audio.ctrl paused")
		}
		if !audio.status.Paused {
			t.Errorf("expect audio.status paused")
		}
	}

	wantContinue := func() {
		if audio.ctrl.Paused {
			t.Errorf("expect audio.ctrl non-paused")
		}
		if audio.status.Paused {
			t.Errorf("expect audio.status non-paused")
		}
	}

	// playpause
	wantContinue()
	audio.PlayPause()
	wantPaused()
	audio.PlayPause()
	wantContinue()

	// pause / continue
	audio.Pause()
	wantPaused()

	audio.Pause()
	wantPaused()

	audio.Continue()
	wantContinue()

	audio.Continue()
	wantContinue()
}

func TestAudio_SetVolume(t *testing.T) {
	logrus.SetLevel(logrus.WarnLevel)
	tests := []struct {
		name       string
		volume     interfaces.AudioVolume
		wantVolume float64
		wantSilent bool
	}{
		{
			volume:     interfaces.AudioVolume(0),
			wantVolume: -6,
			wantSilent: true,
		},
		{
			volume:     interfaces.AudioVolume(20),
			wantVolume: -4.800000190734863,
			wantSilent: false,
		},
		{
			volume:     interfaces.AudioVolume(50),
			wantVolume: -3,
			wantSilent: false,
		},
		{
			volume:     interfaces.AudioVolume(75),
			wantVolume: -1.5,
			wantSilent: false,
		},
		{
			volume:     interfaces.AudioVolume(100),
			wantVolume: 0,
			wantSilent: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newAudio()
			a.SetVolume(tt.volume)

			if a.status.Volume != tt.volume {
				t.Errorf("audio.status.volume, got: %d, want %d", a.status.Volume, tt.volume)
			}

			if a.volume.Silent != tt.wantSilent {
				t.Errorf("audio.volume.silent, got: %t, want %t", a.volume.Silent, tt.wantSilent)
			}

			if a.volume.Volume != tt.wantVolume {
				t.Errorf("audio.volume.volume (dB), got: %f, want %f", a.volume.Volume, tt.wantVolume)
			}
		})
	}
}

func TestAudio_SetMute(t *testing.T) {
	logrus.SetLevel(logrus.WarnLevel)
	audio := newAudio()

	if audio.status.Muted {
		t.Errorf("audio is muted on init")
	}

	audio.SetMute(true)

	if !audio.status.Muted {
		t.Errorf("want audio.status muted")
	}
	if !audio.volume.Silent {
		t.Errorf("want audio.volume muted")
	}

	audio.SetMute(false)
	if audio.status.Muted {
		t.Errorf("want audio.status not muted")
	}
	if audio.volume.Silent {
		t.Errorf("want audio.volume not muted")
	}
}
