package main

// сюда писать кодpackage main

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
	// crc32(data)+"~"+crc32(md5(data))
	var wg sync.WaitGroup
	throttle := make(chan struct{}, 1)
	throttle <- struct{}{}
	for data := range in {
		wg.Add(1)
		data := fmt.Sprintf("%v", data)
		fmt.Println("SingleHash data:", data)
		go func(dt string, o chan interface{}, th chan struct{}, w *sync.WaitGroup) {
			defer w.Done()
			var o1, o2 chan string

			o1 = make(chan string, 1)
			go func(d string) {
				crc32 := DataSignerCrc32(d)
				fmt.Println("SingleHash crc32(data):", crc32)
				o1 <- crc32
				close(o1)
			}(dt)

			o2 = make(chan string, 1)
			go func(d string, t chan struct{}) {
				<-t
				md5 := DataSignerMd5(d)
				t <- struct{}{}
				fmt.Println("SingleHash md5(data):", md5)
				o2 <- md5
			}(dt, th)
			go func() {
				md5 := <-o2
				crcmd := DataSignerCrc32(md5)
				fmt.Println("SingleHash crc32(md5(data)):", crcmd)
				o2 <- crcmd
				close(o2)
			}()

			result := <-o1 + "~" + <-o2
			fmt.Println("SingleHash result:", result)

			out <- result
		}(data, out, throttle, &wg)
	}
	wg.Wait()
	close(throttle)
}
