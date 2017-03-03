#!/bin/bash
if [ $# -ne 1 ]; then
    echo "USAGE: deploy.sh <PATH.TO.SERVER>"
	exit
fi

scp -r ./templates $1:~/npc/
scp -r ./static $1:~/npc/

