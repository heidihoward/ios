#!/bin/bash

TIME=`date '+%m-%d-%H%M%S'`

# testing latency for 3 to 9 servers with a single client
for (( i = 3; i <= 9; i = i+2 )); do

	# start
	./scripts/start_system.sh $i 1 scripts/results/$TIME

	# stop 
	sleep 5
	./scripts/stop_system.sh
done