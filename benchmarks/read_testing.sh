#!/bin/bash
# READ TESTING

if [ ! -z $1 ] 
then 
	DIR=$1
else
	TIME=`date '+%m-%d-%H%M%S'`
    DIR=results/$TIME/read
fi

# Read: test simple system with different read percentages
for (( i = 0; i <= 100; i = i+5 )); do
	mkdir -p $DIR/"$i"r/config
	./scripts/generate_workload_conf.sh $i $DIR/"$i"r/config
	# start
	./benchmarks/start_system.sh 3 3 $DIR/"$i"r

	# stop 
	sleep 10
	./benchmarks/stop_system.sh
	sleep 1
done