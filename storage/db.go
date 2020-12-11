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

// Package local implements having local cache
// of remote serve relational data (artist,album,playlist,song).
package storage

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"strings"
	"time"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/storage/migrations"
)

const schemaLevel = 1

// Db implements storing relational data to local database as cache.
// Schema reflects the data coming from server and tries to store updated content
// and not enforce relational integrity.
// This should also mean getting data from db and server
// should return same results for same queries.
type Db struct {
	id   string
	file string

	builder squirrel.StatementBuilderType
	engine  *sqlx.DB
}

func NewDb(id string) (*Db, error) {
	err := os.Mkdir(config.AppConfig.Player.LocalCacheDir, 0760)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
		} else {
			return nil, fmt.Errorf("create cache dir: %v", err)
		}
	}

	fileName := path.Join(config.AppConfig.Player.LocalCacheDir, id+".db")
	logrus.Debugf("use cache: %s", fileName)

	db := &Db{
		id:      id,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question),
		file:    fileName,
	}

	db.engine, err = sqlx.Connect("sqlite3", fmt.Sprintf("file:%s?_fk=true&_cslike=false", fileName))
	if err != nil {
		return db, err
	}

	err = db.checkSchema()
	if err != nil {
		if strings.Contains(err.Error(), "schema is invalid") {
			err = db.initDb()
		} else {
			return db, err
		}
	}
	return db, err
}

// create file + schema
func (db *Db) initDb() error {

	tx, err := db.engine.Begin()
	if err != nil {
		return err
	}

	txOk := false

	commit := func() {
		if txOk {
			err = tx.Commit()
		} else {
			err = tx.Rollback()
		}

		if err != nil {
			logrus.Fatalf("end tx: %v", err)
		}
	}

	defer commit()

	_, err = tx.Exec(migrations.SchemaV1)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO schema VALUES (1)")
	if err == nil {
		txOk = true
	}
	return err
}

func (db *Db) checkSchema() error {

	sql := "SELECT * FROM schema;"

	schema := -1
	err := db.engine.Get(&schema, sql)
	if err != nil {

		if strings.Contains(err.Error(), "no such table") {
			return errors.New("schema is invalid: empty schema")
		}
		return err
	}

	if schema != schemaLevel {
		return fmt.Errorf("database schema is invalid: supported %d, database: %d", schemaLevel, schema)
	}

	return nil
}

func (db *Db) Close() error {
	return db.engine.Close()
}

// database transaction. Start new with db.Begin,
// st tx.ok = true to commit transaction and call tx.Close().
type tx struct {
	*sqlx.Tx
	ok bool
}

func (tx *tx) Close() error {
	if tx.ok {
		return tx.Commit()
	} else {
		return tx.Rollback()
	}
}

func (db *Db) begin() (*tx, error) {
	txX, err := db.engine.Beginx()
	if err != nil {
		return nil, err
	}
	tx := &tx{
		Tx: txX,
		ok: false,
	}

	return tx, nil
}

func (db *Db) GetStats() (models.StorageInfo, error) {
	info := models.StorageInfo{
		DbSize:      0,
		DbFile:      db.file,
		LastUpdated: time.Time{},
	}

	sql := "select updated from state order by updated asc limit 1;"
	t := sqlTime{}
	err := db.engine.Get(&t, sql)
	if err != nil {
		return info, err
	}

	info.LastUpdated = t.Time
	fileStat, err := os.Stat(db.file)
	if err != nil {
		return info, err
	}
	info.DbSize = int(fileStat.Size())
	return info, nil
}

type sqlTime struct {
	Time time.Time
}

func (s *sqlTime) Scan(src interface{}) error {
	srcInt, ok := src.(int64)
	if ok {
		s.Time = time.Unix(0, srcInt)
	} else {
		return fmt.Errorf("not time: %v", src)
	}
	return nil
}

func (s sqlTime) Value() (driver.Value, error) {
	return s.Time.UnixNano(), nil
}
