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
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"path"
	"strings"
	"syscall"
)

// AppConfig is a configuration loaded during startup
var AppConfig *Config

type Config struct {
	Server Server `yaml:"server"`
	Player Player `yaml:"player"`

	configFile string
	configDir  string
}

func (c *Config) ConfigFile() string {
	return c.configFile
}

type Server struct {
	Url       string `yaml:"server_url"`
	Username  string `yaml:"username"`
	Token     string `yaml:"token"`
	UserId    string `yaml:"user_id"`
	DeviceId  string `yaml:"device_id"`
	ServerId  string `yaml:"server_id"`
	MusicView string `yaml:"music_view"`
}

type Player struct {
	PageSize            int    `yaml:"page_size"`
	LogFile             string `yaml:"log_file"`
	LogLevel            string `yaml:"log_level"`
	LimitRecentlyPlayed bool   `yaml:"limit_recent_songs"`
	MouseEnabled        bool   `yaml:"enable_mouse"`
	DoubleClickMs       int    `yaml:"mouse_double_click_interval_ms"`
	AudioBufferingMs    int    `yaml:"audio_buffering_ms"`
	HttpBufferingS      int    `yaml:"http_buffering_s"`
	// memory limit in MiB
	HttpBufferingLimitMem int `yaml:"http_buffering_limit_mem"`
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
}

// initialize new config with some sensible values
func (c *Config) initNewConfig() {
	c.Player.fillDefaults()
	c.Player.MouseEnabled = true
	// booleans are hard to determine whether they are set or not,
	// so only fill this here
	c.Player.LimitRecentlyPlayed = true
}

// can config file be considered empty / not configured
func (c *Config) isEmptyConfig() bool {
	return c.Server.UserId == "" &&
		c.Server.ServerId == "" &&
		c.Server.MusicView == "" &&
		c.Server.Token == ""
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
