package main

import (
	"fmt"
	"sync"
)

// There is a pattern to our pipeline functions:
//
// stages close their outbound channels when all the send operations are done.
// stages keep receiving values from inbound channels until those channels are closed.

func main() {
	in := gen(2, 3)

	// distribute the work
	c1 := sq(in)
	c2 := sq(in)

	// Consume c1 and c2.
	for n := range merge(c1, c2) {
		fmt.Println(n)
	}
}

func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

func sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}
func merge(cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// start an output goroutine for each input channel in cs. output
	// copies values from c to out until c is closed, then calls wg.done
	output := func(c <-chan int) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// start a goroutine to close out once all the output goroutines are done.
	// It must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
