package player

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"time"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/task"
)

//Action includes instruction for player to play
type Action struct {
	// What to do
	State  interfaces.State
	Type   interfaces.Playtype
	Volume int

	// Provide either artist/album/song or audio id
	AudioId  string
	Duration int

	// Metadata, only used when playing new song
	Song   *models.Song
	Artist *models.Artist
	Album  *models.Album
}

type PlaySong struct {
	Action Action
	Song   io.ReadCloser
}

const (
	// How often to update state i.e. push status to playingstate channel
	updateInterval = time.Second
)

// Player is the application structure
type Player struct {
	task.Task

	Api interfaces.Api
	// chanAction is for user interactions
	chanAction chan Action
	// chanState is updated when state is changed
	chanState chan interfaces.PlayingState

	chanStreamComplete chan bool

	chanAddSong chan PlaySong

	ticker *time.Ticker

	state      interfaces.PlayingState
	lastAction *Action

	audio  *audio
	reader io.ReadCloser
	itemId string
}

// NewPlayer constructs new player instance
func NewPlayer(a interfaces.Api) (*Player, error) {
	p := &Player{
		Api:                a,
		chanAction:         make(chan Action, 3),
		chanState:          make(chan interfaces.PlayingState, 3),
		chanStreamComplete: make(chan bool, 3),
		chanAddSong:        make(chan PlaySong, 3),
		ticker:             nil,
		audio:              nil,
	}
	p.audio = newAudio(p.chanStreamComplete)
	err := initAudio()
	if err != nil {
		return p, fmt.Errorf("audio init failed: %v", err)
	}
	p.SetLoop(p.loop)
	p.Name = "AudioPlayer"

	p.audio.pause(true)
	p.state.State = interfaces.Stop
	p.state.Volume = 50
	return p, nil
}

//ActionChannel returns input channel for user actions
func (p *Player) ActionChannel() chan Action {
	return p.chanAction
}

//StateChannel return output channel for player state
func (p *Player) StateChannel() chan interfaces.PlayingState {
	return p.chanState
}

func (p *Player) AddSongChannel() *chan PlaySong {
	return &p.chanAddSong
}

func (p *Player) loop() {
	p.ticker = time.NewTicker(updateInterval)
	defer p.ticker.Stop()
	chunkStarted := time.Now()
	for true {
		select {
		case <-p.ticker.C:
			diff := time.Since(chunkStarted).Seconds()
			if diff >= 10 {
				if p.state.State == interfaces.Play || p.state.State == interfaces.Pause {
					go p.reportStatus(interfaces.EventTimeUpdate)
				}
				// Query new chunk every 10 sec
				chunkStarted = time.Now()
			}
			at := p.audio.timePast()
			p.state.CurrentSongPast = int(at.Seconds())
			p.RefreshState()
		case action := <-p.chanAction:
			// User has requested action
			if ok, event := p.handleAction(action); ok {
				p.lastAction = &action
				p.RefreshState()
				go p.reportStatus(event)
			} else {
				logrus.Error("Invalid action, probably incorrect transition")
			}
		case <-p.chanStreamComplete:
			p.endStream()
		case <-p.StopChan():
			// Program is stopping
			p.stop()
			break
		case song := <-p.chanAddSong:
			p.playSongFromReader(song)

		}
	}
}

