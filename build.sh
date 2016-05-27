#!/bin/sh
set -ev

if [ "$TRAVIS_BRANCH" == "master" ]; then
	go get github.com/fzzy/radix/redis
	go get github.com/glacjay/goini

	CGO_ENABLED=0 go build -ldflags "-s" -a -installsuffix cgo -o slack-jobs

	docker login -u="$DOCKER_USER" -p="$DOCKER_PASSWORD" -e="$DOCKER_EMAIL" $DOCKER_REGISTRY
	docker build -t $DOCKER_REPO .
	docker push $DOCKER_REPO
fi
