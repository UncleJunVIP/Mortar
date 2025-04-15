#!/bin/zsh
env GOOS=linux GOARCH=arm64 go build -gcflags="all=-N -l" -o mortar . || exit