//handle any incoming actions. Return true if state has changed
func (p *Player) handleAction(action Action) (bool, interfaces.ApiPlaybackEvent) {
	defaultEvent := interfaces.EventTimeUpdate
	switch action.State {
	case interfaces.SetVolume:
		if p.state.Volume != action.Volume && action.Volume != -1 {
			if action.Volume > config.AudioMaxVolume {
				action.Volume = config.AudioMaxVolume
			} else if action.Volume < config.AudioMinVolume {
				action.Volume = config.AudioMinVolume
			}
			p.audio.setVolume(action.Volume)
			p.state.Volume = action.Volume
			go p.reportStatus(interfaces.EventVolumeChange)
			return true, interfaces.EventVolumeChange
		}
	case interfaces.Pause:
		if p.state.State == interfaces.Play && p.audio.hasStreamer() {
			p.audio.pause(true)
			p.state.State = interfaces.Pause
			return true, interfaces.EventPause
		}
		return false, defaultEvent
	case interfaces.Play:
		if p.state.State == interfaces.Stop || p.state.State == interfaces.Pause || p.state.State == interfaces.Play {
			if p.PlaySong(action) {
				return true, interfaces.EventPlaylistItemAdd
			}
			return false, defaultEvent
		}
	case interfaces.Stop:
		if p.state.State == interfaces.Pause || p.state.State == interfaces.Play {
			p.stop()
			p.state.State = interfaces.Stop
			return true, interfaces.EventStop
		}
	case interfaces.Continue:
		if p.state.State == interfaces.Pause && p.audio.hasStreamer() {
			p.audio.pause(false)
			p.state.State = interfaces.Play
			return true, interfaces.EventUnpause
		}
		return false, defaultEvent
	case interfaces.EndSong:
		if p.state.State == interfaces.Pause || p.state.State == interfaces.Play {
			p.endStream()
			return true, interfaces.EventStop
		}
	default:
		logrus.Error("Got invalid action: ", action.State)
		return false, defaultEvent
	}
	return false, defaultEvent

}

func (p *Player) stop() {
	if p.state.State == interfaces.Play || p.state.State == interfaces.Pause {
		p.audio.stop()
		p.reportStatus(interfaces.EventStop)
	}
	p.state.State = interfaces.Stop
}

//RefreshState pushes current state into state channel
func (p *Player) RefreshState() {
	p.chanState <- p.state
}

func (p *Player) PlaySong(action Action) bool {
	reader, err := p.Api.GetSongDirect(action.AudioId, "mp3")
	if err != nil {
		logrus.Error("failed to request file over http: ", err.Error())
		return false
	} else {
		err = p.audio.newFileStream(reader, FormatMp3)
		if err != nil {
			logrus.Error("Failed to create new stream: ", err.Error())
			return false
		}
	}
	if p.reader != nil {
		p.reader.Close()
	}
	p.itemId = action.AudioId
	p.reader = reader
	p.state.State = interfaces.Play
	p.state.Song = action.Song
	p.state.Artist = action.Artist
	p.state.Album = action.Album
	p.state.CurrentSongDuration = action.Duration

	p.reportStatus(interfaces.EventStart)
	return true
}

func (p *Player) playSongFromReader(play PlaySong) {
	err := p.audio.newFileStream(play.Song, FormatMp3)
	if err != nil {
		logrus.Error("Failed to create new stream: ", err.Error())
		return
	}
	if p.reader != nil {
		p.reader.Close()
	}
	action := play.Action

	p.itemId = action.AudioId
	p.reader = play.Song
	p.state.State = interfaces.Play
	p.state.Song = action.Song
	p.state.Artist = action.Artist
	p.state.Album = action.Album
	p.state.CurrentSongDuration = action.Duration
}

func (p *Player) reportStatus(event interfaces.ApiPlaybackEvent) {
	state := &interfaces.ApiPlaybackState{
		Event:          event,
		ItemId:         p.itemId,
		IsPaused:       false,
		IsMuted:        false,
		PlaylistLength: p.state.CurrentSongDuration,
		Position:       p.state.CurrentSongPast,
		Volume:         p.state.Volume,
	}

	if p.state.State == interfaces.Pause {
		state.IsPaused = true
	}

	err := p.Api.ReportProgress(state)
	if err != nil {
		logrus.Error("Failed to report status: ", err.Error())
	}
}

func (p *Player) endStream() {
	logrus.Debug("Stream complete")
	if p.reader != nil {
		err := p.reader.Close()
		if err != nil {
			logrus.Errorf("Failed to close reader: %v", err)
		}
		p.reader = nil
	}
	p.stop()
	p.state.State = interfaces.SongComplete
	p.state.Clear()
	p.RefreshState()
	p.state.State = interfaces.Stop
}
