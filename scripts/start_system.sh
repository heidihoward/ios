#!/bin/bash          
# simple test with no failures, no failures locally

# first args is number of servers (don't forget to change server config file)
# 2nd args is number of client
# 3rd arg is path from hydra of where to store results

# tidy up from previous tests
SRC=$GOPATH/src/github.com/heidi-ann/hydra
cd $SRC

# make results directory
mkdir -p $3/$1s$2c
cd $3/$1s$2c
mkdir results
mkdir logs
mkdir disk
mkdir config

# generate server and client configuration files
$SRC/scripts/generate_serv_conf.sh $1 config
$SRC/scripts/generate_client_conf.sh $1 500 config

# start servers
echo "starting $1 servers"
for ((id=0; id<$1; id++))
do
	# make logging directory for server
	mkdir logs/s$id.log
	# padding for ports
	if (($id<10))
	then
		port=0$id
	else
		port=$id
	fi
	# start server
	$GOPATH/bin/server -id=$id -client-port=331$port -peer-port=333$port -config=config/serv.conf -log_dir=logs/s$id.log -disk=disk &
done

sleep 1

# start clients 
echo "starting $2 clients"
for ((id=1; id<=$2; id++))
do
	# make logging/results directory for client
	mkdir logs/c$id.log
	# start client
	$GOPATH/bin/client -id=$id -mode=test -stat=results/latency_$id.csv -log_dir=logs/c$id.log -config=config/client.conf -auto=$SRC/test/workload.conf &
done

echo "setup complete, recording results in $3/$1s$2c"
