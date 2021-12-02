package main

import (
	"flag"
	"runtime"
	"sync"
)

const (
	thinBorder  = 0.05
	thickBorder = 0.25
)

func main() {
	flag.Parse()

	wg := &sync.WaitGroup{}
	jobs := make(chan string)
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go worker(jobs, wg)
	}

	for _, fname := range flag.Args() {
		jobs <- fname
	}
	wg.Wait()
}

func worker(jobs chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		doOneFrame(job)
	}
}

func doOneFrame(fname string) {

}
