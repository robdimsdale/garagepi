FROM robdimsdale/garagepi-1.5

RUN wget -qO- https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar -xvz -C /root

RUN cd /usr/local/go/src && \
  GOROOT_BOOTSTRAP=/root/go GOARCH=arm ./make.bash
