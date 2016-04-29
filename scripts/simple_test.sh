#!/bin/bash

TIME=`date '+%m-%d-%H%M%S'`

# start servers
echo "Staring servers and clients"
./scripts/start_system.sh 3 1 scripts/results/$TIME

# stop experiment
sleep 5
echo "ending experiment"
./scripts/stop_system.sh
