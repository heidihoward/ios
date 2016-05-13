#!/bin/bash
# LOAD TESTING

if [ ! -z $1 ] 
then 
	DIR=$1
else
	TIME=`date '+%m-%d-%H%M%S'`
    DIR=results/$TIME/load
fi

# testing latency from 1 to 10 clients with a 3 server system
for (( i = 1; i < 15; i++ )); do
	DIR2=$DIR/"$i"c
	mkdir -p $DIR2
	# start
	./benchmarks/start_system.sh 3 $i $DIR2

	# stop 
	sleep 5
	./benchmarks/stop_system.sh
	sleep 1
done