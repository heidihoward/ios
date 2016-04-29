#!/bin/bash

TIME=`date '+%m-%d-%H%M%S'`

# start servers
echo "Staring servers and clients"
./scripts/start_system.sh 5 1 scripts/results/$TIME

# stop node ID 0
PID=$(ps | grep "[/]Users/heidi/go/bin/server" | awk '{print $1}' | head -1)
sleep 0.03
echo "stopping node ID 0"
kill $PID

# restart node ID 0
sleep 0.03
cd server

echo "restarting node ID 0"
$GOPATH/bin/server -id=0 -client-port=8080 -peer-port=8090 -config=../scripts/serv.conf -log_dir=../scripts/results/$TIME/3s1c/s2.log &
cd ..

# stop experiment
sleep 5
echo "ending experiment"
./scripts/stop_system.sh
