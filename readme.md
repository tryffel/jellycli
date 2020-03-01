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

Download package
```
go get -u tryffel.net/go/jellycli
```
If you get output 'no go files in ...', run:
```
go get -u tryffel.net/go/jellycli
go get -u tryffel.net/go/jellycli/cmd
```
Build & run
```
go build -o jellycli tryffel.net/go/jellycli/cmd
./jellycli
```

On first time application asks for Jellyfin host, username, password and default collection for music. 
It stores all this information in OS wallet (tested only with KDE KWallet). After this, you should be able to 
browse your music and play it. 



## Acknowledgements
Thanks [natsukagami](https://github.com/natsukagami/mpd-mpris) for implementing Mpris-interface.
