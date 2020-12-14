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

package storage

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

const (
	keyAlbums  = "albums"
	keyArtists = "artists"
	keySongs   = "songs"
)

func (db *Db) updateKey(key string, tx *tx) error {
	sql := `INSERT INTO state (key, updated) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET updated=excluded.updated;`

	_, err := tx.Exec(sql, key, sqlTime{time.Now()})
	if err != nil {
		return err
	}
	return nil
}

// UpdateArtists updates/inserts artists.
func (db *Db) UpdateArtists(artists []*models.Artist) error {

	sql := `INSERT INTO artists(id, name, favorite, total_duration, album_count)
	VALUES %s
	ON CONFLICT(id) DO UPDATE SET
    name=excluded.name, favorite=excluded.favorite,
	total_duration=excluded.total_duration,
	album_count=excluded.album_count;
`

	args := make([]interface{}, len(artists)*5)

	argFmt := ""

	for i, v := range artists {
		if i > 0 {
			argFmt += ", "
		}
		argFmt += "(?, ?, ?, ?, ?)"

		args[i*5] = v.Id
		args[i*5+1] = v.Name
		args[i*5+2] = v.Favorite
		args[i*5+3] = v.TotalDuration
		args[i*5+4] = v.AlbumCount
	}

	sql = fmt.Sprintf(sql, argFmt)

	tx, err := db.begin()
	if err != nil {
		return err
	}
	defer tx.Close()

	res, err := tx.Exec(sql, args...)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		logrus.Infof("Updated/inserted %d artists", affected)
	}

	err = db.updateKey(keyArtists, tx)
	if err != nil {
		return err
	}
	tx.ok = true
	return nil
}

func (db *Db) GetArtists(query *interfaces.QueryOpts) (artists []*models.Artist, count int, err error) {

	stmt := db.builder.
		Select("*").From("artists")

	if !query.Filter.Empty() {
		if query.Filter.Favorite {
			stmt = stmt.Where("favorite = TRUE")
		}
	}

	if query.Sort.Field != "" {
		mode := query.Sort.Mode
		switch query.Sort.Field {
		case interfaces.SortByName:
			stmt = stmt.OrderBy("name " + mode)
		case interfaces.SortByRandom:
			stmt = stmt.OrderBy("RANDOM()")
		default:
			stmt = stmt.OrderBy("name " + mode)
		}
	}

	stmt = stmt.Offset(uint64(query.Paging.Offset()))
	stmt = stmt.Limit(uint64(query.Paging.PageSize))

	var sql string
	var args []interface{}

	sql, args, err = stmt.ToSql()
	if err != nil {
		return
	}

	a := &[]models.Artist{}

	err = db.engine.Select(a, sql, args...)
	if err != nil {
		return
	}

	artists = make([]*models.Artist, len(*a))
	for i, _ := range *a {
		artists[i] = &(*a)[i]
	}

	sql = "SELECT COUNT(id) FROM ARTISTS"
	err = db.engine.Get(&count, sql)
	return
}

func (db *Db) UpdateAlbums(albums []*models.Album) error {
	sql := `INSERT INTO albums(id, name, year, duration, favorite, artist, song_count, image_id, disc_count)
	VALUES %s
	ON CONFLICT(id) DO UPDATE SET
    name=excluded.name, favorite=excluded.favorite,
	year=excluded.year, duration=excluded.duration,
	artist=excluded.artist, song_count=excluded.song_count,
	image_id=excluded.image_id, disc_count=excluded.disc_count;
`

	args := make([]interface{}, len(albums)*9)

	argFmt := ""

	for i, v := range albums {
		if i > 0 {
			argFmt += ", "
		}
		argFmt += "(?, ?, ?, ?, ?, ?, ?, ?, ?)"

		args[i*9] = v.Id
		args[i*9+1] = v.Name
		args[i*9+2] = v.Year

		args[i*9+3] = v.Duration
		args[i*9+4] = v.Favorite
		args[i*9+5] = v.Artist

		args[i*9+6] = v.SongCount
		args[i*9+7] = v.ImageId
		args[i*9+8] = v.DiscCount
	}

	sql = fmt.Sprintf(sql, argFmt)

	tx, err := db.begin()
	if err != nil {
		return err
	}
	defer tx.Close()

	_, err = tx.Exec(sql, args...)
	if err != nil {
		return err
	}

	err = db.updateKey(keyAlbums, tx)
	if err != nil {
		return err
	}
	tx.ok = true
	return nil
}

