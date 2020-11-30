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
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"path"
	"reflect"
	"testing"
	"tryffel.net/go/jellycli/models"
)

func TestConfigToFromViper(t *testing.T) {
	// test every var is read & written
	conf := &Config{
		Jellyfin: Jellyfin{
			Url:       "http://localhost",
			Token:     "jellytoken",
			UserId:    "jellyuser",
			DeviceId:  "jellydevice",
			ServerId:  "jellyserver",
			MusicView: "jellyview",
		},
		Subsonic: Subsonic{
			Url:      "https://localhost",
			Username: "subuser",
			Salt:     "subsalt",
			Token:    "subtoken",
		},
		Player: Player{
			Server:                "jellyfin",
			LogFile:               "/var/log/jellyfin.log",
			LogLevel:              "info",
			AudioBufferingMs:      150,
			HttpBufferingS:        5,
			HttpBufferingLimitMem: 20,
			EnableRemoteControl:   true,
		},
		Gui: Gui{
			PageSize:               100,
			DebugMode:              true,
			LimitRecentlyPlayed:    true,
			MouseEnabled:           true,
			DoubleClickMs:          200,
			SearchTypes:            []models.ItemType{"Artist", "Album"},
			SearchResultsLimit:     10,
			EnableSorting:          true,
			EnableFiltering:        true,
			EnableResultsFiltering: true,
		},
	}

	viper.Reset()

	// set known values
	configFrom(conf)

	// set config
	UpdateViper()

	// clear config
	configFrom(&Config{})

	//read config
	err := ConfigFromViper()
	if err != nil {
		t.Fatalf("read config from viper: %v", err)
	}

	if !reflect.DeepEqual(conf, AppConfig) {
		t.Errorf("written / read viper configs do not match")
	}
}

func TestInitEmptyConfig(t *testing.T) {
	// test new config file is sane

	tmpDir := t.TempDir()
	configFile := path.Join(tmpDir, "jellycli.yaml")
	viper.SetConfigFile(configFile)

	conf := &Config{
		Jellyfin: Jellyfin{},
		Subsonic: Subsonic{},
		Player: Player{
			Server:                "jellyfin",
			LogFile:               "/tmp/jellycli.log",
			LogLevel:              "info",
			AudioBufferingMs:      150,
			HttpBufferingS:        5,
			HttpBufferingLimitMem: 20,
			EnableRemoteControl:   true,
		},
		Gui: Gui{
			PageSize:            100,
			DebugMode:           false,
			LimitRecentlyPlayed: true,
			MouseEnabled:        true,
			DoubleClickMs:       220,
			SearchTypes: []models.ItemType{models.TypeArtist, models.TypeAlbum,
				models.TypeSong, models.TypePlaylist},
			SearchResultsLimit:     30,
			EnableSorting:          false,
			EnableFiltering:        false,
			EnableResultsFiltering: true,
		},
	}

	viper.Reset()

	//read config
	err := ConfigFromViper()
	if err != nil {
		t.Fatalf("read config from viper: %v", err)
	}

	cmp.AllowUnexported()
	diff := cmp.Diff(conf, AppConfig)
	if diff != "" {
		t.Errorf("written / read viper config differs: %s", diff)

	}
}

func TestSanitizeConfig(t *testing.T) {
	// test existing config file is sanitized

	invalidConf := &Config{
		Jellyfin: Jellyfin{
			Url:       "http://localhost",
			Token:     "jellytoken",
			UserId:    "jellyuser",
			DeviceId:  "jellydevice",
			ServerId:  "jellyserver",
			MusicView: "jellyview",
		},
		Subsonic: Subsonic{
			Url:      "https://localhost",
			Username: "subuser",
			Salt:     "subsalt",
			Token:    "subtoken",
		},
		Player: Player{
			Server:                "",
			LogFile:               "/var/log/jellyfin.log",
			LogLevel:              "",
			AudioBufferingMs:      0,
			HttpBufferingS:        0,
			HttpBufferingLimitMem: 0,
			EnableRemoteControl:   true,
		},
		Gui: Gui{
			PageSize:               1000,
			DebugMode:              true,
			LimitRecentlyPlayed:    true,
			MouseEnabled:           true,
			DoubleClickMs:          0,
			SearchTypes:            []models.ItemType{"Artist", "Album"},
			SearchResultsLimit:     0,
			EnableSorting:          true,
			EnableFiltering:        true,
			EnableResultsFiltering: true,
		},
	}

	viper.Reset()

	// set known values
	configFrom(invalidConf)

	// set config
	UpdateViper()

	invalidConf.Player.LogLevel = "warning"
	invalidConf.Player.AudioBufferingMs = 150
	invalidConf.Player.HttpBufferingS = 5
	invalidConf.Player.HttpBufferingLimitMem = 20

	invalidConf.Gui.PageSize = 100
	invalidConf.Gui.DoubleClickMs = 220
	invalidConf.Gui.SearchResultsLimit = 30

	// clear config
	configFrom(&Config{})

	//read config
	err := ConfigFromViper()
	if err != nil {
		t.Fatalf("read config from viper: %v", err)
	}

	cmp.AllowUnexported()
	diff := cmp.Diff(invalidConf, AppConfig)
	if diff != "" {
		t.Errorf("sanitized config invalid: %s", diff)
	}
}
