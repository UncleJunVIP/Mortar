#!/bin/zsh
env GOOS=linux GOARCH=arm64 go build -o mortar .

adb push ./game-manager "/mnt/SDCARD/Tools/tg5040/Mortar.pak"

echo '\a'

echo "Donzo Washington!"
