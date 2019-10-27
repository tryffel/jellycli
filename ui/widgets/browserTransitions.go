/*
 * Copyright 2019 Tero Vierimaa
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package widgets

import (
	"fmt"
	"tryffel.net/pkg/jellycli/models"
)

func (b *Browser) transitionEnter(split panelSplit) error {
	// Show children if there's any
	b.lock.Lock()
	defer b.lock.Unlock()

	_, item := b.getSelectedItem(split)
	if item.GetType() == models.TypeSong {
		// No children
		return nil
	}

	if !item.HasChildren() {
		return fmt.Errorf("item has no children")
	}
	if split == panelR {
		b.listR.Clear()
		b.dataL = b.dataR
		b.listL.SetData(b.dataL)
		b.panelAwaiting = split
	} else {
		b.panelAwaiting = split.Other()
	}

	go b.controller.GetChildren(item.GetId(), item.GetType())
	return nil
}

func (b *Browser) transitionBack(split panelSplit) error {
	// Show parent if there's any
	b.lock.Lock()
	defer b.lock.Unlock()

	_, item := b.getSelectedItem(split)
	if item.GetType() == models.TypeSong {
		// No children
		return nil
	}

	if !item.HasChildren() {
		return fmt.Errorf("item has no children")
	}
	if split == panelL {
		b.listL.Clear()
		b.dataR = []models.Item{}
		copy(b.dataR, b.dataL)
		b.listR.SetData(b.dataR)
	} else {
		b.panelAwaiting = split.Other()
	}

	go b.controller.GetChildren(item.GetId(), item.GetType())
	return nil
}

func (b *Browser) transitionArtistShowAlbums(split panelSplit) error {
	// Move right panel content to left panel, show albums in right
	b.lock.Lock()
	defer b.lock.Unlock()

	_, item := b.getSelectedItem(split)
	if !item.HasChildren() {
		return fmt.Errorf("item has no children")
	}

	b.panelAwaiting = panelR
	go b.controller.GetChildren(item.GetId(), item.GetType())

	// 1. move content from right to left panel (if needed)
	// 2. set right panel pending
	// 3. request new data
	// 4. somewhere else: apply data to right panel

	return nil
}

func (b *Browser) transitionAlbumShowSongs(split panelSplit) error {
	// If split == panelR, move panelR_data -> panelL_data
	// Get songs to panelR

	return nil
}

func (b *Browser) transitionSongsShowAlbums(split panelSplit) error {
	// if split == panelR, move panelL_data -> panel_R_data
	// Get artists to panelL
	return nil
}

func (b *Browser) transitionAlbumsShowArtists(split panelSplit) error {
	// if split == panelR, mover panelL_data -> panel_r_data
	// get artists to panelL

	return nil
}
