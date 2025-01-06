package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	chIn := make(chan interface{})
	wg := new(sync.WaitGroup)

	for _, job := range jobs {
		chOut := make(chan interface{})
		wg.Add(1)
		go func(job func(in, out chan interface{}), in, out chan interface{}) {
			defer close(out)
			defer wg.Done()
			job(in, out)
		}(job, chIn, chOut)
		chIn = chOut
	}
	wg.Wait()
}

func Crc32Go(data string, ch chan<- interface{}) {
	ch <- DataSignerCrc32(data)
}

func SingleHashGo(data interface{}, out chan interface{}, mutex *sync.Mutex, waiter *sync.WaitGroup) {
	defer waiter.Done()
	ch1, ch2 := make(chan interface{}), make(chan interface{})
	dataStr := fmt.Sprint(data)

	go Crc32Go(dataStr, ch1)

	mutex.Lock()
	md5 := DataSignerMd5(dataStr)
	mutex.Unlock()

	go Crc32Go(md5, ch2)

	valLeft := (<-ch1).(string)
	valRight := (<-ch2).(string)

	out <- valLeft + "~" + valRight
}

func SingleHash(in, out chan interface{}) {
	wg := new(sync.WaitGroup)
	mu := new(sync.Mutex)

	for input := range in {
		wg.Add(1)
		go SingleHashGo(input, out, mu, wg)
	}
	wg.Wait()
}

func MultiHashGo(data interface{}, out chan interface{}, waiter *sync.WaitGroup) {
	defer waiter.Done()

	var res [6]string
	wgLocal := new(sync.WaitGroup)
	wgLocal.Add(6)
	dataStr := fmt.Sprint(data)
	for i := range 6 {
		go func(idx int) {
			res[idx] = DataSignerCrc32(fmt.Sprint(idx) + dataStr)
			wgLocal.Done()
		}(i)
	}
	wgLocal.Wait()

	var resStr string
	for _, v := range res {
		resStr += v
	}

	out <- resStr
}

func MultiHash(in, out chan interface{}) {
	wg := new(sync.WaitGroup)

	for input := range in {
		wg.Add(1)
		go MultiHashGo(input, out, wg)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	arr := []string{}
	for input := range in {
		arr = append(arr, input.(string))
	}
	sort.Strings(arr)
	result := strings.Join(arr, "_")
	out <- result
}
