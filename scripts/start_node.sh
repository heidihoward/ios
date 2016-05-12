#!/bin/bash
# (re)starting node $1. logging to dir $2

# pad node ID for ports
id=$1

if (($id<10))
then
	port=0$id
else
	port=$id
fi

# makr logging dir
mkdir -p $2

# start server
$GOPATH/bin/server -id=$id -client-port=331$port -peer-port=333$port -config=../scripts/serv.conf -log_dir=$2 &

