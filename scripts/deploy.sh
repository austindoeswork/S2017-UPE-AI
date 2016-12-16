#!/bin/sh
commithash=$(git rev-parse --verify HEAD)
echo "building..."
GOOS=linux go build -o aicomp  -ldflags "-X main.commithash=$commithash" .
echo "built."
echo "deploying..."
scp ./aicomp aicomp.io:~/bin/
scp -r ./static/* aicomp.io:~/static/
rm ./aicomp
echo "donezo."
