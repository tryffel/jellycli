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

package models

import (
	"fmt"
)

// Stats are application-wide statistics and usage info
type Stats struct {
	// Heap describes memory allocation in bytes
	Heap int

	//LogFile contains log file location
	LogFile string

	// ConfigFile contains config file location
	ConfigFile string

	// ServerInfo contains remote server information
	ServerInfo *ServerInfo
}

// HeapString returns heap usage in human-readable format
func (s *Stats) HeapString() string {
	bytes := s.Heap
	if bytes < 1024 {
		return fmt.Sprint(bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%d KiB", bytes/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%d MiB", bytes/1024/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%d GiB", bytes/1024/1024/1024)
	}
	return ""
}

// ServerInfo contains general info on server and connection to it.
type ServerInfo struct {
	// ServerType describes protocol/server type.
	ServerType string
	// Name is server instance name, if it has one.
	Name string

	// Id is server instance id, if it has one.
	Id string

	// Version is server version.
	Version string

	// Message contains server message, if any.
	Message string

	// Misc contains any non-standard information, that use might be interested in.
	Misc map[string]string
}
