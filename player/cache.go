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

package player

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
	"tryffel.net/go/jellycli/api"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

// UpdateLocalArtists pulls latest info on artists from server and stores/updates on local database.
func (i *Items) UpdateLocalArtists(limit int) error {
	logrus.Debugf("Refresh artists from remote server")
	start := time.Now()
	retrieved := 0
	totalArtists := 0

	query := interfaces.DefaultQueryOpts()
	query.Paging.PageSize = 200
	if 0 < limit && limit < 200 {
		query.Paging.PageSize = limit
	}
	query.Paging.CurrentPage = 0

	for {
		artists, n, err := i.browser.GetArtists(query)
		if err != nil {
			return fmt.Errorf("pull artists: %v", err)
		}
		totalArtists = n

		if len(artists) == 0 {
			logrus.Debugf("no artists found")
			return nil
		}

		err = i.db.UpdateArtists(artists)
		if err != nil {
			return fmt.Errorf("save artists: %v", err)
		}
		query.Paging.CurrentPage += 1
		retrieved += len(artists)

		logrus.Debugf("retrieved %d artists", retrieved)

		if (limit != 0 && retrieved >= limit) || len(artists) < query.Paging.PageSize || totalArtists <= retrieved {
			break
		}
	}

	if totalArtists != retrieved {
		logrus.Warningf("not all artists were updated: (%d - %d)", totalArtists, retrieved)

	}

	took := time.Now().Sub(start)

	logrus.Infof("Updated %d artists in %.2f s", retrieved, float32(took.Milliseconds())/1000)
	return nil
}

// UpdateLocalAlbums pulls latest info on albums from server and stores/updates on local database.
func (i *Items) UpdateLocalAlbums(limit int) error {
	logrus.Debugf("Refresh album from remote server")
	start := time.Now()
	retrieved := 0
	totalAlbums := 0

	query := interfaces.DefaultQueryOpts()
	query.Paging.PageSize = 100
	if 0 < limit && limit < 100 {
		query.Paging.PageSize = limit
	}
	query.Paging.CurrentPage = 0

	for {
		albums, n, err := i.browser.GetAlbums(query)
		if err != nil {
			return fmt.Errorf("pull albums: %v", err)
		}
		totalAlbums = n

		if len(albums) == 0 {
			logrus.Debugf("no albums found")
			return nil
		}

		err = i.db.UpdateAlbums(albums)
		if err != nil {
			return fmt.Errorf("save albums: %v", err)
		}

		query.Paging.CurrentPage += 1
		retrieved += len(albums)

		if (limit != 0 && retrieved >= limit) || len(albums) < query.Paging.PageSize {
			break
		}
	}

	if totalAlbums != retrieved {
		logrus.Warningf("not all albums were updated: (%d - %d)", totalAlbums, retrieved)

	}

	took := time.Now().Sub(start)

	logrus.Infof("Updated %d albums in %.2f s", retrieved, float32(took.Milliseconds())/1000)
	return nil

}

func (i *Items) UpdateLocalSongs(limit int) error {
	logrus.Debugf("Refresh songs from remote server")

	pullSongsDirectly := false

	cacher, ok := i.browser.(api.Cacher)
	if ok {
		if cacher.CanCacheSongs() {
			pullSongsDirectly = true
		}
	}

	start := time.Now()
	var err error

	if pullSongsDirectly {
		err = i.pullSongs(limit)
	} else {
		err = i.pullSongsByAlbums(limit)
	}

	took := time.Now().Sub(start)

	if err == nil {
		logrus.Infof("Updated songs in %.2f s", float32(took.Milliseconds())/1000)
	}
	return err
}

func (i *Items) pullSongs(limit int) error {
	logrus.Infof("Pull songs from server")
	retrieved := 0
	totalSongs := 0
	query := interfaces.DefaultQueryOpts()
	query.Paging.PageSize = 100
	if 0 < limit && limit < 100 {
		query.Paging.PageSize = limit
	}
	query.Paging.CurrentPage = 0

	for {
		songs, n, err := i.browser.GetSongs(query.Paging.CurrentPage, query.Paging.PageSize)
		if err != nil {
			return fmt.Errorf("pull songs: %v", err)
		}
		totalSongs = n

		if len(songs) == 0 {
			logrus.Debugf("no songs found")
			return nil
		}

		err = i.db.UpdateSongs(songs)
		if err != nil {
			return fmt.Errorf("save songs: %v", err)
		}

		query.Paging.CurrentPage += 1
		retrieved += len(songs)

		if (limit != 0 && retrieved >= limit) || len(songs) < query.Paging.PageSize {
			break
		}
	}

	if totalSongs != retrieved {
		logrus.Warningf("not all songs were updated: (%d - %d)", totalSongs, retrieved)
	}

	return nil

}

func (i *Items) pullSongsByAlbums(limit int) error {
	logrus.Infof("Pull album songs from server")

	totalAlbums := 0
	retrieved := 0
	totalSongs := 0
	failed := 0
	query := interfaces.DefaultQueryOpts()
	query.Paging.PageSize = 100
	if 0 < limit && limit < 100 {
		query.Paging.PageSize = limit
	}
	query.Paging.CurrentPage = 0

	var err error
	var albums []*models.Album

	albums, totalAlbums, err = i.GetAlbums(query)

	totalPages := totalAlbums/query.Paging.PageSize + 1

	for page := 0; page < totalPages; page++ {
		logrus.Debugf("get albums, page %d", page)

		if page > 0 {
			query.Paging.CurrentPage += 1
			albums, _, err = i.GetAlbums(query)
			if err != nil {
				logrus.Error("get albums: %v", err)
				continue
			}
		}
		for _, album := range albums {
			songs, err := i.browser.GetAlbumSongs(album.Id)
			if err != nil {
				logrus.Errorf("get album songs: %v", err)
				failed += 1
				continue
			} else {
				totalSongs += len(songs)
				retrieved += 1
			}

			if len(songs) == 0 {
				logrus.Debugf("no songs found")
			}

			err = i.db.UpdateSongs(songs)
			if err != nil {
				return fmt.Errorf("save songs: %v", err)
			}
		}
	}

	logrus.Infof("Cached songs for %d albums, %d failed", retrieved, failed)
	return err
}
