#!/bin/bash
# BATCH TESTING

if [ ! -z $1 ] 
then 
	DIR=$1
else
	TIME=`date '+%m-%d-%H%M%S'`
    DIR=results/$TIME/batch
fi

# Batch: test how batching improves throughput
for (( i = 2; i <= 64; i = i*2 )); do
	mkdir -p $DIR/"$i"r/config
	./scripts/generate_serv_conf.sh 3 $DIR/"$i"r/config $i
	# start
	./benchmarks/start_system.sh 3 64 $DIR/"$i"r

	# stop 
	sleep 5
	./benchmarks/stop_system.sh
	sleep 1
done