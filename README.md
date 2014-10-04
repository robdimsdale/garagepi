#Garage-pi

[![Build Status](https://travis-ci.org/robdimsdale/garage-pi.svg?branch=master)](https://travis-ci.org/robdimsdale/garage-pi)

A webserver written in golang to display output of Raspberry Pi camera module and trigger gpio.

A typical use would be to view the interior of a garage and trigger the garage door opener via gpio (and a relay).

Copyright (c) 2014, Robert Dimsdale. Licensed under [MIT License].

##Getting started
Install Go, [WiringPi] and jacksonliam's [experimental mjpg-streamer].

###Go dependencies
The Go webserver assumes the GOPATH is set to /go.
```
go get github.com/GeertJohan/go.rice
go get github.com/gorilla/mux
```

###Init scripts
Copy the init scripts to /etc/init.d/ and set them to run automatically on boot:

```
sudo cp init-scripts/* /etc/init.d/
sudo update-rc.d garage-pi defaults
sudo update-rc.d garagerelay defaults
sudo update-rc.d garagestreamer defaults
```

###Logging

By default logs are sent to `/dev/null`. To log to a specific file edit `init-scripts/garage-pi` and change the following line:

```
OUT_LOG=/dev/null
```
to a file of your choice e.g.:
```
OUT_LOG=/home/pi/garage-pi.log
```

Logging for the relay and streamer can be achieved by modifying the `init-scripts/garagestreamer` and `init-scripts/garagerelay` to pipe the outputs of the respective processess to files.

##Performance

###Multiple Pis
Performance can be improved by using multiple Pis - one for the mjpg streamer (with the camera attached) and one for the Go webserver (with the gpio attached). The responsiveness of the Go webserver is significantly improved and the framerate of the streamer improved slightly. Stability appears much better (the webserver/streamer crash more frequently when colocated on the same Pi).

The gpio utility is lightweight and so it may be installed on both, but it is only required to be installed on the Pi directly attached to the relay. The streamer utility, however, requires much more resouce and therefore should only be installed on the Pi with the camera attached.

On the Pi with the camera, copy only the garagestreamer start script:

```
sudo cp init-scripts/garagestreamer /etc/init.d/
sudo update-rc.d garagestreamer defaults
```

On the Pi with the Go webserver and gpio, copy only the garage-pi and garagerelay start scripts:

```
sudo cp init-scripts/garage-pi /etc/init.d/
sudo cp init-scripts/garagerelay /etc/init.d/
sudo update-rc.d garage-pi defaults
sudo update-rc.d garagerelay defaults
```

Configuring an external router to port-forward requests to these two Pis will work with the code as-is; without this configuration the javascript in the Go webserver will need to know the hostname/ip of the mjpg streamer. If this is the case, change the following line in templates/homepage.html:

```
img.src = "https://" + window.location.hostname + ":9998" + "/?action=snapshot&n=" + (++imageNr);
```

to:

```
img.src = "https://hostname_of_pi_with_streamer_running:9998" + "/?action=snapshot&n=" + (++imageNr);
```

[MIT License]: https://github.com/robdimsdale/garage-pi/raw/master/LICENSE

[WiringPi]: https://github.com/WiringPi/WiringPi

[experimental mjpg-streamer]: https://github.com/jacksonliam/mjpg-streamer

