#!/bin/bash
if [ $# -ne 1 ]; then
    echo "USAGE: deploy.sh <PATH.TO.SERVER>"
	exit
fi

scp -r ./S2017-UPE-AI $1:~/npc/
scp -r ./templates $1:~/npc/
scp -r ./static $1:~/npc/

scp ./dbinterface/CREDENTIALS $1:~/npc/dbinterface/
scp ./dbinterface/words/*.txt $1:~/npc/dbinterface/words/


