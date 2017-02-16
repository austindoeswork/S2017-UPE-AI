#!/bin/bash
commithash=$(git rev-parse --verify HEAD)
GOOS=linux go build -ldflags "-X main.commithash=$commithash" .
