#!/bin/bash

TIME=`date '+%m-%d-%H%M%S'`


./scripts/start_system.sh 3 1 scripts/results/$TIME -stderrthreshold=INFO

PID=$(ps | grep "[/]Users/heidi/go/bin/server" | awk '{print $1}' | head -1)
sleep 0.03
kill $PID

sleep 5
./scripts/stop_system.sh
