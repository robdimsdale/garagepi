FROM golang:1.5

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

RUN git clone git://github.com/ariya/phantomjs.git && \
  cd phantomjs && \
  git checkout 2.0 && \
  ./build.sh --confirm && \
  cd ../ && \
  mv phantomjs/bin/phantomjs /usr/bin && \
  rm -rf phantomjs/
