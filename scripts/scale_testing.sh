#!/bin/bash
# SCALE TESTING

if [ ! -z $1 ] 
then 
	DIR=$1
else
	TIME=`date '+%m-%d-%H%M%S'`
    DIR=results/$TIME/scale
fi

# testing latency for 3 to 15 servers with a single client
for (( i = 3; i <= 15; i++ )); do
	DIR2=$DIR/"$i"s
	mkdir -p $DIR2
	# start
	./scripts/start_system.sh $i 1 $DIR2

	# stop 
	sleep 5
	./scripts/stop_system.sh
	sleep 1
done