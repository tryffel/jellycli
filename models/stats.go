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

package models

import (
	"fmt"
	"time"
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

	StorageInfo StorageInfo
}

// HeapString returns heap usage in human-readable format
func (s *Stats) HeapString() string {
	bytes := s.Heap
	return byteToString(bytes)
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

type StorageInfo struct {
	DbSize      int
	DbFile      string
	LastUpdated time.Time
}

func (s StorageInfo) DbSizeString() string {
	return byteToString(s.DbSize)
}

func (s StorageInfo) LastUpdatedString() string {
	return s.LastUpdated.Format(time.ANSIC)

}

func byteToString(bytes int) string {
	f := float32(bytes)
	if bytes < 1024 {
		return fmt.Sprint(bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.2f KiB", f/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MiB", f/1024/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f GiB", f/1024/1024/1024)
	}
	return ""
}
