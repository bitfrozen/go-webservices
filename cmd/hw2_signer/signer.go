package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	var input, output chan interface{}
	var wg sync.WaitGroup
	wg.Add(len(jobs))

	for _, jb := range jobs {
		output = make(chan interface{}, 100)

		go func(in, out chan interface{}, j job, w *sync.WaitGroup) {
			defer w.Done()
			defer close(out)
			j(in, out)
		}(input, output, jb, &wg)

		// If function signature would allow returning output channel, we could assign return value here,
		// but since we work with function parameters only, we switch channels here, giving next job in line
		// closed channel and creating new output channel at the start of the cycle
		input = output
	}

	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	throttle := make(chan struct{}, 1)

	for data := range in {
		data := fmt.Sprintf("%v", data)
		wg.Add(1)

		go func(dt string, o chan interface{}, th chan struct{}, w *sync.WaitGroup) {
			defer w.Done()
			var o1, o2 chan string

			o1 = make(chan string, 1)
			go func(d string) {
				crc32 := DataSignerCrc32(d)
				o1 <- crc32
				close(o1)
			}(dt)

			o2 = make(chan string, 1)
			go func(d string, t chan struct{}) {
				t <- struct{}{}
				md5 := DataSignerMd5(d)
				<-t
				o2 <- md5
			}(dt, th)
			go func() {
				md5 := <-o2
				crcmd := DataSignerCrc32(md5)
				o2 <- crcmd
				close(o2)
			}()

			result := <-o1 + "~" + <-o2
			out <- result
		}(data, out, throttle, &wg)
	}

	wg.Wait()
	close(throttle)
}

func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	for d := range in {
		data := fmt.Sprintf("%v", d)
		wg.Add(1)

		go func(dt string, ot chan interface{}, w *sync.WaitGroup) {
			defer w.Done()
			var results [6]chan string
			for i := 0; i < 6; i++ {
				results[i] = make(chan string, 1)
				go func(o chan string, th int, d string) {
					val := DataSignerCrc32(strconv.Itoa(th) + d)
					o <- val
				}(results[i], i, dt)

			}

			var result string
			for _, c := range results {
				result += <-c
				close(c)
			}
			ot <- result
		}(data, out, &wg)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var results []string

	for d := range in {
		data := fmt.Sprintf("%v", d)
		results = append(results, data)
	}
	sort.Strings(results)
	result := strings.Join(results, "_")

	out <- result
}

func main() {

}
