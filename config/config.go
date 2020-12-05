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

var configIsEmpty bool

type Config struct {
	Jellyfin Jellyfin `yaml:"jellyfin"`
	Subsonic Subsonic `yaml:"subsonic"`
	Player   Player   `yaml:"player"`
	Gui      Gui      `yaml:"gui"`
}

type Gui struct {
	PageSize            int  `yaml:"page_size"`
	DebugMode           bool `yaml:"debug_mode"`
	LimitRecentlyPlayed bool `yaml:"limit_recent_songs"`
	MouseEnabled        bool `yaml:"enable_mouse"`
	DoubleClickMs       int  `yaml:"mouse_double_click_interval_ms"`
	// valid types: artist,album,song,playlist,genre
	SearchTypes        []models.ItemType `yaml:"search_types"`
	SearchResultsLimit int               `yaml:"search_results_limit"`

	VolumeSteps int `yaml:"volume_steps"`

	// EnableSorting enables sorting on remote server
	EnableSorting bool `yaml:"enable_sorting"`
	// EnableFiltering enables filtering on remote server
	EnableFiltering bool `yaml:"enable_filtering"`
	// EnableResultsFiltering enables filtering existing results, 'search inside results'.
	EnableResultsFiltering bool `yaml:"enable_results_filtering"`
}

type Player struct {
	Server           string `yaml:"server"`
	LogFile          string `yaml:"log_file"`
	LogLevel         string `yaml:"log_level"`
	AudioBufferingMs int    `yaml:"audio_buffering_ms"`
	HttpBufferingS   int    `yaml:"http_buffering_s"`
	// memory limit in MiB
	HttpBufferingLimitMem int  `yaml:"http_buffering_limit_mem"`
	EnableRemoteControl   bool `yaml:"enable_remote_control"`
}

func (g *Gui) sanitize() {
	if g.PageSize <= 0 || g.PageSize > 500 {
		g.PageSize = 100
		PageSize = g.PageSize
	}
	if g.DoubleClickMs <= 0 {
		g.DoubleClickMs = 220
	}
	if len(g.SearchTypes) == 0 {
		g.SearchTypes = []models.ItemType{models.TypeArtist, models.TypeAlbum, models.TypeSong, models.TypePlaylist}
	}
	if g.SearchResultsLimit == 0 {
		g.SearchResultsLimit = 30
	}
	if g.VolumeSteps < 2 || g.VolumeSteps > 50 {
		g.VolumeSteps = 20
	}
}

