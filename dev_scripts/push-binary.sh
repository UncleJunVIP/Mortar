#!/bin/zsh
adb push ./mortar "/mnt/SDCARD/Tools/tg5040/Mortar.pak"
adb push ./config.yml "/mnt/SDCARD/Tools/tg5040/Mortar.pak"

adb shell rm "/mnt/SDCARD/Tools/tg5040/Mortar.pak/mortar.log" || true

printf "Mortar has been pushed to device!"

printf "\a"
