#!/bin/bash

TIME=`date '+%m-%d-%H%M%S'`

# testing latency from 1 to 10 clients with a 3 server system
for (( i = 1; i < 10; i++ )); do

	# start
	./scripts/start_system.sh 3 $i scripts/results/$TIME

	# stop 
	sleep 5
	./scripts/stop_system.sh
done