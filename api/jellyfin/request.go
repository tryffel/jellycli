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

package jellyfin

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	errInvalidRequest       = "invalid request"
	errUnexpectedStatusCode = "unexpected statuscode"
	errServerError          = "server error"
	errNotFound             = "page not found"
	errUnauthorized         = "needs authorization"
	errForbidden            = "forbidden"
)

func (a *Jellyfin) defaultParams() *params {
	params := *(&params{})
	params["UserId"] = a.userId
	params["DeviceId"] = a.DeviceId
	return &params
}

func (a *Jellyfin) get(url string, params *params) (io.ReadCloser, error) {
	resp, err := a.makeRequest("GET", url, nil, params, nil)
	if resp != nil {
		return resp.Body, err
	}
	return nil, err
}

func (a *Jellyfin) post(url string, body *[]byte, params *params) (io.ReadCloser, error) {
	resp, err := a.makeRequest("POST", url, body, params, nil)
	if resp != nil {
		return resp.Body, err
	}
	return nil, err
}

//Construct request
// Set authorization header and build url query
// Make request, parse response code and raise error if needed. Else return response body
func (a *Jellyfin) makeRequest(method, url string, body *[]byte, params *params,
	headers map[string]string) (*http.Response, error) {
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
		return &http.Response{}, fmt.Errorf("failed to make request: %v", err)
	}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Emby-Token", a.token)

	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	if params != nil {
		q := req.URL.Query()
		for i, v := range *params {
			q.Add(i, v)
		}
		req.URL.RawQuery = q.Encode()
	}
	start := time.Now()
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed make request: %v", err)
	}
	took := time.Since(start)
	logrus.Debugf("%s %s: %d (%d ms)", req.Method, req.URL.Path, resp.StatusCode, took.Milliseconds())

	if resp.StatusCode == 200 || resp.StatusCode == 204 {
		return resp, nil
	}
	bytes, _ := ioutil.ReadAll(resp.Body)
	var msg string = "no body"
	if len(bytes) > 0 {
		msg = string(bytes)
	}
	var errMsg string
	switch resp.StatusCode {
	case 400:
		errMsg = errInvalidRequest
	case 401:
		errMsg = errUnauthorized
	case 403:
		errMsg = errForbidden
	case 404:
		errMsg = errNotFound
	case 500:
		errMsg = errServerError
	default:
		errMsg = errUnexpectedStatusCode
	}
	return resp, fmt.Errorf("%s, code: %d, msg: %s", errMsg, resp.StatusCode, msg)
}
