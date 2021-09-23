package main

import (
	"fmt"
	"sync"
)

// While this fixes the blocked goroutine in this program, this
// is bad code. The choice of buffer size of 1 here depends on
// knowing the number of values merge will receive and the
// number of values downstream stages will consume. This is
// fragile: if we pass an additional value to gen, or
// if the downstream stage reads any fewer values, we will
// again have blocked goroutines.

func main() {
	in := gen(2, 3)

	// distribute the work
	c1 := sq(in)
	c2 := sq(in)

	// Consume c1 and c2.
	out := merge(c1, c2)
	fmt.Println(<-out)
	return
}

func gen(nums ...int) <-chan int {
	out := make(chan int, len(nums))
	for _, n := range nums {
		out <- n
	}
	close(out)
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
	out := make(chan int, 1)

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
