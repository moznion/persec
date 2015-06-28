#!/bin/bash

set -ex

. ./test/util.sh

go build

bash ./test/log.sh &
sleep 1s # to wait to open a log file

tail -f log_for_test | ./persec --delta 1 --pattern "^foobar" > test.tmp &

sleep 5s

kill $(jobs -p)

throughput_row_num=`grep 0.00\ lines/sec test.tmp | wc -l | awk '{print $1}'`
if [ $throughput_row_num -lt 4 ] ; then
  echo "Number of throughput row is not matched to expected"
  echo_red "FAIL: ${0##*/}"
  exit 1
fi

throughput_row_num=`grep 1.00\ lines/sec test.tmp | wc -l | awk '{print $1}'`
if [ $throughput_row_num -ne 1 ] ; then
  echo "Number of throughput row is not matched to expected"
  echo_red "FAIL: ${0##*/}"
  exit 1
fi

# finalize
rm log_for_test
rm test.tmp

echo_green "PASS: ${0##*/}"
exit 0

