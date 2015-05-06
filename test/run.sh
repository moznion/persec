#!/bin/bash

for t in $(ls test/test_*) ; do
  bash $t
done

