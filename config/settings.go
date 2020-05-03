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

package config

import (
	"fmt"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	AppName      = "Jellycli"
	AppNameLower = "jellycli"
	Version      = "0.4.1"

	LogLevel = logrus.WarnLevel
)

var (
	// PageSize controls paging and is read from config file
	PageSize = 100
)

// audio configuration
const (
	AudioSamplingRate = 44100
	AudioBufferPeriod = time.Millisecond * 50

	// Volume range in decibels
	AudioMinVolumedB = -6
	AudioMaxVolumedB = 0

	AudioMinVolume = 0
	AudioMaxVolume = 100

	// Audio volume is logarithmic, which base to use
	AudioVolumeLogBase = 2

	Debug = false

	CacheTimeout = time.Minute * 5

	SongHistorySize = 100

	LatestMusicCount = "50"
)

// AppNameVersion returns string containing application name and current version
func AppNameVersion() string {
	return fmt.Sprintf("%s v%s", AppName, Version)
}

// LogFile is log file location
var LogFile string

// ConfigFile is absolute location for configuration file
var ConfigFile string

var GridBordersShow = true
var GridBordersColor = Color.Border

//DebugGridBorders enables grid borders if config.Debug is true and if config.GridBordesShow is true
// Else do nothing
func DebugGridBorders(view *tview.Grid) {
	if Debug && GridBordersShow {
		view.SetBorders(true)
		view.SetBordersColor(GridBordersColor)

	}

}
