#!/bin/bash 
# generates a client config file called client.conf for $1 nodes with $2 timeout

echo "[addresses]" > client.conf

for ((id=0; id<$1; id++))
do
	if (($id<10))
	then
		port=3310$id
	else
		port=331$id
	fi

	echo "address = 127.0.0.1:$port" >> client.conf
done

echo "[parameters]
retries = 1
timeout = $2
" >> client.conf