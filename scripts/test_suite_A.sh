#!/bin/bash
# Run complete test suite

TIME=`date '+%m-%d-%H%M%S'`

# start servers
echo "Staring servers and clients"
./scripts/start_system.sh 3 1 results/$TIME/simple

# stop experiment
sleep 5
echo "ending experiment"
./scripts/stop_system.sh

# testing latency from 1 to 10 clients with a 3 server system
for (( i = 1; i < 15; i++ )); do
	DIR=results/$TIME/load/"$i"c
	mkdir -p $DIR
	# start
	./scripts/start_system.sh 3 $i $DIR

	# stop 
	sleep 5
	./scripts/stop_system.sh
	sleep 1
done



# testing latency for 3 to 15 servers with a single client
for (( i = 3; i <= 15; i++ )); do
	DIR=results/$TIME/scale/"$i"s
	mkdir -p $DIR
	# start
	./scripts/start_system.sh $i 1 $DIR

	# stop 
	sleep 5
	./scripts/stop_system.sh
	sleep 1
done

# testing master failure
DIR=results/$TIME/failure

# start servers
echo "Staring servers and clients"
./scripts/start_system.sh 5 1 $DIR

# stop node ID 0
sleep 0.1
echo "stopping node ID 0"
./scripts/stop_node.sh 0

# restart node ID 0
sleep 0.1
cd server

echo "restarting node ID 0"
../scripts/start_node.sh 0 ../$DIR
cd ..

# stop node ID 1
sleep 0.1
echo "stopping node ID 1"
./scripts/stop_node.sh 1

# restart node ID 1
sleep 0.1
cd server

echo "restarting node ID 1"
../scripts/start_node.sh 1 ../$DIR
cd ..

# stop experiment
sleep 5
echo "ending experiment"
./scripts/stop_system.sh

# Read: test simple system with different read percentages
DIR=results/$TIME/read
mkdir -p $DIR

for (( i = 0; i <= 100; i = i+5 )); do
	mkdir -p $DIR/"$i"r/config
	./scripts/generate_workload_conf.sh $i $DIR/"$i"r/config
	# start
	./scripts/start_system.sh 3 1 $DIR/"$i"r

	# stop 
	sleep 5
	./scripts/stop_system.sh
	sleep 1
done
