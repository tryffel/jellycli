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
