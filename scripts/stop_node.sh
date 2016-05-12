#!/bin/bash
# stop node $1 by killing its process


PID=$(ps | grep "[/]Users/heidi/go/bin/server" | awk '{print $1}' | tail -n +$1 | head -1)
kill $PID