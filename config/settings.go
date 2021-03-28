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

package config

import (
	"fmt"
	"time"
)

const (
	AppName      = "Jellycli"
	AppNameLower = "jellycli"
	Version      = "0.9.1"
)

var (
	// PageSize controls paging and is read from config file
	PageSize            = 100
	LimitRecentlyPlayed = false
	// how many recent song to show, if limited
	LimitedRecentlyPlayedCount = 24
	AudioBufferPeriod          = time.Millisecond * 100

	VolumeStepSize = 5
)

// audio configuration
const (
	// AudioSamplingRate is default sampling rate. This may vary depending on song being played.
	AudioSamplingRate = 44100

	// Volume range in decibels
	AudioMinVolumedB = -6
	AudioMaxVolumedB = 0

	AudioMinVolume = 0
	AudioMaxVolume = 100

	// Audio volume is logarithmic, which base to use
	AudioVolumeLogBase = 2

	CacheTimeout = time.Minute * 5
)

// AppNameVersion returns string containing application name and current version
func AppNameVersion() string {
	return fmt.Sprintf("%s v%s", AppName, Version)
}

// LogFile is log file location
var LogFile string

// ConfigFile is absolute location for configuration file
var ConfigFile string
