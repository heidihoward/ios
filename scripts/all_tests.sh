#!/bin/bash
# Run complete test suite

TIME=`date '+%m-%d-%H%M%S'`

./scripts/simple_testing.sh results/$TIME/simple

./scripts/load_testing.sh results/$TIME/load

./scripts/scale_testing.sh results/$TIME/scale

./scripts/major_failure_testing.sh results/$TIME/failure

./scripts/read_testing.sh results/$TIME/read

./scripts/batch_testing.sh results/$TIME/batch

