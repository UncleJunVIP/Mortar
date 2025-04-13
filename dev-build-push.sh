#!/bin/zsh
env GOOS=linux GOARCH=arm64 go build -o mortar .

adb push ./mortar "/mnt/SDCARD/Tools/tg5040/Mortar.pak"

echo '\a'

echo "Donzo! Mortar has been pushed to device!"