func (p *Player) sanitize() {

	if p.LogFile == "" {
		dir := os.TempDir()
		p.LogFile = path.Join(dir, AppNameLower+".log")
	}
	if p.LogLevel == "" {
		p.LogLevel = logrus.WarnLevel.String()
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

}

// initialize new config with some sensible values
func (c *Config) initNewConfig() {
	c.Player.sanitize()
	c.Gui.sanitize()
	c.Gui.MouseEnabled = true
	c.Player.EnableRemoteControl = true
	// booleans are hard to determine whether they are set or not,
	// so only fill this here
	c.Gui.LimitRecentlyPlayed = true
	if c.Player.Server == "" {
		c.Player.Server = "jellyfin"
	}
	c.Player.LogLevel = logrus.InfoLevel.String()

	tempDir := os.TempDir()
	c.Player.LogFile = path.Join(tempDir, "jellycli.log")

	c.Gui.EnableResultsFiltering = true
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
			LogFile:               viper.GetString("player.logfile"),
			LogLevel:              viper.GetString("player.loglevel"),
			AudioBufferingMs:      viper.GetInt("player.audio_buffering_ms"),
			HttpBufferingS:        viper.GetInt("player.http_buffering_s"),
			HttpBufferingLimitMem: viper.GetInt("player.http_buffering_limit_mem"),
			EnableRemoteControl:   viper.GetBool("player.enable_remote_control"),
		},
		Gui: Gui{
			PageSize:            viper.GetInt("gui.pagesize"),
			DebugMode:           viper.GetBool("gui.debug_mode"),
			LimitRecentlyPlayed: viper.GetBool("gui.limit_recently_played"),
			MouseEnabled:        viper.GetBool("gui.mouse_enabled"),
			DoubleClickMs:       viper.GetInt("gui.double_click_ms"),
			SearchResultsLimit:  viper.GetInt("gui.search_results_limit"),
			VolumeSteps:         viper.GetInt("gui.volume_steps"),

			EnableSorting:          viper.GetBool("gui.enable_sorting"),
			EnableFiltering:        viper.GetBool("gui.enable_filtering"),
			EnableResultsFiltering: viper.GetBool("gui.enable_results_filtering"),
		},
	}

	searchTypes := viper.GetStringSlice("gui.search_types")
	for _, v := range searchTypes {
		searchType := models.ItemType(v)
		AppConfig.Gui.SearchTypes = append(AppConfig.Gui.SearchTypes, searchType)
	}

	if len(searchTypes) == 0 {
		AppConfig.Gui.SearchTypes = []models.ItemType{models.TypeArtist, models.TypeAlbum,
			models.TypeSong, models.TypePlaylist}

	}

	if AppConfig.Jellyfin.Url == "" && AppConfig.Subsonic.Url == "" {
		configIsEmpty = true
		setDefaults()
	} else {
		AppConfig.Player.sanitize()
		AppConfig.Gui.sanitize()
	}
	AudioBufferPeriod = time.Millisecond * time.Duration(AppConfig.Player.AudioBufferingMs)
	VolumeStepSize = (AudioMinVolume + AudioMaxVolume) / AppConfig.Gui.VolumeSteps
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
	if configIsEmpty {
		AppConfig.initNewConfig()
		err := SaveConfig()
		if err != nil {
			logrus.Errorf("save config file: %v", err)
		}
	}
}

// set AppConfig. This is needed for testing.
func configFrom(conf *Config) {
	AppConfig = conf
}

func UpdateViper() {
	viper.Set("jellyfin.url", AppConfig.Jellyfin.Url)
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
	viper.Set("player.logfile", AppConfig.Player.LogFile)
	viper.Set("player.loglevel", AppConfig.Player.LogLevel)
	viper.Set("player.http_buffering_s", AppConfig.Player.HttpBufferingS)
	viper.Set("player.http_buffering_limit_mem", AppConfig.Player.HttpBufferingLimitMem)
	viper.Set("player.enable_remote_control", AppConfig.Player.EnableRemoteControl)
	viper.Set("player.audio_buffering_ms", AppConfig.Player.AudioBufferingMs)

	viper.Set("gui.search_results_limit", AppConfig.Gui.SearchResultsLimit)
	viper.Set("gui.debug_mode", AppConfig.Gui.DebugMode)
	viper.Set("gui.limit_recently_played", AppConfig.Gui.LimitRecentlyPlayed)
	viper.Set("gui.mouse_enabled", AppConfig.Gui.MouseEnabled)
	viper.Set("gui.double_click_ms", AppConfig.Gui.DoubleClickMs)
	viper.Set("gui.pagesize", AppConfig.Gui.PageSize)
	viper.Set("gui.volume_steps", AppConfig.Gui.VolumeSteps)

	sTypes := make([]string, len(AppConfig.Gui.SearchTypes))
	for i, v := range AppConfig.Gui.SearchTypes {
		sTypes[i] = string(v)
	}

	viper.Set("gui.search_types", sTypes)

	viper.Set("gui.enable_sorting", AppConfig.Gui.EnableSorting)
	viper.Set("gui.enable_filtering", AppConfig.Gui.EnableFiltering)
	viper.Set("gui.enable_results_filtering", AppConfig.Gui.EnableResultsFiltering)
}
