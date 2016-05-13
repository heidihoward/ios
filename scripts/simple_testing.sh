#!/bin/bash
# SIMPLE TEST

if [ ! -z $1 ] 
then 
	DIR=$1
else
	TIME=`date '+%m-%d-%H%M%S'`
    DIR=results/$TIME/simple
fi

# start servers
echo "Staring servers and clients"
./scripts/start_system.sh 3 1 $DIR

# stop experiment
sleep 5
echo "ending experiment"
./scripts/stop_system.sh