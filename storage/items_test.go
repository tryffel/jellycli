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
	"github.com/google/go-cmp/cmp"
	"testing"
	"tryffel.net/go/jellycli/api"
	"tryffel.net/go/jellycli/interfaces"
)

func TestDb_UpdateArtists(t *testing.T) {
	artists := api.MockArtists

	for i, _ := range artists {
		artists[i].Albums = nil
	}

	if len(artists) == 0 {
		t.Errorf("no artists")
		return
	}

	db := testDb(t)
	if db == nil {
		return
	}

	defer closeDb(t, db)

	err := db.UpdateArtists(artists[1:])
	if err != nil {
		t.Errorf("insert artists: %v", err)
	}

	err = db.UpdateArtists(artists)
	if err != nil {
		t.Errorf("update artists: %v", err)
	}

	gotArtists, count, err := db.GetArtists(interfaces.DefaultQueryOpts())
	if err != nil {
		t.Errorf("get artists: %v", err)
	}

	if count != len(artists) {
		t.Errorf("invalid artists count: %d, want: %d", count, len(artists))
	}

	diff := cmp.Diff(artists, gotArtists)

	if diff != "" {
		t.Errorf("artists differ: %s", diff)
	}
}

func TestDb_UpdateAlbums(t *testing.T) {
	albums := api.MockAlbums

	for i, _ := range albums {
		albums[i].Songs = nil
	}

	if len(albums) == 0 {
		t.Errorf("no albums")
		return
	}

	db := testDb(t)
	if db == nil {
		return
	}

	defer closeDb(t, db)

	err := db.UpdateAlbums(albums[1:])
	if err != nil {
		t.Errorf("insert albums: %v", err)
	}

	err = db.UpdateAlbums(albums)
	if err != nil {
		t.Errorf("update albums: %v", err)
	}

	gotAlbums, count, err := db.GetAlbums(interfaces.DefaultQueryOpts())
	if err != nil {
		t.Errorf("get albums: %v", err)
	}

	if count != len(albums) {
		t.Errorf("invalid albums count: %d, want: %d", count, len(albums))
	}

	diff := cmp.Diff(albums, gotAlbums)

	if diff != "" {
		t.Errorf("albums differ: %s", diff)
	}
}

func TestDb_UpdateSongs(t *testing.T) {
	songs := api.MockSongs

	if len(songs) == 0 {
		t.Errorf("no songs")
		return
	}

	for i, _ := range songs {
		songs[i].AlbumArtist = ""
	}

	db := testDb(t)
	if db == nil {
		return
	}

	defer closeDb(t, db)

	err := db.UpdateSongs(songs[1:])
	if err != nil {
		t.Errorf("insert songs: %v", err)
	}

	err = db.UpdateSongs(songs)
	if err != nil {
		t.Errorf("update songs: %v", err)
	}

	gotSongs, count, err := db.GetSongs(0, 10)
	if err != nil {
		t.Errorf("get songs: %v", err)
	}

	if count != len(songs) {
		t.Errorf("invalid songs count: %d, want: %d", count, len(songs))
	}

	diff := cmp.Diff(songs, gotSongs)

	if diff != "" {
		t.Errorf("songs differ: %s", diff)
	}
}

func TestDb_UpdatePlaylists(t *testing.T) {
	playlists := api.MockPlaylists

	if len(playlists) == 0 {
		t.Errorf("no playlists")
		return
	}

	db := testDb(t)
	if db == nil {
		return
	}

	defer closeDb(t, db)

	err := db.UpdateSongs(api.MockSongs)

	err = db.UpdatePlaylists(playlists[1:])
	if err != nil {
		t.Errorf("insert playlists: %v", err)
	}

	err = db.UpdatePlaylists(playlists)
	if err != nil {
		t.Errorf("update playlists: %v", err)
	}

	gotPlaylists, err := db.GetPlaylists()
	if err != nil {
		t.Errorf("get playlists: %v", err)
	}

	if len(gotPlaylists) != len(playlists) {
		t.Errorf("invalid playlists count returned")
	}

	for i, v := range gotPlaylists {
		v.Songs = playlists[i].Songs
		diff := cmp.Diff(v, gotPlaylists[i])
		if diff != "" {
			t.Errorf("playlist differs: %s", diff)
		}
	}
}
