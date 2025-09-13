#!/bin/bash

(cd ./firestarter && go build .)
(cd ./wss && go build .)

docker build -t repo.synnefo.solutions/devaraj/playground .