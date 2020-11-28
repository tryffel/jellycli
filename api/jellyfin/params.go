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
	"tryffel.net/go/jellycli/models"
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

func (p *params) setSortingByType(itemType models.ItemType, sort interfaces.Sort) {

	field := "SortName"
	order := "Ascending"

	if sort.Mode == interfaces.SortAsc {
		order = "Ascending"
	} else if sort.Mode == interfaces.SortDesc {
		order = "Descending"
	}

	switch sort.Field {
	case interfaces.SortByDate:
		field = "ProductionYear,ProductionYear,SortName"
	case interfaces.SortByName:
		field = "SortName"
		// Todo: following depend on item type
	case interfaces.SortByAlbum:
		field = "Album,SortName"
	case interfaces.SortByArtist:
		field = "Artist,SortName"
	case interfaces.SortByPlayCount:
		field = "PlayCount,SortName"
	case interfaces.SortByRandom:
		field = "Random,SortName"
	}

	p.setSorting(field, order)
}

func (p *params) setFilter(tItem models.ItemType, filter interfaces.Filter) {
	f := ""
	if filter.Favorite {
		f = appendFilter(f, "IsFavorite", ",")
	}

	// jellyfin server does not seem to like sorting artists by play status.
	// https://github.com/jellyfin/jellyfin/issues/2672
	if tItem != models.TypeArtist {
		if filter.FilterPlayed == interfaces.FilterIsPlayed {
			f = appendFilter(f, "IsPlayed", ",")
		} else if filter.FilterPlayed == interfaces.FilterIsNotPlayed {
			f = appendFilter(f, "IsUnPlayed", ",")
		}
	}

	if tItem != models.TypeArtist {
		if filter.YearRangeValid() && filter.YearRange[0] > 0 {
			years := ""
			totalYears := filter.YearRange[1] - filter.YearRange[0]
			if totalYears == 0 {
				years = strconv.Itoa(filter.YearRange[0])
			} else {
				for i := 0; i < totalYears+1; i++ {
					year := filter.YearRange[0] + i
					years = appendFilter(years, strconv.Itoa(year), ",")
				}
			}
			(*p)["Years"] = years
		}
	}

	if len(filter.Genres) > 0 {
		genres := ""
		for _, v := range filter.Genres {
			genres = appendFilter(genres, v.Name, "|")
		}
		(*p)["Genres"] = genres
	}

	if f != "" {
		(*p)["Filters"] = f
	}
}

func appendFilter(old, new string, separator string) string {
	if old == "" {
		return new
	}
	return old + separator + new
}
