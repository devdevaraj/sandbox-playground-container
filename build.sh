#!/bin/bash

IMAGE=${1:-"repo.synnefo.solutions/devaraj/playground"}

(cd ./new_firestarter && go build .)
# (cd ./firestarter && go build .)
# (cd ./wss && go build .)

sudo docker build -t $IMAGE .