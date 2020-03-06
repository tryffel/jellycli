# Jellycli

Terminal client for Jellyfin, mostly for music at the moment.

![Screenshot](screenshot.png)

## Features
* Play songs
* Play albums
* Add songs to queue
* Control (and view) play state through Dbus integration
* Control from other clients through websocket. Currently implemented:
    * Play / pause / stop
    * Set volume
    * Next track

## Building:
**You will need Go 1.13 installed and configured**

Download & build package
```
go get -u tryffel.net/go/jellycli
```
Jellycli-binary should now reside in $GOPATH/bin/jellycli. You can build it manually too:
```
go build tryffel.net/go/jellycli
```
Run
```
$GOPATH/bin/jellycli 
```

On first time application asks for Jellyfin host, username, password and default collection for music. 
All this is stored in configuration file at ~/.config/jellycli/jellycli.yaml.
You can use multiple config files by providing argument:
```
jellycli --config temp.yaml
```

## Acknowledgements
Thanks [natsukagami](https://github.com/natsukagami/mpd-mpris) for implementing Mpris-interface.
