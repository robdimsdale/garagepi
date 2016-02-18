FROM golang:1.6

RUN chmod -R 777 /usr/local/go

RUN apt-get update && \
  apt-get install -y \
    build-essential \
    g++ \
    flex \
    bison \
    gperf \
    ruby \
    perl \
    libsqlite3-dev \
    libfontconfig1-dev \
    libicu-dev \
    libfreetype6 \
    libssl-dev \
    libpng-dev \
    libjpeg-dev \
    python \
    libx11-dev \
    libxext-dev && \
  apt-get autoremove -y && \
  apt-get clean all

RUN curl \
    -L \
    -o phantomjs.tar.bz2 \
    https://bitbucket.org/ariya/phantomjs/downloads/phantomjs-2.1.1-linux-x86_64.tar.bz2 && \
  tar -xvf phantomjs.tar.bz2 && \
  mv phantomjs-*/bin/phantomjs /usr/bin && \
  rm -rf phantomjs*
