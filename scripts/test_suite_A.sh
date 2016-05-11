#!/bin/bash
# Run complete test suite

TIME=`date '+%m-%d-%H%M%S'` 

# start servers
echo "Staring servers and clients"
./scripts/start_system.sh 3 1 scripts/results/$TIME/simple

# stop experiment
sleep 5
echo "ending experiment"
./scripts/stop_system.sh

# testing latency from 1 to 10 clients with a 3 server system
for (( i = 1; i < 15; i++ )); do

	# start
	./scripts/start_system.sh 3 $i scripts/results/$TIME/load

	# stop 
	sleep 5
	./scripts/stop_system.sh
	sleep 1
done



# testing latency for 3 to 15 servers with a single client
for (( i = 3; i <= 15; i++ )); do

	# start
	./scripts/start_system.sh $i 1 scripts/results/$TIME/scale

	# stop 
	sleep 5
	./scripts/stop_system.sh
	sleep 1
done