func (db *Db) GetAlbums(query *interfaces.QueryOpts) (albums []*models.Album, n int, err error) {
	stmt := db.builder.
		Select("*").From("albums")

	if !query.Filter.Empty() {
		if query.Filter.Favorite {
			stmt = stmt.Where("favorite = TRUE")
		}
	}

	if query.Sort.Field != "" {
		mode := query.Sort.Mode
		switch query.Sort.Field {
		case interfaces.SortByName:
			stmt = stmt.OrderBy("name " + mode)
		case interfaces.SortByRandom:
			stmt = stmt.OrderBy("RANDOM()")
		default:
			stmt = stmt.OrderBy("name " + mode)
		}
	}

	stmt = stmt.Offset(uint64(query.Paging.Offset()))
	stmt = stmt.Limit(uint64(query.Paging.PageSize))

	var sql string
	var args []interface{}

	sql, args, err = stmt.ToSql()
	if err != nil {
		return
	}

	a := &[]models.Album{}

	err = db.engine.Select(a, sql, args...)
	if err != nil {
		return
	}

	albums = make([]*models.Album, len(*a))
	for i, _ := range *a {
		albums[i] = &(*a)[i]
	}

	sql = "SELECT COUNT(id) FROM albums"
	err = db.engine.Get(&n, sql)
	return
}

func (db *Db) UpdateSongs(songs []*models.Song) error {
	sql := `INSERT INTO songs(id, name, duration, song_index, disc_number, favorite, album)
	VALUES %s
	ON CONFLICT(id) DO UPDATE SET
    name=excluded.name, duration=excluded.duration,
	song_index=excluded.song_index, disc_number=excluded.disc_number,
	favorite=excluded.favorite, album=excluded.album;
`

	args := make([]interface{}, len(songs)*7)

	argFmt := ""

	for i, v := range songs {
		if i > 0 {
			argFmt += ", "
		}
		argFmt += "(?, ?, ?, ?, ?, ?, ?)"

		args[i*7] = v.Id
		args[i*7+1] = v.Name
		args[i*7+2] = v.Duration

		args[i*7+3] = v.Index
		args[i*7+4] = v.DiscNumber
		args[i*7+5] = v.Favorite
		args[i*7+6] = v.Album
	}

	sql = fmt.Sprintf(sql, argFmt)

	tx, err := db.begin()
	if err != nil {
		return err
	}
	defer tx.Close()

	_, err = tx.Exec(sql, args...)
	if err != nil {
		return err
	}

	err = db.updateKey(keySongs, tx)
	if err != nil {
		return err
	}
	tx.ok = true
	return nil
}

func (db *Db) GetSongs(page int, pageSize int) ([]*models.Song, int, error) {
	stmt := db.builder.
		Select("*").From("songs")

	stmt = stmt.Offset(uint64(page * pageSize))
	stmt = stmt.Limit(uint64(pageSize))
	stmt = stmt.OrderBy("name")

	var sql string
	var args []interface{}

	sql, args, err := stmt.ToSql()
	if err != nil {
		return nil, 0, err
	}

	s := &[]models.Song{}

	err = db.engine.Select(s, sql, args...)
	if err != nil {
		return nil, 0, err
	}

	songs := make([]*models.Song, len(*s))
	for i, _ := range *s {
		songs[i] = &(*s)[i]
	}

	count := 0

	sql = "SELECT COUNT(id) FROM songs"
	err = db.engine.Get(&count, sql)
	return songs, count, nil
}
