#!/bin/bash          
# simple test with no failures, no failures locally

# first args is number of servers (don't forget to change server config file)
# 2nd args is number of client
# 3rd arg is path from hydra of where to store results

# tidy up from previous tests
cd $GOPATH/src/github.com/heidi-ann/hydra

rm server/*.temp
rm scripts/serv.conf

# make results directory
mkdir -p $3/$1s$2c

# generate server and client configuration files
cd scripts
./generate_serv_conf.sh $1
./generate_client_conf.sh $1 500
cd ..

# start servers
cd server
echo "starting $1 servers"
for ((id=0; id<$1; id++))
do
	# make logging directory for server
	mkdir ../$3/$1s$2c/s$id.log
	# padding for ports
	if (($id<10))
	then
		port=0$id
	else
		port=$id
	fi
	# start server
	$GOPATH/bin/server -id=$id -client-port=331$port -peer-port=333$port -config=../scripts/serv.conf -log_dir=../$3/$1s$2c/s$id.log &
done

sleep 1

# start clients 
cd ../client
echo "starting $2 clients"
for ((id=1; id<=$2; id++))
do
	# make logging directory for client
	mkdir ../$3/$1s$2c/c$id.log
	# start client
	$GOPATH/bin/client -id=$id -mode=test -stat=../$3/$1s$2c/latency_$id.csv -log_dir=../$3/$1s$2c/c$id.log -config=../scripts/client.conf &
done

echo "setup complete, recording results in $3/$1s$2c"
