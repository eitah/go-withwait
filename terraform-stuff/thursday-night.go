package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

var red *color.Color
var green *color.Color

func main() {
	if err := mainErr(); err != nil {
		red.Printf("Error: %s.\n", err)
	}
}

type Target struct {
	Name    string
	Error   error
	Skipped bool
}

var names []string
var targets []Target

func doMeAfter(successCount, errorCount, lentargets, skippedCount int, errors []string) error {

	if errorCount > 0 {
		fmt.Fprintln(os.Stderr)
		red.Fprintf(os.Stderr, "Failed targets (⌘ + F):\n")
		for _, name := range errors {
			fmt.Fprintf(os.Stderr, "• %s \n", name)
		}
		if skippedCount > 0 {
			fmt.Fprintf(os.Stderr, "%d targets skipped due to apply failure.\n", skippedCount)
		}
		return fmt.Errorf("%d/%d targets with errors. See above", errorCount, lentargets)
	}

	green.Fprintln(os.Stderr, "🌈 Applying Terraform succeeded 🌈")

	return nil
}

func init() {
	red = color.New(color.FgRed)
	green = color.New(color.FgGreen)
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
	for i := 0; i < 2; i++ {
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
					// case <-done:
					// 	applyHasFailed = true
				}
			}
		}()
	}

	errors := []string{}
	var errorCount int
	var successCount int
	var skippedCount int
	go func() {
		for t := range results {
			if t.Error != nil {
				errorCount++
				errors = append(errors, t.Name)
				red.Fprintf(os.Stderr, "%s ❌ ERROR\n", t.Name)
				color.Unset() // Don't forget to unset
				fmt.Println(t.Error)
			} else if t.Skipped {
				// fmt.Fprintf(os.Stderr, ">>>>>>>>>>>>>>>>>>>>>>>>>>>> skipped : %s\n", t.Name)
				skippedCount++
			} else {
				green.Fprintf(os.Stderr, "%s ✅ SUCCESS\n", t.Name)
				color.Unset() // Don't forget to unset
				fmt.Fprintln(os.Stderr, "good")
				successCount++
			}
			wg.Done()
		}
	}()

	for _, target := range targets {
		jobs <- target
	}
	wg.Wait()
	return doMeAfter(successCount, errorCount, len(targets), skippedCount, errors)
}

func (t *Target) apply() error {
	time.Sleep(time.Second)

	if t == nil {
		fmt.Fprintln(os.Stderr, "catch a apply target is nil")
		return nil
	}

	// fmt.Printf("applying %s\n", t.Name)

	if t.Name == "f" || t.Name == "c" {
		// if t.Name == "f" {
		return fmt.Errorf("ZOMG BAD")
	}

	return nil
}
