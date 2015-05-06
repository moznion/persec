persec [![Build Status](https://travis-ci.org/moznion/persec.svg?branch=master)](https://travis-ci.org/moznion/persec)
==

A command to measure the throughput of per-sec which is on the basis of number of lines from STDIN.

![persec_demo](https://dl.dropboxusercontent.com/u/14832699/persec.gif)

Usage
--

```sh
$ some_command | persec [Options]
```

Description
--

This tool measures the throughput of per-sec which is on the basis of number of lines from STDIN (e.g. piping with `tail -f something`).
Measuring mechanism is really simple;

1. Count the number of lines while fixed interval (default 60sec).
2. Counted number divide by fixed interval, then it is enable to evaluate throughput of per-sec.

Default, this command also output lines which is from STDIN (like a `tee` command).
If you don't want to do that, please append `--notee` option.

Options
--

-  --delta=60: Interval as seconds to measure the throughput
-  --help=false: Show helps
-  --limit=0: It measures the throughput until number which is specified by this option. If this option is zero or negative, it measures unlimited.
-  --notee=false: Don't tee if this option is true
-  --out="": Output destination of throughput. If this option is unspecified, results will be written into STDOUT.
-  --pattern="": A regexp pattern to filter the line. Filtering means this command measures throughput by matched lines only. If this option is unspecified, it doesn't filter.

Author
--

moznion (<moznion@gmail.com>)

License
--

MIT

