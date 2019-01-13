package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 100)

	for _, oneJob := range jobs {
		wg.Add(1)
		out := make(chan interface{}, 100)
		go func(jobFunction job, in, out chan interface{}) {
			defer close(out)
			defer wg.Done()
			jobFunction(in, out)
		}(oneJob, in, out)
		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	md5Wait := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for data := range in {
		wg.Add(1)
		dataString := strconv.Itoa(data.(int))

		go func(input string) {
			md5dataString := make(chan string)
			crc32dataString := make(chan string, 1)
			crc32md5dataString := make(chan string, 1)

			go func(str string) {
				md5Wait.Lock()
				md5dataString <- DataSignerMd5(str)
				md5Wait.Unlock()
			}(input)

			go func(str string) {
				crc32dataString <- DataSignerCrc32(str)
			}(input)

			go func(str string) {
				crc32md5dataString <- DataSignerCrc32(str)
			}(<-md5dataString)

			result := <-crc32dataString + "~" + <-crc32md5dataString
			out <- result
			wg.Done()
		}(dataString)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := sync.WaitGroup{}

	for data := range in {
		wg.Add(1)
		dataString := data.(string)

		go func(input string) {
			var crc32 [6]chan string

			for th := 0; th <= 5; th++ {
				crc32[th] = make(chan string)
				go func(i int, str string) {
					crc32[i] <- DataSignerCrc32(strconv.Itoa(i) + str)
				}(th, input)
			}
			result := ""
			for th := 0; th <= 5; th++ {
				result += <-crc32[th]
			}
			out <- result
			wg.Done()
		}(dataString)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var lines []string

	for data := range in {
		lines = append(lines, data.(string))
	}
	sort.Strings(lines)
	out <- strings.Join(lines, "_")
}

func main() {

}
