#!/bin/bash
# TODO: replace hardcode path with $GOPATH

# kill servers
kill $(ps | grep "[/]Users/heidi/go/bin/server" | awk '{print $1}')

# kill clients
CLIENTS=$(ps | grep "[/]Users/heidi/go/bin/client" | awk '{print $1}')

if [ -n "$CLIENTS" ]  
then
  echo "Experiment unsuccessful"
  kill $CLIENTS
else  
  echo "Experiment successful"
fi 
