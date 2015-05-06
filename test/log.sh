#!/bin/bash

i=1
while true
do
  date=`LC_TIME="C" date`
  echo "[$date] sample log" >> log_for_test
  if [ `expr $i % 5` -eq 0 ] ; then
    echo "foobarbuz" >> log_for_test
  fi
  sleep 1

  i=`expr $i + 1`
done

