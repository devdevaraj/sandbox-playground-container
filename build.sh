#!/bin/bash

IMAGE=${1:-"repo.synnefo.in/dev/playground"}

(cd ./firestarter && go build .)
# (cd ./firestarter && go build .)
# (cd ./wss && go build .)

sudo docker build -t $IMAGE . --

sudo docker push $IMAGE