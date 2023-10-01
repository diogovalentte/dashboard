#!/bin/bash

apt update
apt install -y bzip2 curl

# Install golang
wget -O /tmp/go1.21.1.linux-amd64.tar.gz https://go.dev/dl/go1.21.1.linux-amd64.tar.gz
tar -C /tmp/ -xzf /tmp/go1.21.1.linux-amd64.tar.gz
export PATH=$PATH:/tmp/go/bin

# Install firefox v117.0.1
wget -O /tmp/firefox-117.0.1.tar.bz2 "https://download.mozilla.org/?product=firefox-117.0.1&os=linux64&lang=en-US"
sudo tar -C /opt -xvf /tmp/firefox-117.0.1.tar.bz2

# Start API
go build -o dashboard-api main.go
./dashboard-api
