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
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

// NewConfigFile creates new config file in given location.
// If locatin is empty, use default user directory ~/.config/jellycli.
func NewConfigFile(location string) error {
	var err error
	var dir string
	var file string

	if location != "" {
		dir, file = path.Split(location)
	} else {
		var err error
		dir, err = os.UserConfigDir()
		if err != nil {
			return err
		}
		file = "jellycli.yaml"
		location = path.Join(dir, file)
	}

	logrus.Warningf("Create new config file %s", location)

	if dir != "" {
		err = ensureConfigDirExists(dir)
		if err != nil {
			return err
		}
	}

	err = ensureFileExists(file)
	if err != nil {
		return err
	}

	fd, err := os.Create(location)
	if err != nil {
		return fmt.Errorf("create config file '%s': %v", file, err)
	}
	return fd.Close()
}

func ensureConfigDirExists(dir string) error {
	dirExists, err := directoryExists(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	if dirExists {
		return nil
	} else {
		err = createDirectory(dir)
		return err
	}
}

func ensureFileExists(name string) error {
	exists, err := fileExists(name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return createFile(name)
}

func directoryExists(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return true, nil
	}
	return false, fmt.Errorf("not directory")
}

func createDirectory(dir string) error {
	return os.Mkdir(dir, 0760)
}

func createFile(name string) error {
	file, err := os.Create(name)
	defer file.Close()
	if err != nil {
		return err
	}
	return nil
}

func fileExists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
