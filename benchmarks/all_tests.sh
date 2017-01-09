#!/bin/bash
# Run complete test suite

TIME=`date '+%m-%d-%H%M%S'`

SRC=$GOPATH/src/github.com/heidi-ann/ios
cd $SRC

./benchmarks/simple_testing.sh results/$TIME/simple

./benchmarks/load_testing.sh results/$TIME/load

./benchmarks/scale_testing.sh results/$TIME/scale

./benchmarks/major_failure_testing.sh results/$TIME/failure

./benchmarks/read_testing.sh results/$TIME/read

./benchmarks/batch_testing.sh results/$TIME/batch

