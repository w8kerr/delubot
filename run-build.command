#!/bin/bash

cd ~/github/delubot
rm main
GOOS=linux GOARCH=amd64 go build -o main main.go
