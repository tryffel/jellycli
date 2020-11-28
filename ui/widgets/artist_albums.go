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

package widgets

import (
	"fmt"
	"gitlab.com/tslocum/cview"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
	"tryffel.net/go/twidgets"
)

//AlbumCover is a simple cover for album, it shows
// album name, year and possible artists
type AlbumCover struct {
	*cview.TextView
	album   *models.Album
	index   int
	name    string
	year    int
	artists []string
}

func NewAlbumCover(index int, album *models.Album) *AlbumCover {
	a := &AlbumCover{
		TextView: cview.NewTextView(),
		album:    album,
		index:    index,
	}

	a.SetBorder(false)
	a.SetBackgroundColor(config.Color.Background)
	a.SetBorderPadding(0, 0, 1, 1)
	a.SetTextColor(config.Color.Text)
	ar := printArtists(a.artists, 40)
	text := fmt.Sprintf("%d. %s\n%d", index, album.Name, album.Year)
	if ar != "" {
		text += "\n" + ar
	}

	a.TextView.SetText(text)
	return a
}

func (a *AlbumCover) SetRect(x, y, w, h int) {
	_, _, currentW, currentH := a.GetRect()
	// todo: compact name & artists if necessary
	if currentH != h {
	}
	if currentW != w {
	}
	a.TextView.SetRect(x, y, w, h)
}

func (a *AlbumCover) SetSelected(selected twidgets.Selection) {
	switch selected {
	case twidgets.Selected:
		a.SetBackgroundColor(config.Color.BackgroundSelected)
		a.SetTextColor(config.Color.TextSelected)
	case twidgets.Blurred:
		a.SetBackgroundColor(config.Color.TextDisabled)
	case twidgets.Deselected:
		a.SetBackgroundColor(config.Color.Background)
		a.SetTextColor(config.Color.Text)
	}
}

func (a *AlbumCover) setText(text string) {
	a.TextView.SetText(text)
}

//print multiple artists
func printArtists(artists []string, maxWidth int) string {
	var out string
	need := 0
	for i, v := range artists {
		need += len(v)
		if i > 0 {
			need += 2
		}
	}

	if need > maxWidth {
		out = fmt.Sprintf("%d artists", len(artists))
		if len(out) > maxWidth {
			return ""
		} else {
			return out
		}
	}

	for i, v := range artists {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	return out
}

type ArtistAlbumList struct {
	*AlbumList

	artist *models.Artist
}

func NewArtistAlbumList(selectAlbum func(album *models.Album), context contextOperator,
	queryFunc func(opts *interfaces.QueryOpts), filterFunc openFilterFunc) *ArtistAlbumList {

	a := &ArtistAlbumList{
		AlbumList: NewAlbumList(selectAlbum, context, queryFunc, filterFunc),
		artist:    nil,
	}

	if a.context != nil {
		a.list.AddContextItem("Instant mix", 0, func(index int) {
			if index < len(a.albumCovers) && a.context != nil {
				album := a.albumCovers[index]
				a.context.InstantMix(album.album)
			}
		})
		a.options.AddOption("Show similar", func() {
			if a.similarEnabled {
				a.showSimilar()
			}
		})
		a.options.AddOption("Show in browser", func() {
			a.context.OpenInBrowser(a.artist)
		})
	}
	return a

}

// SetPlaylists sets albumList cover
func (a *ArtistAlbumList) SetArtist(artist *models.Artist) {
	a.artist = artist
	if artist != nil {
		favorite := ""
		if artist.Favorite {
			favorite = charFavorite + " "
		}

		a.description.SetText(fmt.Sprintf("%s%s\nAlbums: %d, Total: %s",
			favorite, a.artist.Name, a.artist.AlbumCount, util.SecToStringApproximate(a.artist.TotalDuration)))
	} else {
		a.description.SetText("")
	}

}

// SetPlaylist sets albums
func (a *ArtistAlbumList) SetAlbums(albums []*models.Album) {
	a.list.Clear()
	a.albumCovers = make([]*AlbumCover, len(albums))

	offset := 0
	if a.pagingEnabled {
		offset = a.page.Offset()
	}

	items := make([]twidgets.ListItem, len(albums))
	for i, v := range albums {
		cover := NewAlbumCover(offset+i+1, v)
		items[i] = cover
		a.albumCovers[i] = cover

		var artist = ""
		if len(v.AdditionalArtists) > 0 {
			artist = v.AdditionalArtists[0].Name
		}
		text := fmt.Sprintf("%d. %s\n     %s - %d", offset+i+1, v.Name, artist, v.Year)
		cover.setText(text)
	}
	a.list.AddItems(items...)
}

func (a *ArtistAlbumList) showSimilar() {
	if a.similarFunc != nil {
		a.similarFunc(a.artist.Id)
	}
}
