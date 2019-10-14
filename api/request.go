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
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	errInvalidRequest       = "invalid request"
	errUnexpectedStatusCode = "unexpected statuscode"
	errServerError          = "server error"
)

func (a *Api) defaultParams() *map[string]string {
	params := *(&map[string]string{})
	params["UserId"] = a.userId
	params["DeviceId"] = a.DeviceId
	return &params
}

func (a *Api) get(url string, params *map[string]string) (io.ReadCloser, error) {
	return a.makeRequest("GET", url, nil, params)
}

func (a *Api) post(url string, body []byte, params *map[string]string) (io.ReadCloser, error) {
	return a.makeRequest("POST", url, nil, params)
}

//Construct request
// Set authorization header and build url query
// Make request, parse response code and raise error if needed. Else return response body
func (a *Api) makeRequest(method, url string, body *[]byte, params *map[string]string) (io.ReadCloser, error) {
	var reader *bytes.Buffer
	var req *http.Request
	var err error
	if body != nil {
		reader = bytes.NewBuffer(*body)
		req, err = http.NewRequest(method, a.host+url, reader)
	} else {
		req, err = http.NewRequest(method, a.host+url, nil)
	}

	if err != nil {
		return http.Response{}.Body, fmt.Errorf("failed to make request: %v", err)
	}
	req.Header.Set("X-Emby-Authorization", a.token)

	if params != nil {
		q := req.URL.Query()
		for i, v := range *params {
			q.Add(i, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed make request: %v", err)
	}

	if resp.StatusCode == 200 {
		return resp.Body, nil
	}
	if resp.StatusCode > 400 && resp.StatusCode < 500 {
		return resp.Body, fmt.Errorf("%s, code: %d", errInvalidRequest, resp.StatusCode)
	}
	if resp.StatusCode >= 500 {
		return resp.Body, errors.New(errServerError)
	} else {
		return resp.Body, errors.New(errUnexpectedStatusCode)
	}
}
