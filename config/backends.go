/*
 * Copyright 2020 Tero Vierimaa
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package config

import "github.com/spf13/viper"

type Backend interface {
	DumpConfig() interface{}
	GetType() string
}

type Jellyfin struct {
	Url       string `yaml:"server_url"`
	Username  string `yaml:"username"`
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
