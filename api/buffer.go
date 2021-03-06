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

package api

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
)

func MimeToAudioFormat(mimeType string) (format interfaces.AudioFormat, err error) {
	format = interfaces.AudioFormatNil
	switch mimeType {
	case "audio/mpeg":
		format = interfaces.AudioFormatMp3
	case "audio/flac":
		format = interfaces.AudioFormatFlac
	case "audio/ogg":
		format = interfaces.AudioFormatOgg
	case "audio/wav":
		format = interfaces.AudioFormatWav

	default:
		err = fmt.Errorf("unidentified audio format: %s", mimeType)
	}
	return
}

// StreamBuffer is a buffer that reads whole http body in the background and copies it to local buffer.
type StreamBuffer struct {
	lock           *sync.Mutex
	url            string
	headers        map[string]string
	params         map[string]string
	client         *http.Client
	buff           *bytes.Buffer
	bitrate        int
	req            *http.Request
	resp           *http.Response
	cancelDownload chan bool
}

func (s *StreamBuffer) Read(p []byte) (n int, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	n, err = s.buff.Read(p)
	return
}

func (s *StreamBuffer) Close() error {
	logrus.Debug("Close stream download")
	return s.resp.Body.Close()
}

func (s *StreamBuffer) Len() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.buff.Len()
}

func (s *StreamBuffer) SecondsBuffered() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	buffered := s.buff.Len()
	return buffered / s.bitrate
}

func (s *StreamBuffer) AudioFormat() (format interfaces.AudioFormat, err error) {
	if s.resp != nil {
		return MimeToAudioFormat(s.resp.Header.Get("Content-Type"))
	}
	return interfaces.AudioFormatNil, errors.New("no http response")
}

func NewStreamDownload(url string, headers map[string]string, params map[string]string,
	client *http.Client, duration int) (*StreamBuffer, error) {
	stream := &StreamBuffer{
		lock:           &sync.Mutex{},
		url:            url,
		headers:        headers,
		params:         params,
		bitrate:        duration,
		buff:           bytes.NewBuffer(make([]byte, 0, 1024)),
		cancelDownload: make(chan bool),
	}
	if client == nil {
		client = http.DefaultClient
	}
	stream.client = client

	var err error
	stream.req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return stream, fmt.Errorf("init http request: %v", err)
	}

	for k, v := range headers {
		stream.req.Header.Add(k, v)
	}

	if params != nil {
		q := stream.req.URL.Query()
		for i, v := range params {
			q.Add(i, v)
		}
		stream.req.URL.RawQuery = q.Encode()
	}

	stream.resp, err = stream.client.Do(stream.req)
	if err != nil {
		return stream, fmt.Errorf("make http request: %v", err)

	}
	if stream.resp.StatusCode != 200 {
		return stream, fmt.Errorf("http request error, statuscode: %d", stream.resp.StatusCode)

	}

	sLength := stream.resp.Header.Get("Content-Length")
	length, err := strconv.Atoi(sLength)

	stream.bitrate = length / duration
	for {
		if stream.buff.Len() > stream.bitrate*config.AppConfig.Player.HttpBufferingS {
			break
		}
		failed := stream.readData()
		if failed {
			return stream, fmt.Errorf("initial buffer failed")
		}
	}
	go stream.bufferBackground()
	return stream, err
}

func (s *StreamBuffer) bufferBackground() {
	logrus.Debug("Start buffered stream")
	timer := time.NewTimer(time.Millisecond)
	defer timer.Stop()
loop:
	for {
		select {
		case <-timer.C:
			if s.buff.Len()/1024/1024 > config.AppConfig.Player.HttpBufferingLimitMem {
				logrus.Tracef("Buffer is full")
				timer.Reset(time.Second)
			} else {
				if !s.readData() {
					timer.Reset(time.Second)
				} else {
					break loop
				}
			}
		case <-s.cancelDownload:
			logrus.Debug("Stop buffered stream")
			break loop
		}
	}

	close(s.cancelDownload)
	s.cancelDownload = nil
}

func (s *StreamBuffer) readData() bool {
	var nHttp int
	var nBuff int
	var err error
	buf := make([]byte, s.bitrate*5)

	s.lock.Lock()
	defer s.lock.Unlock()
	nHttp, err = s.resp.Body.Read(buf)
	stop := false
	if err != nil {
		if err == io.EOF {
			if nHttp == 0 {
				logrus.Debugf("buffer download complete")
				stop = true
			} else {
				// pass
			}
		} else {
			logrus.Errorf("buffer read bytes from body: %v", err)
			stop = true
		}
	}

	buf = buf[0:nHttp]
	if nHttp > 0 {
		nBuff, err = s.buff.Write(buf)
		if err != nil {
			if err == io.EOF {
			} else {
				logrus.Warningf("Copy buffer: %v", err)
			}
		}
		if nBuff != nHttp {
			logrus.Warningf("incomplete buffer read: have %d B, want %d B", nBuff, nHttp)
		}
	}
	size := s.buff.Len()
	if size > 0 && s.bitrate > 0 {
		logrus.Tracef("Buffer: %d KiB, %d sec, bitrate %d bit/s", size/1024, size/s.bitrate, s.bitrate)
	} else {
		logrus.Tracef("Empty buffer, size %d B, bitrate %d", size, s.bitrate)

	}
	return stop
}
