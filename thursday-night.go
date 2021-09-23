package main

import (
	"fmt"
	"github.com/fatih/color"
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
	var wg sync.WaitGroup

	jobs := make(chan Target)
	results := make(chan Target)
	done := make(chan struct{})

	wg.Add(len(targets))

	// c := make(chan Target, len(targets))
	for i := 0; i < 8; i++ {
		go func() {
			applyHasFailed := false
			for {
				select {
				case t := <-jobs:
					if applyHasFailed {
						t.Skipped = true
						fmt.Printf("skipped later target: %s\n", t.Name)
					} else {
						if err := t.apply(); err != nil {
							t.Error = err
							applyHasFailed = true
							// this select is dumb but it means that the routine will
							// check if any others have failed first before trying to
							// close done twice and causing a panic.
							select {
							case <-done:
							default:
								close(done)
							}
						}
					}
					results <- t
				case <-done:
					applyHasFailed = true
				}
			}
		}()
	}

	var errorCount int
	var successCount int
	var skippedCount int
	go func() {
		for t := range results {
			if t.Error != nil {
				errorCount++
				color.New(color.FgRed).Fprintf(os.Stderr, "%s ❌ ERROR\n", t.Name)
			} else if t.Skipped {
				fmt.Fprintf(os.Stderr, ">>>>>>>>>>>>>>>>>>>>>>>>>>>> skipped : %s\n", t.Name)
				skippedCount++
			} else {
				color.New(color.FgGreen).Fprintf(os.Stderr, "%s ✅ SUCCESS\n", t.Name)
				color.Unset() // Don't forget to unset
				successCount++
			}
			wg.Done()
		}
	}()

	for _, target := range targets {
		jobs <- target
	}
	wg.Wait()
	return doMeAfter(0, 0, 0)
}

func (t *Target) apply() error {
	time.Sleep(time.Second)

	if t == nil {
		fmt.Fprintln(os.Stderr, "catch a apply target is nil")
		return nil
	}

	fmt.Printf("applying %s\n", t.Name)

	if t.Name == "f" || t.Name == "c" {
		// if t.Name == "f" {
		return fmt.Errorf("ZOMG BAD")
	}
	return nil
}
