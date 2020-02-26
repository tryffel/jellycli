# Jellycli

Terminal client for Jellyfin, mostly for music at the moment. This is very much work-in-progress.

## Features
* Play songs
* Add songs to queue
* Control (and view) play state through Dbus integration
* Control from other clients through websocket. Currently implemented:
    * Play / pause / stop
    * Set volume
    * Next track

## Building:
Assuming go installed:
```
go get -u tryffel.net/go/jellycli/cmd
# go to /cmd
go run .
```

On first time application asks for Jellyfin host, username, password and default collection for music. 
It stores all this information in OS wallet (tested only with KDE KWallet). After this, you should be able to 
browse your music and play it. 



## Acknowledgements
Thanks [natsukagami](https://github.com/natsukagami/mpd-mpris) for implementing Mpris-interface.
