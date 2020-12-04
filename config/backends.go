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

import "github.com/spf13/viper"

type Backend interface {
	DumpConfig() interface{}
	GetType() string
}

type Jellyfin struct {
	Url       string `yaml:"server_url"`
	Token     string `yaml:"token"`
	UserId    string `yaml:"user_id"`
	DeviceId  string `yaml:"device_id"`
	ServerId  string `yaml:"server_id"`
	MusicView string `yaml:"music_view"`
}

func (j *Jellyfin) DumpConfig() interface{} {
	return j
}

func (j *Jellyfin) GetType() string {
	return "jellyfin"
}

type Subsonic struct {
	Url      string `yaml:"server_url"`
	Username string `yaml:"username"`
	Salt     string `yaml:"salt"`
	Token    string `yaml:"token"`
}

func (s *Subsonic) DumpConfig() interface{} {
	return s
}

func (s *Subsonic) GetType() string {
	return "subsonic"
}

// KeyValueProvider provides means to request new values for outdated values,
// to request new password or url.
type KeyValueProvider interface {
	// Get returns value for key. Sensitive flags key as hidden.
	// Key is of format block.value from config file. Label is user-friendly label.
	Get(key string, sensitive bool, label string) (string, error)
}

// StdinConfigProvider reads config keys from stdin.
type StdinConfigProvider struct{}

func (s *StdinConfigProvider) Get(key string, sensitive bool, label string) (string, error) {
	return ReadUserInput(label, sensitive)
}

// ViperStdConfigProvider reads key first from viper (config file & env) and after that reads from stdin.
type ViperStdConfigProvider struct{}

func (s *ViperStdConfigProvider) Get(key string, sensitive bool, label string) (string, error) {
	val := viper.GetString(key)
	if val != "" {
		return val, nil
	}
	return ReadUserInput(label, sensitive)
}
