#!/bin/bash

TIME=`date '+%m-%d-%H%M%S'`

# start servers
echo "Staring servers and clients"
./scripts/start_system.sh 5 1 scripts/results/$TIME

# stop node ID 0
sleep 0.03
echo "stopping node ID 0"
./scripts/stop_node.sh 0

# restart node ID 0
sleep 0.03
cd server

echo "restarting node ID 0"
../scripts/start_node.sh 0 scripts/results/$TIME/5s1c/s0.log &
cd ..

# stop experiment
sleep 5
echo "ending experiment"
./scripts/stop_system.sh
