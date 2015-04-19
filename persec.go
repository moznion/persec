package main

import "os"
import "io"
import "sync"
import "sync/atomic"
import "regexp"
import "fmt"
import "time"

func main() {
	nl_regex, err := regexp.Compile("\r?\n")
	if err != nil {
		fmt.Errorf("%s", err)
		os.Exit(1)
	}

	var counter uint64 = 0
	in_chan := make(chan []byte, 1)

	// register a worker
	go func() {
		for {
			go func(term []byte) {
				os.Stdout.Write(term)

				// TODO filter here

				n := len(nl_regex.FindAll(term, -1))
				atomic.AddUint64(&counter, uint64(n))
			}(<-in_chan)
		}
	}()

	// time manager
	ticker := make(chan struct{}, 1)
	go func() {
		for {
			sec := 60 // TODO
			time.Sleep(time.Duration(sec) * time.Second)

			ticker <- struct{}{} // Pause to read from STDIN

			fmt.Printf("%f rows/sec\n", float64(counter)/float64(sec)) // TODO

			counter = 0
			ticker <- struct{}{} // Resume to read from STDIN
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

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

		// STDIN reader
		buf := make([]byte, 4096)
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
