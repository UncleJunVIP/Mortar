#!/bin/sh

ctrl_c() {
  printf "\n\nCtrl+C detected! Killing Mortar\n"
  killall mortar || exit 0
  exit 0
}

cd /mnt/SDCARD/Tools/tg5040/Mortar.pak || exit

touch mortar.log

rm mortar.log

touch mortar.log

trap ctrl_c INT

tail -f mortar.log
