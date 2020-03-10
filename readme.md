# Jellycli

Terminal client for Jellyfin, mostly for music at the moment.

![Screenshot](screenshot.png)

## Features
* Play artists, songs, albums, playlists, favorite artists
* Add songs to queue
* Control (and view) play state through Dbus integration
* Control from other clients through websocket. Currently implemented:
    * Play / pause / stop
    * Set volume
    * Next track

## Building:
**You will need Go 1.13 or Go 1.14 installed and configured**

* On linux you need to have alsalib-dev installed.
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
You can use multiple config files by providing argument:
```
jellycli --config temp.yaml
```

## Acknowledgements
Thanks [natsukagami](https://github.com/natsukagami/mpd-mpris) for implementing Mpris-interface.
