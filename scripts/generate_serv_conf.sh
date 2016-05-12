#!/bin/bash 
# generates a server config file called serv.conf for $1 nodes

echo "[peers]" > serv.conf

for ((id=0; id<$1; id++))
do
	if (($id<10))
	then
		port=3330$id
	else
		port=333$id
	fi
	echo "address = 127.0.0.1:$port" >> serv.conf
done

echo "[options]
length = 100000
" >> serv.conf