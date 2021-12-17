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
	// crc32(data)+"~"+crc32(md5(data))
	var wg sync.WaitGroup
	throttle := make(chan struct{}, 1)
	throttle <- struct{}{}

	for data := range in {
		data := fmt.Sprintf("%v", data)
		fmt.Println("SingleHash data:", data)
		wg.Add(1)

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

func MultiHash(in, out chan interface{}) {
	// crc32(th+data)) th=0..5
	var wg sync.WaitGroup
	for d := range in {
		data := fmt.Sprintf("%v", d)
		fmt.Println("MultiHash data:", data)
		wg.Add(1)

		go func(dt string, ot chan interface{}, w *sync.WaitGroup) {
			defer w.Done()
			result := ""
			var results [6]chan string

			for i := 0; i < 6; i++ {
				results[i] = make(chan string, 1)
				go func(o chan string, th int, d string) {
					val := DataSignerCrc32(strconv.Itoa(th) + d)
					fmt.Println("MultiHash crc32(th+data):", val)
					o <- val
				}(results[i], i, dt)

			}
			for _, c := range results {
				result += <-c
			}
			fmt.Println("MultiHash result:", result)

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
	fmt.Println("Combine result:", result)
	out <- result
}

func main() {
	testResult := "NOT_SET"
	inputData := []int{0, 1}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				fmt.Println("pushing to out channel value", fibNum)
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				fmt.Println("cant convert result data to string")
				return
			}
			testResult = data
		}),
	}
	ExecutePipeline(hashSignJobs...)
}
