#!/bin/bash

function do_start {
	${PROGRAM_PATH} -C ${CONFIG_PATH} 2>&1 > /dev/null &
	pid=$!
	echo "${pid}" > ${PID_FILE}
	disown ${pid}
}

function do_stop {
	pid=$(cat ${PID_FILE})	
	kill ${pid}
	rm ${PID_FILE}
}

function get_status {
	if [ -f "${PID_FILE}" ]; then
		pid=$(cat ${PID_FILE})	
		kill -0 "${pid}" 2> /dev/null
		return $?
	else
		return 1
	fi
}

{
cd "$(dirname $0)"
# start | stop | restart
CMD=$1
PROGRAM=slack-jobs
PROGRAM_PATH=./slack-jobs
CONFIG_PATH=./slack-jobs.ini
PID_FILE=./slack-jobs.pid

case $CMD in
start)
	get_status || do_start
	;;
stop)
	get_status && do_stop
	;;
restart)
	get_status && do_stop
	sleep 1
	do_start
	;;
status)
	get_status
	ec=$?
	if [ "${ec}" = "0" ]; then
		echo "${PROGRAM} is running"
	else
		echo "${PROGRAM} is NOT running"
	fi
	;;
*)
	echo "Usage: ${0} {start|stop|restart|status}"
	;;
esac

exit
}
