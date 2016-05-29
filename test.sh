#!/bin/bash

run_test() {
	go build
	setup_env
	trap clean_up EXIT
	go test
}

setup_env() {
	REDIS=$(docker run -d redis:3-alpine)

	REDIS_IP_ADDRESS=$(docker inspect -f '{{.NetworkSettings.IPAddress}}' ${REDIS})
	REDIS_PORT=6379

	docker exec -it ${REDIS} redis-cli ping > /dev/null

	run_slack_jobs
}

run_slack_jobs(){
	./slack-jobs -C example.ini -r ${REDIS_IP_ADDRESS}:${REDIS_PORT} &
	echo $! > slack-jobs.pid
}

clean_up() {
	kill `cat slack-jobs.pid`
	rm slack-jobs.pid
	docker stop ${REDIS} > /dev/null
	docker rm ${REDIS} > /dev/null
}

run_test $@
