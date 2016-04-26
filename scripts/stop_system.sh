#!/bin/bash

# TODO: replace hardcode path with $GOPATH
kill $(ps | grep "[/]Users/heidi/go/bin/server" | awk '{print $1}')
kill $(ps | grep "[/]Users/heidi/go/bin/client" | awk '{print $1}')