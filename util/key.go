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
	"github.com/gdamore/tcell"
	"strings"
)

//ShortCutName returns name for given key
func KeyBindingName(key tcell.Key) string {
	return tcell.KeyNames[key]
}

//PackKeyBindingName returns shorter for given key
// Maxlength controls maximum length for text.
// If 0, disable limiting
// 'F6' -> F6
// 'Ctrl+Space' -> 'C-sp'
func PackKeyBindingName(key tcell.Key, maxLength int) string {
	name := KeyBindingName(key)
	if maxLength == 0 {
		return name
	}
	if strings.Contains(name, "Ctrl") {
		name = strings.TrimPrefix(name, "Ctrl-")
		name = "C-" + name
	}
	return name
}
