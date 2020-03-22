# Jellycli

[![Godoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/tryffel.net/go/jellycli)
[![Go Report Card](https://goreportcard.com/badge/tryffel.net/go/jellycli)](https://goreportcard.com/report/tryffel.net/go/jellycli)

Terminal client for Jellyfin, mostly for music at the moment.

![Screenshot](screenshot.png)

## Features
* Play artists, songs, albums, playlists, favorite artists, favorite albums
* Add songs to queue, clear queue
* Control (and view) play state through Dbus integration
* Remote control over Jellyfin server. Currently implemented:
    * Play / pause / stop
    * Set volume
    * Next/previous track
* Supported formats (server transcodes rest to mp3): mp3,ogg,flac,wav

## Building
**You will need Go 1.13 or Go 1.14 installed and configured**

* For additional audio libraries required, see [Hajimehoshi/oto](https://github.com/hajimehoshi/oto). 
On linux you need libasound2-dev.
* Currently jellycli has issues with Windows and is unable to start properly.

**Warning: for the time being, use git clone directly instead of go get.** There is an issue with dependency 
(rivo/tview) being automatically upgraded and causing deadlocks.

Download & build package
```
git clone https://github.com/tryffel/jellycli.git
cd jellycli
go build .
./jellycli
```

On first time application asks for Jellyfin host, username, password and default collection for music. 
All this is stored in configuration file at ~/.config/jellycli/jellycli.yaml. 
Configuration file location is visible in help page. 
You can use multiple config files by providing argument:
```
jellycli --config temp.yaml
```

Log file is located at '/tmp/jellycli.log' by default. This can be overridden with config file. 
At the moment jellycli does not inform user about errors but rather just silently logs them.
For development purposes you should set log-level either to debug or trace.

## Acknowledgements
Thanks [natsukagami](https://github.com/natsukagami/mpd-mpris) for implementing Mpris-interface.
