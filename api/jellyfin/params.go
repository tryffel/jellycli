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

package jellyfin

import (
	"strconv"
	"tryffel.net/go/jellycli/interfaces"
)

type params map[string]string

// get pointer to map for convinience
func (p *params) ptr() map[string]string {
	return *p
}

func (p *params) setPaging(paging interfaces.Paging) {
	ptr := p.ptr()
	ptr["Limit"] = strconv.Itoa(paging.PageSize)
	ptr["StartIndex"] = strconv.Itoa(paging.Offset())
}

func (p *params) setLimit(n int) {
	(*p)["Limit"] = strconv.Itoa(n)
}

func (p *params) setIncludeTypes(itemType mediaItemType) {
	ptr := p.ptr()
	ptr["IncludeItemTypes"] = itemType.String()
}

func (p *params) enableRecursive() {
	(*p)["Recursive"] = "true"
}

func (p *params) setParentId(id string) {
	(*p)["ParentId"] = id
}

func (p *params) setSorting(name string, order string) {
	(*p)["SortBy"] = name
	(*p)["SortOrder"] = order
}
