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
}

func main() {
	opt := new(opt)

	flag.IntVar(&opt.delta, "delta", 60, "Span as seconds to measure the throughput")
	flag.StringVar(&opt.pattern, "pattern", "", "A regexp pattern to filter the line. Filtering means this command measures throughput by matched lines only. If this option is unspecified, it doesn't filter.")
	flag.IntVar(&opt.limit, "limit", 0, "It measures the throughput until number which is specified by this option. If this option is zero or negative, it measures unlimited.")
	flag.StringVar(&opt.out, "out", "", "Output destination of throughput. If this option is unspecified, results will be written into STDOUT.")

	var f *os.File
	if output_path := opt.out; len(output_path) > 0 {
		f, err := os.OpenFile(output_path, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Errorf("%s", err)
			os.Exit(1)
		}
		defer f.Close()
	} else {
		f = os.Stdout
	}

	flag.Parse()

	var filter_re *regexp.Regexp
	filter_re = nil
	if pattern := opt.pattern; len(pattern) > 0 {
		filter_re, _ = regexp.Compile(pattern)
	}

	nl_re, _ := regexp.Compile("\r?\n")

	var counter uint64 = 0
	in_chan := make(chan []byte, 1)

	// register a worker
	go func() {
		for {
			go func(term []byte) {
				os.Stdout.Write(term)

				lines := nl_re.Split(string(term), -1)
				n := len(lines)

				// filter it here
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

	limit := opt.limit
	iteration := 1

	var wg sync.WaitGroup
	wg.Add(1)

	// time manager
	ticker := make(chan struct{}, 1)
	go func() {
		for {
			sec := opt.delta
			time.Sleep(time.Duration(sec) * time.Second)

			ticker <- struct{}{} // Pause to read from STDIN

			throughput := fmt.Sprintf("%f lines/sec\n", float64(counter)/float64(sec))
			f.WriteString(throughput)

			counter = 0
			ticker <- struct{}{} // Resume to read from STDIN

			iteration++
			if limit > 0 && iteration > limit {
				wg.Done()
			}
		}
	}()

	go func() {
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
				continue
			}

			n, err := os.Stdin.Read(buf)
			if err != nil {
				if err == io.EOF {
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
