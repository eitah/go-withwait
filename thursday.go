package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

func main() {
	if err := mainErr(); err != nil {
		fmt.Printf("\nError: %s\n", err)
	}
}

type Target struct {
	Name    string
	Error   error
	Skipped bool
}

var names []string
var targets []Target

func doMeAfter(errorCount, lentargets, skippedCount int) error {

	if errorCount > 0 {
		fmt.Println("")
		return fmt.Errorf("%d/%d targets with errors. %d targets skipped. See above", errorCount, lentargets, skippedCount)

	}

	fmt.Fprintln(os.Stderr, "🌈 Applying Terraform succeeded 🌈")

	return nil
}

func init() {
	names = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}

	for _, name := range names {
		targets = append(targets, Target{Name: name})
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(targets), func(i, j int) { targets[i], targets[j] = targets[j], targets[i] })

	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()

	fmt.Println("Start---------")

}

func mainErr() error {
	// mainerr closes done when it returns in this case we're pretending
	// we want to kill everything immediately when done is called
	done := make(chan struct{})
	// defer close(done)

	jobs, _ := genJobs(done, targets)

	countConcurrent := 8
	c := make(chan Target, len(targets))
	for i := 0; i < countConcurrent; i++ {
		go func() {
			tfRunner(done, jobs, c)
		}()
	}

	// in this implementation errc is buffered so no need for select stmt
	// note with this statement by itself it blocks the main thread unless
	// 1) i do the below in a goroutien
	// 2) i close the err channel
	// if err := <-errc; err != nil {
	// 	return err
	// }
	for res := range c {
		fmt.Printf("res:%s\n", res.Name)
	}

	return doMeAfter(0, 0, 0)
}

func tfRunner(done <-chan struct{}, jobs <-chan Target, results chan<- Target) {
	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		fmt.Print("After add\n")
		wg.Done()
		select {
		case t := <-jobs:
			fmt.Printf("wg finished: %s\n", t.Name)

			results <- t
		case <-done:
			fmt.Println("walk canceled")
		}
		wg.Done()
	}()

	go func() {
		close(results)
	}()

	// no select needed here since errc is buffered
	// errc <- err
}

// genJobs turns an array of targets into a channel of targets
// for prod I could replace with a for loop like it was but meh.
// this is a learning thing
func genJobs(done <-chan struct{}, targets []Target) (<-chan Target, <-chan error) {
	jobs := make(chan Target)
	errc := make(chan error, 1)

	go func() {
		// guarantee jobs is always closed
		defer close(jobs)

		for _, target := range targets {
			select {
			case jobs <- target:
			case <-done:
				errc <- errors.New("terraform canceled")
			}
		}
	}()

	return jobs, errc
}

func (t *Target) apply() error {
	time.Sleep(time.Second)

	if t == nil {
		fmt.Fprintln(os.Stderr, "catch a apply target is nil")
		return nil
	}

	fmt.Printf("applying %s\n", t.Name)

	// if t.Name == "f" || t.Name == "c" {
	if t.Name == "f" {
		return fmt.Errorf("ZOMG BAD")
	}
	return nil
}
