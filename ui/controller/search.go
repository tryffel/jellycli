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

package controller

import (
	"github.com/sirupsen/logrus"
	"sync"
	"tryffel.net/pkg/jellycli/api"
	"tryffel.net/pkg/jellycli/task"
)

type Action int

const (
	ActionSearch Action = 0
)

type Content struct {
	task.Task
	api  *api.Api
	lock sync.RWMutex

	searchResults *api.SearchResult
	chanComplete  chan Action
}

func NewContent(a *api.Api) *Content {
	c := &Content{
		api: a,
	}
	c.SetLoop(c.loop)
	c.chanComplete = make(chan Action)
	return c
}

// Search performs search query
func (c *Content) Search(q string) {
	results, err := c.api.Search(q, 20)
	if err != nil {
		logrus.Error("Search failed: ", err.Error())
	} else {
		if results != nil {
			c.searchResults = results
		}
	}
	c.chanComplete <- ActionSearch
	logrus.Debug("Content search copmlete")

}

//SearchResults returns latest search results from index to index.
func (c *Content) SearchResults() *api.SearchResult {
	return c.searchResults
}

func (c *Content) SearchCompleteChan() chan Action {
	return c.chanComplete
}

func (c *Content) loop() {
	for true {
		select {
		case <-c.StopChan():
			break
		}
	}

}
