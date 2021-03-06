persec [![Build Status](https://travis-ci.org/moznion/persec.svg?branch=master)](https://travis-ci.org/moznion/persec) [![wercker status](https://app.wercker.com/status/58831f8aea401a8e2209351e359988f8/s/master "wercker status")](https://app.wercker.com/project/bykey/58831f8aea401a8e2209351e359988f8)
==

A command to measure the throughput of per-sec which is on the basis of number of lines from STDIN.

![persec_demo](https://dl.dropboxusercontent.com/u/14832699/persec.gif)

Usage
--

### Basic

```sh
$ some_command | persec [Options]
```

### Bar chart mode

```sh
$ some_command | persec --chart 100 [Other Options]
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
-  --timestamp=false: Prepend timestamp
-  --chart=-1: Show throughput as a bar chart. This option receives int value as a maximum value of a chart. Default value of -1 means disable the chart mode. If 0 value is set, it will sample 5 time to determine the value of 100%.
- --color="reset": Colorize output. You can use colors which are supported by github.com/mgutz/ansi

How to install
--

### By `go get`

```
$ go get github.com/moznion/persec
```

### From GitHub Releases

Access to [https://github.com/moznion/persec/releases](https://github.com/moznion/persec/releases)
and get an archive which is suitable your architecture.

Author
--

moznion (<moznion@gmail.com>)

License
--

MIT

