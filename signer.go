package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код
func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 100)

	for _, job := range jobs {
		wg.Add(1)
		out := make(chan interface{}, 100)
		go startWorker(in, out, wg, job)
		in = out
	}
	wg.Wait()
}

func startWorker(in, out  chan interface{}, wg *sync.WaitGroup, job job) {
	defer wg.Done()
	job(in, out)
	close(out)
}


func SingleHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for i := range in {
		wg.Add(1)
		go func(in interface{}, mu *sync.Mutex) {
			defer wg.Done()
			data := strconv.Itoa(in.(int))

			mu.Lock()
			md5 := DataSignerMd5(data)
			mu.Unlock()

			dataCh1 := make(chan string)
			dataCh2 := make(chan string)

			go func(ch1 chan string, data string) {
				ch1 <- DataSignerCrc32(data)
			}(dataCh1, data)

			go func(ch2 chan string, data string) {
				ch2 <- DataSignerCrc32(data)
			}(dataCh2, md5)

			crc32 := <-dataCh1
			crc32md5 := <-dataCh2

			out <- crc32 + "~" + crc32md5
		}(i, mu)
	}
	wg.Wait()
}


func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for input := range in {
		wg.Add(1)
		go func(in interface{}) {
			defer wg.Done()

			muCRC32 := &sync.Mutex{}
			wgCRC32 := &sync.WaitGroup{}

			res := make([]string, 6)

			for i := 0; i < 6; i++ {
				wgCRC32.Add(1)

				go func(wgCRC32 *sync.WaitGroup, muCRC32 *sync.Mutex, i int, res []string) {
					defer wgCRC32.Done()
					data := strconv.Itoa(i) + in.(string)
					crc32Data := DataSignerCrc32(data)

					muCRC32.Lock()
					res[i] = crc32Data
					muCRC32.Unlock()
				}(wgCRC32, muCRC32, i, res)
			}
			wgCRC32.Wait()

			concatRes := strings.Join(res, "")
			out <- concatRes
		}(input)
	}
	wg.Wait()
}


func CombineResults(in, out chan interface{}) {
	var data []string
	for i := range in {
		data = append(data, i.(string))
	}
	sort.Strings(data)
	combinedRes := strings.Join(data, "_")
	out <- combinedRes
}