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

package util

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"
	"tryffel.net/go/jellycli/config"
)

// DumpGoroutines dumps all goroutines into a file in same directory as log file, with timestamped name.
func DumpGoroutines() error {
	buf := make([]byte, 1024*1024)
	runtime.Stack(buf, true)
	//remove unused bytes
	buf = bytes.TrimRight(buf, "\x00")

	dir := path.Dir(config.AppConfig.Player.LogFile)
	now := time.Now()
	fileName := fmt.Sprintf("jellycli-dump_%d-%d-%d.%d-%d-%d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())

	file, err := os.Create(path.Join(dir, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("Jellycli version " + config.Version + " dump at " + now.String() + "\n\n")
	_, err = file.Write(buf)

	file.WriteString("\n")
	return err
}
