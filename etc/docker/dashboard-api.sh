#!/bin/bash

# Start API
export PATH=$PATH:/tmp/go/bin
go build -o dashboard-api main.go
./dashboard-api
