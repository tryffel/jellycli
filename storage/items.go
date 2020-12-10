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
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
)

// UpdateArtists updates/inserts artists.
func (db *Db) UpdateArtists(artists []*models.Artist) error {

	sql := `INSERT INTO artists(id, name, favorite)
	VALUES %s
	ON CONFLICT(id) DO UPDATE SET
    name=excluded.name, favorite=excluded.favorite;
`

	args := make([]interface{}, len(artists)*3)

	argFmt := ""

	for i, v := range artists {
		if i > 0 {
			argFmt += ", "
		}
		argFmt += "(?, ?, ?)"

		args[i*3] = v.Id
		args[i*3+1] = v.Name
		args[i*3+2] = v.Favorite
	}

	sql = fmt.Sprintf(sql, argFmt)

	res, err := db.engine.Exec(sql, args...)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		logrus.Infof("Updated/inserted %d artists", affected)
	}

	return nil
}

func (db *Db) GetArtists(query *interfaces.QueryOpts) (artists []*models.Artist, count int, err error) {

	stmt := db.builder.Select("*").From("artists")

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
