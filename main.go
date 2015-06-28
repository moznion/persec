package main

import "log"
import "os"
import "io"
import "strings"
import "sync"
import "sync/atomic"
import "regexp"
import "fmt"
import "flag"
import "time"

import "github.com/mgutz/ansi"

type opt struct {
	delta   int
	pattern string
	limit   int
	out     string
	help    bool
	notee   bool
	chart   int
	color   string
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
	flag.IntVar(&o.delta, "delta", 60, "Interval as seconds to measure the throughput")
	flag.StringVar(&o.pattern, "pattern", "", "A regexp pattern to filter the line. Filtering means this command measures throughput by matched lines only. You can use golang's regular expression as this option. If this option is unspecified, it doesn't filter.")
	flag.IntVar(&o.limit, "limit", 0, "It measures the throughput until number which is specified by this option. If this option is zero or negative, it measures unlimited.")
	flag.StringVar(&o.out, "out", "", "Output destination of throughput. If this option is unspecified, results will be written into STDOUT.")
	flag.BoolVar(&o.notee, "notee", false, "Don't tee if this option is true")
	flag.BoolVar(&o.help, "help", false, "Show helps")
	flag.IntVar(&o.chart, "chart", 0, "Show throughput as a bar chart. This option receives int value as a maximum value of a chart.")
	flag.StringVar(&o.color, "color", "reset", "Colorize output. You can use colors which are supported by github.com/mgutz/ansi")

	flag.Parse()

	return o
}

func run(o *opt) {
	if o.help {
		flag.Usage()
		os.Exit(0)
	}

	var f *os.File
	if outputPath := o.out; len(outputPath) > 0 {
		var err error

		// open a file which is specified by option with append mode
		f, err = os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		defer f.Close()
	} else {
		// choose STDOUT as a destination
		f = os.Stdout
	}

	var filterRe *regexp.Regexp
	filterRe = nil
	if pattern := o.pattern; len(pattern) > 0 {
		filterRe, _ = regexp.Compile(pattern)
	}

	nlRe, _ := regexp.Compile("\r?\n") // line splitter

	var counter uint64
	inChan := make(chan []byte, 1)

	// counter
	tee := !o.notee
	go func() {
		for {
			go func(term []byte) {
				if tee {
					os.Stdout.Write(term)
				}

				lines := nlRe.Split(string(term), -1)
				n := len(lines) - 1 // `- 1`: to ignore last empty line

				// Apply filtering here
				if filterRe != nil {
					n = 0
					for _, line := range lines {
						if filterRe.MatchString(line) {
							n++
						}
					}
				}

				atomic.AddUint64(&counter, uint64(n))
			}(<-inChan)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	limit := o.limit
	iteration := 1

	// time manager
	ticker := make(chan struct{}, 1)
	delta := o.delta
	go func() {
		for {
			time.Sleep(time.Duration(delta) * time.Second)

			ticker <- struct{}{} // Pause to read from STDIN

			throughput := float64(counter) / float64(delta)
			var result string

			if o.chart > 0 {
				percentage := throughput / float64(o.chart) * 100
				meter := int64(percentage) / 5
				over := " "

				if int(percentage)%5*2 >= 5 { // round off
					meter++
				}
				if meter >= 20 { // cut off
					meter = 20
					over = "="
				}

				result = fmt.Sprintf("%6.2f%% [%s%s]%s %.2f lines/sec\n",
					percentage, strings.Repeat("=", int(meter)), strings.Repeat(" ", 20-int(meter)),
					ansi.Color(over, "red"), throughput)
			} else {
				result = ansi.Color(fmt.Sprintf("%.2f lines/sec\n", throughput), o.color)
			}

			_, err := f.WriteString(result)
			if err != nil {
				log.Fatal(err)
				f.Close()
				os.Exit(1)
			}

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

	shouldWait := false
	go func() {
		for {
			go func(tick struct{}) {
				shouldWait = !shouldWait
			}(<-ticker)
		}
	}()

	// Read from STDIN and fire event for counter
	go func() {
		defer wg.Done()

		// Read from STDIN
		buf := make([]byte, 1000000)
		for {
			if shouldWait == true {
				// Block to read from STDIN while outputting throughput result
				continue
			}

			n, err := os.Stdin.Read(buf)
			if err != nil {
				if err == io.EOF {
					throughput := float64(counter) / float64(delta)
					var result string

					// XXX inaccuracy
					if o.chart > 0 {
						percentage := throughput / float64(o.chart) * 100
						meter := int64(percentage) / 5
						over := " "

						if int(percentage)%5*2 >= 5 { // round off
							meter++
						}
						if meter >= 20 { // cut off
							meter = 20
							over = "="
						}

						result = fmt.Sprintf("%6.2f%% [%s%s]%s %.2f lines/sec\n",
							percentage, strings.Repeat("=", int(meter)), strings.Repeat(" ", 20-int(meter)),
							ansi.Color(over, "red"), throughput)
					} else {
						result = ansi.Color(fmt.Sprintf("%.2f lines/sec\n", throughput), o.color)
					}

					f.WriteString(result)
					break
				}
				log.Fatal(err)
				f.Close()
				os.Exit(1)
			}
			inChan <- buf[:n]
		}
	}()

	wg.Wait()
}
