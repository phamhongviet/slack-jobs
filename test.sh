#!/bin/bash
set -ev

run_test() {
	go build
	setup_env
	run_slack_jobs
	go test
	clean_up
}

setup_env() {
	REDIS=$(docker run -d -P redis:3-alpine)

	REDIS_IP_ADDRESS=$(docker inspect ${REDIS} | jq -r '.[] | .NetworkSettings.IPAddress')
	REDIS_PORT=$(docker inspect ${REDIS} | jq -r '.[] | .NetworkSettings.Ports."6379/tcp" | .[0] | .HostPort')
}

run_slack_jobs(){
	./slack-jobs -C example.ini -r ${REDIS_IP_ADDRESS}:${REDIS_PORT} &
	echo $! > slack-jobs.pid
}

clean_up() {
	kill `cat slack-jobs.pid`
	rm slack-jobs.pid
	docker stop ${REDIS}
	docker rm ${REDIS}
}
