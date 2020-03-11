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

import "tryffel.net/go/jellycli/api"

// Player
type Player struct {
	*Audio
	*Queue
	*Items

	songComplete chan bool

	api *api.Api
}

// initialize new player. This also initializes faiface.Speaker, which should be initialized only once.
func newPlayer(api *api.Api) (*Player, error) {
	var err error
	p := &Player{
		songComplete: make(chan bool, 3),
	}

	p.Audio, err = newAudio()
	p.Queue = newQueue()
	if err != nil {
		return p, err
	}

	return p, nil
}

func (p *Player) songCompleted() {
	p.songComplete <- true
}
