#!/bin/bash

apt update
apt install bzip2

curl -L -o /tmp/firefox-117.0.1.tar.bz2 "https://download.mozilla.org/?product=firefox-117.0.1&os=linux64&lang=en-US"
tar -xvf /tmp/firefox-117.0.1.tar.bz2 -C ./etc/

curl -L -o /tmp/geckodriver.tar.gz https://github.com/mozilla/geckodriver/releases/download/v0.33.0/geckodriver-v0.33.0-linux64.tar.gz
mkdir -p /opt/geckodriver/
mkdir -p etc/geckodriver
tar -xvf /tmp/geckodriver.tar.gz -C ./etc/geckodriver/

go build -o dashboard-api main.go
./dashboard-api
