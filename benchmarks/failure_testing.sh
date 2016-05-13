#!/bin/bash

TIME=`date '+%m-%d-%H%M%S'`

# start servers
echo "Staring servers and clients"
./benchmarks/start_system.sh 3 1 scripts/results/$TIME

# stop node ID 2
PID=$(ps | grep "[/]Users/heidi/go/bin/server" | awk '{print $1}' | head -3 | tail -1)
sleep 0.03
echo "stopping node ID 2"
kill $PID

# restart node ID 2
sleep 0.03
cd server

echo "restarting node ID 2"
$GOPATH/bin/server -id=2 -client-port=8082 -peer-port=8092 -config=../scripts/serv.conf -log_dir=../scripts/results/$TIME/3s1c/s2.log &
cd ..

# stop experiment
sleep 5
echo "ending experiment"
./scripts/stop_system.sh
