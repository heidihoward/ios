#!/bin/bash 
# generates a workload config file called workload.conf for $1 reads in dir $2

echo "[commands]
reads = $1
conflicts = 2
interval = 0

[termination]
requests = 1000
" >> $2/workload.conf