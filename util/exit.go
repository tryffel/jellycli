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
	"github.com/sirupsen/logrus"
)

// Exit logs exit message to log and calls os.exit. This function can be overridden for testing purposes.
// LogrusInstance allows overriding default instance to pass additional arguments e.g. with
// logrus.WithField. It can also be set to nil.
var Exit = func(logrusInstance *logrus.Entry, msg string) {
	println("Fatal error, see log file")
	if logrusInstance != nil {
		logrusInstance.Fatalf(msg)
	} else {
		logrus.Fatal(msg)
	}
}
