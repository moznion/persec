package main

import "os"
import "io"
import "sync"
import "sync/atomic"
import "regexp"
import "fmt"
import "flag"
import "time"

type opt struct {
	delta   int
	pattern string
	limit   int
	out     string
	help    bool
}

func main() {
	o := parseOpt()
	run(o)
}

func parseOpt() *opt {
	flag.Usage = func() {
		fmt.Printf(`Usage:
  some_command | persec [Options]

Options:
`)
		flag.PrintDefaults()
	}

	o := new(opt)
	flag.IntVar(&o.delta, "delta", 60, "Span as seconds to measure the throughput")
	flag.StringVar(&o.pattern, "pattern", "", "A regexp pattern to filter the line. Filtering means this command measures throughput by matched lines only. If this option is unspecified, it doesn't filter.")
	flag.IntVar(&o.limit, "limit", 0, "It measures the throughput until number which is specified by this option. If this option is zero or negative, it measures unlimited.")
	flag.StringVar(&o.out, "out", "", "Output destination of throughput. If this option is unspecified, results will be written into STDOUT.")
	flag.BoolVar(&o.help, "help", false, "Show helps")

	flag.Parse()

	return o
}

func run(o *opt) {
	if o.help {
		flag.Usage()
		os.Exit(0)
	}

	var f *os.File
	if output_path := o.out; len(output_path) > 0 {
		// open a file which is specified by option with append mode
		f, err := os.OpenFile(output_path, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Errorf("%s", err)
			os.Exit(1)
		}
		defer f.Close()
	} else {
		// choose STDOUT as a destination
		f = os.Stdout
	}

	var filter_re *regexp.Regexp
	filter_re = nil
	if pattern := o.pattern; len(pattern) > 0 {
		filter_re, _ = regexp.Compile(pattern)
	}

	nl_re, _ := regexp.Compile("\r?\n") // line splitter

	var counter uint64 = 0
	in_chan := make(chan []byte, 1)

	// counter
	go func() {
		for {
			go func(term []byte) {
				os.Stdout.Write(term)

				lines := nl_re.Split(string(term), -1)
				n := len(lines)

				// Apply filtering here
				if filter_re != nil {
					n = 0
					for _, line := range lines {
						if filter_re.MatchString(line) {
							n++
						}
					}
				}

				atomic.AddUint64(&counter, uint64(n))
			}(<-in_chan)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	limit := o.limit
	iteration := 1

	// time manager
	ticker := make(chan struct{}, 1)
	sec := o.delta
	go func() {
		for {
			time.Sleep(time.Duration(sec) * time.Second)

			ticker <- struct{}{} // Pause to read from STDIN

			throughput := fmt.Sprintf("%f lines/sec\n", float64(counter)/float64(sec))
			f.WriteString(throughput)

			counter = 0
			ticker <- struct{}{} // Resume to read from STDIN

			iteration++
			if limit > 0 && iteration > limit {
				// Terminate if `limit` is specified through opt and
				// iteration overs that
				wg.Done()
			}
		}
	}()

	// Read from STDIN and fire event for counter
	go func() {
		defer wg.Done()

		should_wait := false
		go func() {
			for {
				go func(tick struct{}) {
					should_wait = !should_wait
				}(<-ticker)
			}
		}()

		// Read from STDIN
		buf := make([]byte, 1000000)
		for {
			if should_wait == true {
				// Block to read from STDIN while outputting throughput result
				continue
			}

			n, err := os.Stdin.Read(buf)
			if err != nil {
				if err == io.EOF {
					throughput := fmt.Sprintf("%f lines/sec\n", float64(counter)/float64(sec)) // XXX inaccuracy
					f.WriteString(throughput)
					break
				}
				fmt.Errorf("%s", err)
				os.Exit(1)
			}
			in_chan <- buf[:n]
		}
	}()

	wg.Wait()
}
