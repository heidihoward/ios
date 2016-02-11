#!/bin/bash          
# simple test, 1 server, 3 clients, no failures locally

# tidy up from previous tests
cd $GOPATH/src/github.com/heidi-ann/hydra

rm server/persistent.log
rm client/latency.csv

# start server
$GOPATH/bin/server &

# start clients 
$GOPATH/bin/client -config=client/example.conf -auto=test/workload.conf -stat=latency_1.csv &
$GOPATH/bin/client -config=client/example.conf -auto=test/workload.conf -stat=latency_2.csv &
$GOPATH/bin/client -config=client/example.conf -auto=test/workload.conf -stat=latency_3.csv &

# stop clients

# stop server

# produce CDF of latency