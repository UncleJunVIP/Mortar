#!/bin/zsh
rm ./mortar.log || true

adb pull "/mnt/SDCARD/Tools/tg5040/Mortar.pak/mortar.log" ./mortar.log

jq . ./.device_logs/mortar.log >> ./.device_logs/mortar.json

printf "All done!"

printf "\a"
