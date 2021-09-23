package main

import (
	"fmt"
	"sync"
)

// Here are the guidelines for pipeline construction:
//
// stages close their outbound channels when all the send operations are done.
// stages keep receiving values from inbound channels until those channels are closed or the senders are unblocked.

func main() {
	// done channel shared by the whole pipeline and is a signal for
	// all pipes to exit.
	done := make(chan struct{})
	defer close(done)

	in := gen(2, 3)

	// distribute the sq work
	c1 := sq(done, in)
	c2 := sq(done, in)

	// consume data from output
	out := merge(done, c1, c2)
	fmt.Println(<-out) // 4 or 9

	// done closed by deferred call
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

func sq(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-done:
				return
			}
		}
	}()
	return out
}

func merge(done <-chan struct{}, cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// Start an output. The deferred wg.done means that every code path gets covered.
	output := func(c <-chan int) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// start a goroutine to close out once all the goroutines are done.
	// It must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
