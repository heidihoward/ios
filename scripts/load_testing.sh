#!/bin/bash

# testing latency from 1 to 10 clients

for (( i = 1; i < 10; i++ )); do
	./scripts/simple_test.sh 3 $i
done