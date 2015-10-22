#!/bin/sh

DIR="/home/xujian/study/go/download/"
export GOPATH=$DIR
go build -o Client client.go 
go build -o Server server.go
go build -o Lookup lookup_server.go
