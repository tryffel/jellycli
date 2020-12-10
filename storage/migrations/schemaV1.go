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

package migrations

const SchemaV1 = `

CREATE TABLE genres (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	song_count INTEGER NOT NULL,
	album_count INTEGER NOT NULL
);


CREATE TABLE artists (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	favorite BOOL NOT NULL
);

CREATE TABLE albums (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	year INTEGER NOT NULL,
	duration INTEGER NOT NULL,
	favorite BOOL NOT NULL,

	-- jellyfin sometimes returns empty artist, so don't require existing artist.
	artist TEXT NOT NULL DEFAULT ''

);


CREATE TABLE songs (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	duration INTEGER NOT NULL,
	song_index INTEGER NOT NULL,
	disc_number INTEGER NOT NULL,
	favorite bool,

	album TEXT,

	FOREIGN KEY (album) REFERENCES albums(id)
);

CREATE TABLE playlists (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE playlist_songs (
	playlist_index INTEGER NOT NULL,
	playlist TEXT,
	song TEXT,

	FOREIGN KEY (playlist) REFERENCES playlists(id),
	FOREIGN KEY (song) REFERENCES songs(id)
);


CREATE TABLE state (
	key TEXT PRIMARY KEY,
	updated INTEGER NOT NULL
);

CREATE TABLE schema (
	level INTEGER PRIMARY KEY
);

CREATE TABLE settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL DEFAULT ''
);

CREATE TABLE downloads (
	-- song id
	id TEXT NOT NULL PRIMARY KEY,
	type TEXT NOT NULL,

	path TEXT NOT NULL,
	size INTEGER

	-- timestamps
	added_at INTEGER,
	play_count INTEGER,
	last_played INTEGER
);

`
