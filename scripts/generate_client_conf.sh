#!/bin/bash 
# generates a client config file called client.conf for $1 nodes with $2 timeout in dir $3

echo "[addresses]" > $3/client.conf

for ((id=0; id<$1; id++))
do
	if (($id<10))
	then
		port=3310$id
	else
		port=331$id
	fi

	echo "address = 127.0.0.1:$port" >> $3/client.conf
done

echo "[parameters]
retries = 1
timeout = $2
" >> $3/client.conf