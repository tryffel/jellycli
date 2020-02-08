/*
 * Copyright 2020 Tero Vierimaa
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
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"tryffel.net/pkg/jellycli/models"
	"tryffel.net/pkg/jellycli/util"
)

type Artist struct {
	*InfoList
	artist          *models.Artist
	albums          []*models.Album
	selectAlbumFunc func(id models.Id)
}

func NewArtist(selectAlbumFunc func(id models.Id)) *Artist {
	a := &Artist{
		InfoList:        NewInfoList(tview.NewList()),
		selectAlbumFunc: selectAlbumFunc,
	}
	list := a.list.(*tview.List)
	list.ShowSecondaryText(false)
	list.SetSelectedFunc(a.selectArtist)
	return a
}

func (a *Artist) SetArtist(ar *models.Artist) {
	a.artist = ar
	text := ar.Name + "\n"
	text += fmt.Sprintf("Total: %s\n", util.SecToString(ar.TotalDuration))
	a.SetInfoText(text)
}

func (a *Artist) SetAlbums(albums []*models.Album) {
	a.albums = albums
	list := a.list.(*tview.List)
	list.Clear()

	for i, v := range albums {
		text := fmt.Sprintf("%d. %s", i, v.Name)
		list.AddItem(text, "", 0, nil)
	}
}

func (a *Artist) selectArtist(index int, text, secondaryText string, shortCut rune) {
	if len(a.albums) == 0 {
		return
	}
	if index >= len(a.albums) {
		logrus.Debug("selected artist album index > num of albums")
	}

	id := a.albums[index].Id
	if a.selectAlbumFunc != nil {
		a.selectAlbumFunc(id)
	}
}
