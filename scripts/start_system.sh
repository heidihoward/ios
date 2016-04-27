#!/bin/bash          
# simple test with no failures, no failures locally

# first args is number of servers (don't forget to change server config file)
# 2nd args is number of client
# 3rd arg is path from hydra of where to store results
# 4th arg is logging arguments

# tidy up from previous tests
cd $GOPATH/src/github.com/heidi-ann/hydra

rm server/*.temp
rm scripts/serv.conf

# make results directory
mkdir -p $3/$1s$2c

# generate server configuration files
cd scripts
./generate_serv_conf.sh $1
cd ..

# start servers
cd server
echo "starting $1 servers"
for ((id=0; id<$1; id++))
do
	$GOPATH/bin/server -id=$id -client-port=808$id -peer-port=809$id -config=../scripts/serv.conf $4 &
done

sleep 1

# start clients 
cd ../client
echo "starting $2 clients"
for ((id=1; id<=$2; id++))
do
	$GOPATH/bin/client -id=$id -mode=test -stat=../$3/$1s$2c/latency_$id.csv $4 &
done

echo "setup complete, recording results in $3/$1s$2c"
