#!/bin/bash          
# simple test, 1 server, 3 clients, no failures locally

# tidy up from previous tests
cd $GOPATH/src/github.com/heidi-ann/hydra

rm server/*.temp
rm scripts/results/*

# start server
cd server
$GOPATH/bin/server -id=0 -client-port=8080 -peer-port=8090 &
$GOPATH/bin/server -id=1 -client-port=8081 -peer-port=8091 &
$GOPATH/bin/server -id=2 -client-port=8082 -peer-port=8092 &

# start clients 
cd ../client
$GOPATH/bin/client -id=0 -mode=test -stat=../scripts/results/latency_0.csv &
$GOPATH/bin/client -id=1 -mode=test -stat=../scripts/results/latency_1.csv &
$GOPATH/bin/client -id=2 -mode=test -stat=../scripts/results/latency_2.csv &

# stop 
sleep 20
kill $(jobs -p)

# produce CDF of latency