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

// Package config contains application-wide configurations and constants. Parts of configuration are user-editable
// and per-instance and needs to be persisted. Others are static and meant for tuning the application.
// It also contains some helper methods to read and write config files and create directories when needed.
package config

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"path"
	"strings"
	"syscall"
	"time"
	"tryffel.net/go/jellycli/models"
)

// AppConfig is a configuration loaded during startup
var AppConfig *Config

type Config struct {
	Jellyfin Jellyfin `yaml:"jellyfin"`
	Subsonic Subsonic `yaml:"subsonic"`
	Player   Player   `yaml:"player"`

	configFile string
	configDir  string

	configIsEmpty bool
}

func (c *Config) ConfigFile() string {
	return c.configFile
}

type Player struct {
	Server string `yaml:"server"`

	PageSize            int    `yaml:"page_size"`
	LogFile             string `yaml:"log_file"`
	LogLevel            string `yaml:"log_level"`
	DebugMode           bool   `yaml:"debug_mode"`
	LimitRecentlyPlayed bool   `yaml:"limit_recent_songs"`
	MouseEnabled        bool   `yaml:"enable_mouse"`
	DoubleClickMs       int    `yaml:"mouse_double_click_interval_ms"`
	AudioBufferingMs    int    `yaml:"audio_buffering_ms"`
	HttpBufferingS      int    `yaml:"http_buffering_s"`
	// memory limit in MiB
	HttpBufferingLimitMem int `yaml:"http_buffering_limit_mem"`

	EnableRemoteControl bool `yaml:"enable_remote_control"`

	// valid types: artist,album,song,playlist,genre
	SearchTypes        []models.ItemType `yaml:"search_types"`
	SearchResultsLimit int               `yaml:"search_results_limit"`
}

func (p *Player) fillDefaults() {
	if p.PageSize <= 0 || p.PageSize > 500 {
		p.PageSize = 100
	}
	if p.LogFile == "" {
		dir := os.TempDir()
		p.LogFile = path.Join(dir, AppNameLower+".log")
	}
	if p.LogLevel == "" {
		p.LogLevel = logrus.WarnLevel.String()
	}
	if p.DoubleClickMs <= 0 {
		p.DoubleClickMs = 220
	}
	if p.AudioBufferingMs == 0 {
		p.AudioBufferingMs = 150
	}
	if p.HttpBufferingS == 0 {
		p.HttpBufferingS = 5
	}
	if p.HttpBufferingLimitMem == 0 {
		p.HttpBufferingLimitMem = 20
	}
	if len(p.SearchTypes) == 0 {
		p.SearchTypes = []models.ItemType{models.TypeArtist, models.TypeAlbum, models.TypeSong, models.TypePlaylist}
	}
	if p.SearchResultsLimit == 0 {
		p.SearchResultsLimit = 30
	}
}

// initialize new config with some sensible values
func (c *Config) initNewConfig() {
	c.Player.fillDefaults()
	c.Player.MouseEnabled = true
	c.Player.EnableRemoteControl = true
	// booleans are hard to determine whether they are set or not,
	// so only fill this here
	c.Player.LimitRecentlyPlayed = true
	c.Player.SearchTypes = []models.ItemType{"artist,album,song,playlist,genre"}
	c.Player.Server = "jellyfin"
	c.Player.LogLevel = logrus.InfoLevel.String()

	tempDir := os.TempDir()
	c.Player.LogFile = path.Join(tempDir, "jellycli.log")
}

// can config file be considered empty / not configured
func (c *Config) isEmptyConfig() bool {
	return c.Jellyfin.UserId == "" &&
		c.Subsonic.Url == "" &&
		c.Player.Server == ""
}

// ReadUserInput reads value from stdin. Name is printed like 'Enter <name>. If mask is true, input is masked.
func ReadUserInput(name string, mask bool) (string, error) {
	fmt.Print("Enter ", name, ": ")
	var val string
	var err error
	if mask {
		// needs cast for windows
		raw, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", fmt.Errorf("failed to read user input: %v", err)
		}
		val = string(raw)
		fmt.Println()
	} else {
		reader := bufio.NewReader(os.Stdin)
		val, err = reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read user input: %v", err)
		}
	}
	val = strings.Trim(val, "\n\r")
	return val, nil
}

