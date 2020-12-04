/*
 * Jellycli is a terminal music player for Jellyfin.
 * Copyright (C) 2020 Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
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

func (jf *Jellyfin) defaultParams() *params {
	params := *(&params{})
	params["UserId"] = jf.userId
	params["DeviceId"] = jf.DeviceId
	return &params
}

func (jf *Jellyfin) get(url string, params *params) (io.ReadCloser, error) {
	resp, err := jf.makeRequest("GET", url, nil, params, nil)
	if resp != nil {
		return resp.Body, err
	}
	return nil, err
}

func (jf *Jellyfin) post(url string, body *[]byte, params *params) (io.ReadCloser, error) {
	resp, err := jf.makeRequest("POST", url, body, params, nil)
	if resp != nil {
		return resp.Body, err
	}
	return nil, err
}

//Construct request
// Set authorization header and build url query
// Make request, parse response code and raise error if needed. Else return response body
func (jf *Jellyfin) makeRequest(method, url string, body *[]byte, params *params,
	headers map[string]string) (*http.Response, error) {
	var reader *bytes.Buffer
	var req *http.Request
	var err error
	if body != nil {
		reader = bytes.NewBuffer(*body)
		req, err = http.NewRequest(method, jf.host+url, reader)
	} else {
		req, err = http.NewRequest(method, jf.host+url, nil)
	}

	if err != nil {
		return &http.Response{}, fmt.Errorf("failed to make request: %v", err)
	}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Emby-Token", jf.token)

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
	resp, err := jf.client.Do(req)
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
