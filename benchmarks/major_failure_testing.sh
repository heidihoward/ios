#!/bin/bash
# FAILURE TESTING

if [ ! -z $1 ] 
then 
	DIR=$1
else
	TIME=`date '+%m-%d-%H%M%S'`
    DIR=results/$TIME/failure
fi


# start servers
echo "Staring servers and clients"
./benchmarks/start_system.sh 5 1 $DIR

# stop node ID 0
sleep 0.1
echo "stopping node ID 0"
./benchmarks/stop_node.sh 0

# restart node ID 0
sleep 0.1
cd server

echo "restarting node ID 0"
../benchmarks/start_node.sh 0 ../$DIR
cd ..

# stop node ID 1
sleep 0.1
echo "stopping node ID 1"
./benchmarks/stop_node.sh 1

# restart node ID 1
sleep 0.1
cd server

echo "restarting node ID 1"
../benchmarks/start_node.sh 1 ../$DIR
cd ..

# stop experiment
sleep 5
echo "ending experiment"
./benchmarks/stop_system.sh