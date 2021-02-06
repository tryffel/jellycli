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
	"path"
	"testing"
)

func testDb(t *testing.T) *Db {
	id := "test-123"
	dir := t.TempDir()
	file := path.Join(dir, id+".db")

	db, err := newDb(file, id)
	if err != nil {
		t.Errorf("init db: %v", err)
		return nil
	}
	return db
}

func closeDb(t *testing.T, db *Db) {
	err := db.Close()
	if err != nil {
		t.Errorf("close db: %v", err)
	}
}

func TestNewDb(t *testing.T) {
	db := testDb(t)
	closeDb(t, db)
}
