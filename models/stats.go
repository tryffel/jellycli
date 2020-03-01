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

import "fmt"

// Stats are application-wide statistics and usage info
type Stats struct {
	// Heap describes memory allocation in bytes
	Heap int
	// CacheObjects tells how many items are in cache at the moment
	CacheObjects int
	// ServerName is jellyfin server name
	ServerName string
	// ServerVersion server version
	ServerVersion string
	// WebSocket boolean if websocket is supported and connected
	WebSocket bool
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
