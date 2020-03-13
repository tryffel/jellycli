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

package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path"
)

var configFile = AppNameLower + ".yaml"

//ReadConfigFile reads config file from given file. If file is empty, use default location provided by os
func ReadConfigFile(file string) (*Config, error) {
	if file == "" {
		configDir, _ := GetConfigDirectory()
		dir := path.Join(configDir, AppNameLower)
		file = path.Join(dir, configFile)
	}

	return readConfigFile(file)
}

func readConfigFile(file string) (*Config, error) {
	conf := &Config{}
	conf.configFile = file
	err := EnsureConfigDirExists()
	if err != nil {
		return nil, err
	}

	err = EnsureFileExists(file)
	if err != nil {
		return nil, err
	}

	fd, err := os.Open(file)
	if err != nil {
		return conf, fmt.Errorf("open config file '%s': %v", file, err)
	}
	defer fd.Close()
	err = yaml.NewDecoder(fd).Decode(&conf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			// empty file
			logrus.Warning("Creating new config file")
			err = SaveConfig(conf)
			if err != nil {
				return conf, fmt.Errorf("save empty config file")
			}
		} else {
			return conf, fmt.Errorf("read config file: %v", err)
		}
	}

	conf.Player.fillDefaults()
	return conf, nil
}

func SaveConfig(conf *Config) error {
	logrus.Debugf("Save config file")
	fd, err := os.OpenFile(conf.configFile, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer fd.Close()
	return yaml.NewEncoder(fd).Encode(conf)
}

func GetConfigDirectory() (string, error) {
	return os.UserConfigDir()
}

func EnsureConfigDirExists() error {
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dir := path.Join(userConfig, AppNameLower)
	dirExists, err := DirectoryExists(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	if dirExists {
		return nil
	} else {
		err = CreateDirectory(dir)
		return err
	}
}

func EnsureFileExists(name string) error {
	exists, err := FileExists(name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return CreateFile(name)
}

func DirectoryExists(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return true, nil
	}
	return false, fmt.Errorf("not directory")
}

func CreateDirectory(dir string) error {
	return os.Mkdir(dir, 0760)
}

func CreateFile(name string) error {
	file, err := os.Create(name)
	defer file.Close()
	if err != nil {
		return err
	}
	return nil
}

func FileExists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
