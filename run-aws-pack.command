#!/bin/bash

rm main
rm archive.zip
GOOS=linux GOARCH=amd64 go build -o main main.go
zip -r archive.zip ./client/dist Procfile main
