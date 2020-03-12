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

package player

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"sync"
	"time"
	"tryffel.net/go/jellycli/api"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/task"
)

type songMetadata struct {
	song          *models.Song
	album         *models.Album
	artist        *models.Artist
	albumImageUrl string
	albumImageId  string
	reader        io.ReadCloser
	format        audioFormat
}

// Player wraps all controllers and implements interfaces.QueueController, interfaces.Player and
// interfaces.ItemController.
type Player struct {
	task.Task
	*Audio
	*Queue
	*Items

	lock *sync.RWMutex

	downloadingSong bool

	songComplete   chan bool
	audioUpdated   chan interfaces.AudioStatus
	songDownloaded chan songMetadata

	api *api.Api

	lastApiReport time.Time
}

// initialize new player. This also initializes faiface.Speaker, which should be initialized only once.
func NewPlayer(api *api.Api) (*Player, error) {
	var err error
	p := &Player{
		lock:           &sync.RWMutex{},
		songComplete:   make(chan bool, 3),
		audioUpdated:   make(chan interfaces.AudioStatus, 3),
		songDownloaded: make(chan songMetadata, 3),
		api:            api,
	}
	p.Name = "Player"
	p.Task.SetLoop(p.loop)

	p.Audio = newAudio()
	p.Queue = newQueue()
	p.Items = newItems(api)
	if err != nil {
		return p, err
	}

	err = initAudio()
	if err != nil {
		return p, fmt.Errorf("init audio backend: %v", err)
	}

	p.Audio.songCompleteFunc = p.songCompleted
	p.Audio.AddStatusCallback(p.audioCallback)

	p.Queue.AddQueueChangedCallback(p.queueChanged)
	return p, nil
}

// notify song has completed
func (p *Player) songCompleted() {
	p.songComplete <- true
}

//is download pending / ongoing
func (p *Player) isDownloadingSong() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.downloadingSong
}

func (p *Player) loop() {
	// interval to refresh status. This is the interval gui will be updated.
	ticker := time.NewTicker(time.Second)

	for true {
		select {
		case <-p.StopChan():
			// stop application
			p.Audio.StopMedia()
			break
		case <-p.songComplete:
			// stream / song complete, get next song
			logrus.Debug("song complete")
			p.Queue.songComplete()
			if len(p.Queue.GetQueue()) == 0 {
				p.Audio.StopMedia()
			} else {
				p.downloadSong()
			}
		case status := <-p.audioUpdated:
			logrus.Infof("got audio status: %v", status)
		case <-ticker.C:
			// periodically update status, this will push status to p.audioUpdated
			p.Audio.updateStatus()
		case metadata := <-p.songDownloaded:
			// download complete, send to audio
			err := p.Audio.playSongFromReader(metadata)
			if err != nil {
				logrus.Errorf("play track: %v", err)
			}
		}
	}
}

// download and play next song asynchronously
func (p *Player) downloadSong() {
	if p.isDownloadingSong() || p.Queue.empty() {
		return
	}
	song := p.Queue.GetQueue()[0]

	p.lock.Lock()
	p.downloadingSong = true
	p.lock.Unlock()

	reader, err := p.api.GetSongDirect(song.Id.String(), string(audioFormatMp3))
	if err != nil {
		logrus.Errorf("download song: %v", err)
	} else {
		// fill metadata
		albumId := song.GetParent()
		album, err := p.api.GetAlbum(albumId)
		artist := models.Artist{Name: "unknown artist"}
		var imageId string
		var imageUrl string
		if err != nil {
			logrus.Error("Failed to get album by id: ", err.Error())
			album = models.Album{Name: "unknown album"}
		} else {
			imageId = album.ImageId
			imageUrl = p.api.ImageUrl(album.Id.String(), imageId)
		}
		a, err := p.api.GetArtist(album.GetParent())
		if err != nil {
			logrus.Errorf("Failed to get artist by id: %v", err)
		} else {
			artist = a
			f := func() {
				metadata := songMetadata{
					song:          song,
					album:         &album,
					artist:        &artist,
					albumImageUrl: imageUrl,
					albumImageId:  imageId,
					reader:        reader,
					format:        audioFormatMp3,
				}
				p.songDownloaded <- metadata
			}
			defer f()
		}
	}

	p.lock.Lock()
	p.downloadingSong = false
	p.lock.Unlock()

	// push song to audio
}

// Next plays next song from queue. Override Audio next to ensure there is track to play and download it
func (p *Player) Next() {
	if len(p.Queue.GetQueue()) > 1 {
		p.StopMedia()
		p.Queue.songComplete()
		go p.downloadSong()
	}
}

// Previous plays previous track. Override Audio previous to ensure there is track to play and download it
func (p *Player) Previous() {
	if len(p.Queue.GetHistory(10)) > 0 {
		p.StopMedia()
		p.Queue.playLastSong()
		p.Audio.Previous()
		go p.downloadSong()
	}
}

// report audio status to server
func (p *Player) audioCallback(status interfaces.AudioStatus) {
	p.lock.RLock()
	lastTime := p.lastApiReport
	p.lock.RUnlock()

	if time.Now().Sub(lastTime) < time.Millisecond*9500 && status.Action == interfaces.AudioActionTimeUpdate {
		// jellyfin server instructs to update every 10 sec
		return
	}

	p.lock.Lock()
	p.lastApiReport = time.Now()
	p.lock.Unlock()

	apiStatus := &interfaces.ApiPlaybackState{
		Event:          "",
		ItemId:         "",
		IsPaused:       false,
		IsMuted:        status.Muted,
		PlaylistLength: 0,
		Position:       status.SongPast.Seconds(),
		Volume:         int(status.Volume),
	}

	switch status.Action {
	case interfaces.AudioActionStop:
		apiStatus.Event = interfaces.EventStop
	case interfaces.AudioActionPlay:
		apiStatus.Event = interfaces.EventStart
	case interfaces.AudioActionNext:
		apiStatus.Event = interfaces.EventAudioTrackChange
	case interfaces.AudioActionPrevious:
		apiStatus.Event = interfaces.EventAudioTrackChange
	case interfaces.AudioActionSetVolume:
		apiStatus.Event = interfaces.EventVolumeChange
	case interfaces.AudioActionTimeUpdate:
		apiStatus.Event = interfaces.EventTimeUpdate
	case interfaces.AudioActionPlayPause:
		if status.Paused {
			apiStatus.Event = interfaces.EventPause
		} else {
			apiStatus.Event = interfaces.EventUnpause
		}
	default:
		apiStatus.Event = interfaces.EventTimeUpdate
		logrus.Warningf("cannot map audio state to api event: %v", status.Action)
	}

	apiStatus.IsPaused = status.Paused

	if status.Song != nil {
		apiStatus.ItemId = status.Song.Id.String()
		apiStatus.PlaylistLength = status.Song.Duration
	}
	f := func() {
		err := p.api.ReportProgress(apiStatus)
		if err != nil {
			logrus.Errorf("report audio progress to server: %v", err)
		}
	}
	go f()
}

func (p *Player) queueChanged(queue []*models.Song) {
	// if player has nothing to play, start download
	state := p.Audio.getStatus()
	if state.State == interfaces.AudioStateStopped && len(queue) > 0 {
		go p.downloadSong()
	}
}
