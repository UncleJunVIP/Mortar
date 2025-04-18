#!/bin/zsh
printf "Starting Mortar Remote Debugger"
while true; do
  sshpass -p 'tina' ssh root@192.168.1.16 "sh -c '/mnt/SDCARD/Developer/bin/dlv attach --headless --listen=:2345 --api-version=2 --accept-multiclient \$(pidof mortar)'" > /dev/null 2>&1
  sleep 3
done
