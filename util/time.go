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

import "fmt"

// Print seconds as formatted time:
// 50, 1:50,
// 0:05, 1.05, 1:05:05
func SecToString(sec int) string {
	if sec < 60 {
		return fmt.Sprintf("0:%02d", sec)
	}
	minutes := sec / 60
	if sec < 3600 {
		return fmt.Sprintf("%d:%02d", minutes, sec%60)
	} else {
		hours := sec / 3600
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes-60*hours, sec%3600%60)
	}
}

//SecToStringApproximate gives more verbal output for time duration
// 1:50 -> 2 min
// 10:50 -> 11 min
// 1:11:50 -> 1.1 h
func SecToStringApproximate(sec int) string {
	//if sec < 600 {
	//	return SecToString(sec)
	//}
	minutes := sec / 60
	if sec < 3600 {
		if minutes > 1 {
			return fmt.Sprintf("%d mins", minutes)
		} else {
			return fmt.Sprintf("%d min", minutes)
		}
	} else {
		hours := sec / 3600
		minutes = sec/60 - hours*60
		var hour = "hour"
		var minute = "min"
		if hours > 1 {
			hour = "hours"
		}
		if minutes > 1 {
			minute = "mins"
		}
		return fmt.Sprintf("%d %s %d %s", hours, hour, minutes, minute)
	}
}
