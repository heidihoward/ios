#!/bin/bash 
# generates a server config file called serv.conf for $1 nodes

echo "[peers]" > serv.conf

for ((id=0; id<$1; id++))
do
	echo "address = 127.0.0.1:809$id" >> serv.conf
done