#!/bin/bash
set -ev

if [ "$TRAVIS_BRANCH" != "master" ]; then
	exit
fi

go get github.com/fzzy/radix/redis
go get github.com/glacjay/goini

./test.sh

CGO_ENABLED=0 go build -ldflags "-s" -a -installsuffix cgo -o slack-jobs

docker login -u="$DOCKER_USER" -p="$DOCKER_PASSWORD" -e="$DOCKER_EMAIL" $DOCKER_REGISTRY
docker build -t $DOCKER_REPO .
docker push $DOCKER_REPO
