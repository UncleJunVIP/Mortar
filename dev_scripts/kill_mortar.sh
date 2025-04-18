#!/bin/zsh
printf "Shit broke! Killing Mortar."
sshpass -p 'tina' ssh root@192.168.1.16 "kill  \$(pidof dlv)" > /dev/null 2>&1
sshpass -p 'tina' ssh root@192.168.1.16 "kill  \$(pidof mortar)" > /dev/null 2>&1
