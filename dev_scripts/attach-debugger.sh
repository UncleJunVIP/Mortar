#!/bin/zsh
sshpass -p 'tina' ssh root@192.168.1.16 "sh -c '/mnt/SDCARD/Developer/bin/dlv attach --headless --listen=:2345 --api-version=2 --accept-multiclient \$(pidof mortar)'"
