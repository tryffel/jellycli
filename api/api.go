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

package api

import (
	"errors"
	"net/http"
)

type Api struct {
	host     string
	token    string
	userId   string
	serverId string
	client   *http.Client
	loggedIn bool
}

func NewApi(host string) *Api {
	return &Api{
		host:   host,
		token:  "",
		client: &http.Client{},
	}
}

func (a *Api) Host() string {
	return a.host
}

func (a *Api) Token() string {
	return a.token
}

//Login performs username based login
func (a *Api) Login(username, password string) error {
	return a.login(username, password)
}

//SetToken sets existing token
func (a *Api) SetToken(token string) error {
	a.token = token
	return nil
}

func (a *Api) tokenExists() error {
	if a.token == "" {
		return errors.New("not logged in")
	}
	return nil
}

func (a *Api) SetUserId(id string) {
	a.userId = id
}

func (a *Api) UserId() string {
	return a.userId
}

func (a *Api) IsLoggedIn() bool {
	return a.loggedIn
}