// ConfigFromViper reads full application configuration from viper.
func ConfigFromViper() error {

	AppConfig = &Config{
		Jellyfin: Jellyfin{
			Url:       viper.GetString("jellyfin.url"),
			Username:  viper.GetString("jellyfin.username"),
			Token:     viper.GetString("jellyfin.token"),
			UserId:    viper.GetString("jellyfin.userid"),
			DeviceId:  viper.GetString("jellyfin.device_id"),
			ServerId:  viper.GetString("jellyfin.server_id"),
			MusicView: viper.GetString("jellyfin.music_view"),
		},
		Subsonic: Subsonic{
			Url:      viper.GetString("subsonic.url"),
			Username: viper.GetString("subsonic.username"),
			Salt:     viper.GetString("subsonic.salt"),
			Token:    viper.GetString("subsonic.token"),
		},
		Player: Player{
			Server:                viper.GetString("player.server"),
			PageSize:              viper.GetInt("player.pagesize"),
			LogFile:               viper.GetString("player.logfile"),
			LogLevel:              viper.GetString("player.loglevel"),
			DebugMode:             viper.GetBool("player.debug_mode"),
			LimitRecentlyPlayed:   viper.GetBool("player.limit_recently_played"),
			MouseEnabled:          viper.GetBool("player.mouse_enabled"),
			DoubleClickMs:         viper.GetInt("player.double_click_ms"),
			AudioBufferingMs:      viper.GetInt("player.audio_buffering_ms"),
			HttpBufferingS:        viper.GetInt("player.http_buffering_s"),
			HttpBufferingLimitMem: viper.GetInt("player.http_buffering_limit_mem"),
			EnableRemoteControl:   viper.GetBool("player.enable_remote_control"),
			SearchResultsLimit:    viper.GetInt("player.search_results_limit"),
		},
	}

	searchTypes := viper.GetStringSlice("player_search_types")
	for _, v := range searchTypes {
		searchType := models.ItemType(v)
		AppConfig.Player.SearchTypes = append(AppConfig.Player.SearchTypes, searchType)
	}

	if AppConfig.Jellyfin.Url == "" && AppConfig.Subsonic.Url == "" {
		AppConfig.configIsEmpty = true
		setDefaults()
	}

	AudioBufferPeriod = time.Millisecond * time.Duration(AppConfig.Player.AudioBufferingMs)
	return nil
}

func SaveConfig() error {
	UpdateViper()
	err := viper.WriteConfig()
	if err != nil {
		return fmt.Errorf("save config file: %v", err)
	}
	return nil
}

func setDefaults() {
	if AppConfig.configIsEmpty {
		AppConfig.initNewConfig()
		err := SaveConfig()
		if err != nil {
			logrus.Errorf("save config file: %v", err)
		}
	}
}

func UpdateViper() {
	viper.Set("jellyfin.url", AppConfig.Jellyfin.Url)
	viper.Set("jellyfin.username", AppConfig.Jellyfin.Username)
	viper.Set("jellyfin.token", AppConfig.Jellyfin.Token)
	viper.Set("jellyfin.userid", AppConfig.Jellyfin.UserId)
	viper.Set("jellyfin.device_id", AppConfig.Jellyfin.DeviceId)
	viper.Set("jellyfin.server_id", AppConfig.Jellyfin.ServerId)
	viper.Set("jellyfin.music_view", AppConfig.Jellyfin.MusicView)

	viper.Set("subsonic.url", AppConfig.Subsonic.Url)
	viper.Set("subsonic.username", AppConfig.Subsonic.Username)
	viper.Set("subsonic.salt", AppConfig.Subsonic.Salt)
	viper.Set("subsonic.token", AppConfig.Subsonic.Token)

	viper.Set("player.server", AppConfig.Player.Server)
	viper.Set("player.pagesize", AppConfig.Player.PageSize)
	viper.Set("player.logfile", AppConfig.Player.LogFile)
	viper.Set("player.loglevel", AppConfig.Player.LogLevel)
	viper.Set("player.debug_mode", AppConfig.Player.DebugMode)
	viper.Set("player.limit_recently_played", AppConfig.Player.LimitRecentlyPlayed)
	viper.Set("player.mouse_enabled", AppConfig.Player.MouseEnabled)
	viper.Set("player.double_click_ms", AppConfig.Player.DoubleClickMs)
	viper.Set("player.http_buffering_ms", AppConfig.Player.HttpBufferingS)
	viper.Set("player.http_buffering_limit_mem", AppConfig.Player.HttpBufferingLimitMem)
	viper.Set("player.enable_remote_control", AppConfig.Player.EnableRemoteControl)
	viper.Set("player.search_results_limit", AppConfig.Player.SearchResultsLimit)
}
