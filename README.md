# Garagepi

[![Build Status](https://travis-ci.org/robdimsdale/garagepi.svg?branch=master)](https://travis-ci.org/robdimsdale/garagepi) [![Coverage Status](https://img.shields.io/coveralls/robdimsdale/garagepi.svg)](https://coveralls.io/r/robdimsdale/garagepi?branch=master)

A webserver written in golang to display output of Raspberry Pi camera module and trigger gpio.

A typical use would be to view the interior of a garage and trigger the garage door opener via gpio (and a relay).

Copyright © 2014-2015, Robert Dimsdale. Licensed under the [MIT License](https://github.com/robdimsdale/garagepi/raw/master/LICENSE).

## Getting started

Requires jacksonliam's [experimental mjpg-streamer](https://github.com/jacksonliam/mjpg-streamer).

### Downloading

Obtain the most recent binary from the [releases page](https://github.com/robdimsdale/garagepi/releases).

### Init scripts

Clone this repo, and from within the cloned directory copy the init scripts to `/etc/init.d/` as follows:

```
sudo cp scripts/init-scripts/* /etc/init.d/
```

Set them to run automatically on boot:

```
sudo update-rc.d garagepi defaults
sudo update-rc.d garagestreamer defaults
```

The default location for the `garagepi` binary is `/go/bin/garagepi`. This is controlled by the `GARAGE_PI_BINARY` environment variable in `scripts/init-scripts/garagepi`. Edit the script and set this variable to the location of the downloaded binary.

### Logging

By default logs are sent to the syslog with the tag `garagepi` as well as to the file `/dev/null`. The location of the additional file is controlled by the `OUT_LOG` environment variable in `scripts/init-scripts/garagepi` and `scripts/init-scripts/garagestreamer`. These can either be set to the same file or different files.

### SSL

SSL requires a private key and a public certificate. The certificate should be the concatenated cert chain, e.g.:

```
-----BEGIN CERTIFICATE-----
(Your Primary SSL certificate: your_domain_name.crt)
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
(Your Intermediate certificate: DigiCertCA.crt)
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
(Your Root certificate: TrustedRoot.crt)
-----END CERTIFICATE-----
```

The trusted root CA is generally not required.

## Performance

### TLS

The Raspberry Pi supports TLS termination, but it is relatively slow. This causes a significant decrease in the framerate of the webcam (API and web calls are essentially unaffected).

If the TLS termination can happen upstream of the Pi, forwarding requests on in plain text to the Pi, the perfomance will be significantly improved.

### Multiple clients

The mjpg-streamer supports relatively fast streaming to a single client, but multiple clients significantly decrease the framerate of the webcam.

### Multiple Pis

If using Raspberry Pi 2, there is no performance gained by using multiple Raspberry Pis.

If using Raspberry Pi 1, performance can be improved by using multiple Pis - one for the mjpg streamer (with the camera attached) and one for the Go webserver (with the gpio attached). The responsiveness of the Go webserver is significantly improved and the framerate of the streamer improved slightly. Stability appears much better (the webserver/streamer crash more frequently when co-located on the same Pi).

The gpio utility is lightweight and so it may be installed on both, but it is only required to be installed on the Pi directly attached to the relay. The streamer utility, however, requires much more resouce and therefore should only be installed on the Pi with the camera attached.

On the Pi with the camera, copy only the garage streamer start script:

```
sudo cp scripts/init-scripts/garagestreamer /etc/init.d/
sudo update-rc.d garagestreamer defaults
```

On the Pi with the Go webserver and gpio, copy only the garagepi start script:

```
sudo cp scripts/init-scripts/garagepi /etc/init.d/
sudo update-rc.d garagepi defaults
```

By default, the `garagepi` webserver assumes the webcam is available on `localhost:8080`. This is controlled by the the environment variables `$WEBCAM_HOST` and `$WEBCAM_PORT` in `scripts/init-scripts/garagepi`.

## Development

Requires Golang 1.4 or higher.

### Go dependencies

Dependencies are managed using [godep](https://github.com/tools/godep). Install it as follows:

```
go get -u github.com/tools/godep
```

From within the directory of this cloned repo, fetch the golang dependencies:

```
godep restore
```

To regenerate the embedded assets, install the [esc](https://github.com/mjibson/esc) tool:

```
go get github.com/mjibson/esc
```

and then run the script which creates them:

```
./scripts/create-embedded-assets
```

### Running the tests

The tests require the [ginkgo](https://github.com/onsi/ginkgo/) binary and [phantomJS](https://github.com/ariya/phantomjs/).

Install `ginkgo` via `go get`:

```
go get github.com/onsi/github.com/ginkgo/ginkgo
```

Install `phantomJS` e.g. for OSX:

```
brew install phantomjs
```

Execute the unit and integration tests with:

```
./scripts/unit-test
./scripts/integration-tests
```

## Project administration

- Roadmap: [Pivotal Tracker](https://www.pivotaltracker.com/n/projects/1401690)
