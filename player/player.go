package player

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"time"
	"tryffel.net/pkg/jellycli/api"
	"tryffel.net/pkg/jellycli/config"
	"tryffel.net/pkg/jellycli/task"
)

type State int
type Playtype int
type Status int

//Action includes instruction for player to play
type Action struct {
	// What to do
	State State
	Type  Playtype

	// Provide either artist/album/song or audio id
	Artist  string
	Album   string
	Song    string
	AudioId string
}

//PlayerState holds data about currently playing song if any
type PlayingState struct {
	State       State
	PlayingType Playtype
	Song        string
	Artist      string
	Album       string

	CurrentSongDuration int
	CurrentSongPast     int
	PlaylistDuration    int
	PlaylistLeft        int
}

const (
	// Player states
	Play  State = 1
	Pause State = 2
	Stop  State = 0

	// Playing single song
	Song Playtype = 0
	// Playing album
	Album Playtype = 1
	// Playing artists discography
	Artist Playtype = 2
	// Playing playlist
	Playlist Playtype = 3

	// Last action was ok
	StatusOk Status = 0
	// Last action resulted in error
	StatusError Status = 0

	// How often to update state i.e. push status to playingstate channel
	updateInterval = time.Second
)

// Player is the application structure
type Player struct {
	task.Task

	Api *api.Api
	// chanAction is for user interactions
	chanAction chan Action
	// chanState is updated when state is changed
	chanState chan PlayingState

	chanStreamComplete chan bool

	ticker *time.Ticker

	state      PlayingState
	lastAction *Action

	audio *audio

	file *os.File
}

// NewPlayer constructs new player instance
func NewPlayer(a *api.Api) (*Player, error) {
	p := &Player{
		Api:                a,
		chanAction:         make(chan Action),
		chanState:          make(chan PlayingState),
		chanStreamComplete: make(chan bool),
		ticker:             nil,
		audio:              nil,
	}
	p.audio = newAudio(p.chanStreamComplete)
	err := initAudio()
	if err != nil {
		return p, fmt.Errorf("audio init failed: %v", err)
	}
	p.SetLoop(p.loop)

	file, err := os.Open("my-music-file.mp3")
	if err != nil {
		logrus.Error("failed to open audio file")
	}
	err = p.audio.newStream(file, FormatMp3)
	if err != nil {
		logrus.Error("failed to add stream: %v", err)
	}
	p.playMedia()
	p.state.State = Pause

	p.file = file

	return p, nil
}

//ActionChannel returns input channel for user actions
func (p *Player) ActionChannel() chan Action {
	return p.chanAction
}

//StateChannel return output channel for player state
func (p *Player) StateChannel() chan PlayingState {
	return p.chanState
}

func (p *Player) loop() {
	p.ticker = time.NewTicker(updateInterval)
	defer p.ticker.Stop()
	chunkStarted := time.Now()
	for true {
		select {
		case tick := <-p.ticker.C:
			if (tick.Second() - chunkStarted.Second()) >= 10 {
				// Query new chunk every 10 sec
				chunkStarted = time.Now()
			}
			at := p.audio.timePast()
			p.state.CurrentSongPast = int(at.Seconds())

			err := p.audio.streamer.Err()
			if err != nil {
				logrus.Error("error in streamer: ", err.Error())
			}
			p.RefreshState()
		case action := <-p.chanAction:
			logrus.Debug("Player received action")
			// User has requested action
			p.lastAction = &action
			if p.state.State != action.State {
				if action.State == Play || action.State == Pause {
					_ = p.audio.pause()
					if p.audio.ctrl.Paused {
						p.state.State = Pause
					} else {
						if action.State == Play {
							p.state.State = Play
						}
					}
				}
			}
			p.RefreshState()

		case <-p.chanStreamComplete:
			logrus.Debug("Stream complete")
			p.audio.streamer.Close()
			p.state.State = Stop
			p.RefreshState()
		case <-p.StopChan():
			// Program is stopping
			break
		}
	}
	p.file.Close()
}

//RefreshState pushes current state into state channel
func (p *Player) RefreshState() {
	logrus.Debug("emitting player state")
	p.chanState <- p.state
}

func (p *Player) playMedia() {
	length := p.audio.streamer.Len() / config.AudioSamplingRate
	logrus.Infof("Song length is %d sec.", length)

	err := p.audio.playStream()
	if err != nil {
		logrus.Error("Failed to play media: ", err.Error())
	}
	p.state.State = Play
	p.state.CurrentSongDuration = length
	p.state.CurrentSongPast = 0

}
