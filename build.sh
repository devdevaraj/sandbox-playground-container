#!/bin/bash

IMAGE=${1:-"repo.synnefo.solutions/devaraj/playground"}

(cd ./firestarter && go build .)
(cd ./wss && go build .)

docker build -t $IMAGE .