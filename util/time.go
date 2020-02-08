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
