#!/bin/bash

set -ex

. ./test/util.sh

go build

bash ./test/log.sh &
sleep 1s # to wait to open a log file

tail -f log_for_test | ./persec --delta 1 > test.tmp &

sleep 5s

kill $(jobs -p)

throughput_row_num=`grep lines/sec test.tmp | wc -l | awk '{print $1}'`

if [ $throughput_row_num -lt 5 ] ; then
  echo "Number of throughput row is not matched to expected"
  echo_red "FAIL: ${0##*/}"
  exit 1
fi

# finalize
rm log_for_test
rm test.tmp

echo_green "PASS: ${0##*/}"
exit 0

