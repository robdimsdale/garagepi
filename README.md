#Garagepi

[![Build Status](https://travis-ci.org/robdimsdale/garagepi.svg?branch=master)](https://travis-ci.org/robdimsdale/garagepi) [![Coverage Status](https://img.shields.io/coveralls/robdimsdale/garagepi.svg)](https://coveralls.io/r/robdimsdale/garagepi?branch=master)

A webserver written in golang to display output of Raspberry Pi camera module and trigger gpio.

A typical use would be to view the interior of a garage and trigger the garage door opener via gpio (and a relay).

Copyright (c) 2014, Robert Dimsdale. Licensed under [MIT License].

##Getting started
Install Go, [WiringPi] and jacksonliam's [experimental mjpg-streamer].

###Go dependencies
```
go get github.com/GeertJohan/go.rice
go get github.com/gorilla/mux
```

###Installing
```
go install main.go
```
If this results in an error `go install: no install location for .go files listed on command line (GOBIN not set)` then an alternative is:
```
go build -o $GOPATH/bin/garagepi main.go
```

###Init scripts
Copy the init scripts to `/etc/init.d/` and set them to run automatically on boot with the following commands:

```
sudo cp init-scripts/* /etc/init.d/
sudo update-rc.d garagepi defaults
sudo update-rc.d garagerelay defaults
sudo update-rc.d garagestreamer defaults
```

The default location for the `garagepi` binary is `/go/bin/garagepi`. This is controlled by the `GARAGE_PI_BINARY` environment variable in `init-scripts/garagepi`.

###Logging

By default logs are sent to `/dev/null`. This is controlled by the `OUT_LOG` environment variable in `init-scripts/garagepi` and `init-scripts/garagestreamer`. These can either be set to the same file or different files.

##Performance

###Multiple Pis
Performance can be improved by using multiple Pis - one for the mjpg streamer (with the camera attached) and one for the Go webserver (with the gpio attached). The responsiveness of the Go webserver is significantly improved and the framerate of the streamer improved slightly. Stability appears much better (the webserver/streamer crash more frequently when co-located on the same Pi).

The gpio utility is lightweight and so it may be installed on both, but it is only required to be installed on the Pi directly attached to the relay. The streamer utility, however, requires much more resouce and therefore should only be installed on the Pi with the camera attached.

On the Pi with the camera, copy only the garage streamer start script:

```
sudo cp init-scripts/garagestreamer /etc/init.d/
sudo update-rc.d garagestreamer defaults
```

On the Pi with the Go webserver and gpio, copy only the garagepi and garagerelay start scripts:

```
sudo cp init-scripts/garagepi /etc/init.d/
sudo cp init-scripts/garagerelay /etc/init.d/
sudo update-rc.d garagepi defaults
sudo update-rc.d garagerelay defaults
```

By default, the `garagepi` webserver assumes the webcam is available on `localhost:8080`. This is controlled by the the environment variables `$WEBCAM_HOST` and `$WEBCAM_PORT` in `init-scripts/garagepi`.

[MIT License]: https://github.com/robdimsdale/garagepi/raw/master/LICENSE

[WiringPi]: https://github.com/WiringPi/WiringPi

[experimental mjpg-streamer]: https://github.com/jacksonliam/mjpg-streamer
