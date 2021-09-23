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
var targets []*Target

func init() {
	names = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}

	for _, name := range names {
		targets = append(targets, &Target{Name: name})
	}
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()

	fmt.Println("Start---------")
}

func mainErr() error {
	var wg sync.WaitGroup

	jobs := make(chan *Target)
	// jobs := make(chan *Target, len(targets))
	results := make(chan *Target)
	// results := make(chan *Target, len(targets))
	done := make(chan struct{})
	wg.Add(len(targets))

	// skipMutex := &sync.Mutex{}

	// go func() {
	// 	select {
	// 	case <-done:
	// 		skipMutex.Lock()
	//
	// 		// fmt.Println("foo")
	// 		if shouldDoApply {
	// 			// close(jobs)
	// 			shouldDoApply = false
	// 			skipMutex.Unlock()
	// 		}
	// 	}
	// }()

	for i := 0; i < 8; i++ {
		go func() {
			shouldDoApply := true
			for {
				select {
				case t := <-jobs:
					if shouldDoApply {
						if err := t.DoApply(); err != nil {
							t.Error = err
							if shouldDoApply {
								close(done)
							}
						}
						results <- t
					} else {
						// i was getting a nil panic till I added this not sure why jobs hits
						if t == nil {
							fmt.Fprintln(os.Stderr, "catch a nil target")
							return
						}
						fmt.Fprintf(os.Stderr, ">>>>>>>>>>>>>>>>>>>>>>>>>>>> skipped : %s\n", t.Name)
						t.Skipped = true
						results <- t
					}
				case <-done:
					// fmt.Println("foo")
					if shouldDoApply {
						shouldDoApply = false
					}

				}
			}

		}()
	}

	var errors []error
	// these have to do with printing the resuklts
	var errorCount int
	var successCount int
	var skippedCount int
	go func() {
		for t := range results {
			if t.Error != nil {
				printme("%s ❌ FAIL\n", t.Name, color.FgRed)
				// color.Unset() // Don't forget to unset
				errorCount++
				errors = append(errors, t.Error)
			} else if t.Skipped {
				fmt.Fprintf(os.Stderr, ">>>>>>>>>>>>>>>>>>>>>>>>>>>> skipped : %s\n", t.Name)
				skippedCount++
			} else {
				printme("%s ✅ SUCCESS\n", t.Name, color.FgGreen)
				// color.New(color.FgGreen).Fprintf(os.Stderr, "%s ✅ SUCCESS\n", t.Name)
				// color.Unset() // Don't forget to unset
				successCount++
			}

			wg.Done()
		}
	}()

	// has to do with supplying jobs
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(targets), func(i, j int) { targets[i], targets[j] = targets[j], targets[i] })
	for _, target := range targets {
		jobs <- target
	}
	close(jobs)

	wg.Wait()
	close(results)

	if errorCount > 0 {
		fmt.Println("")
		return fmt.Errorf("%d/%d targets with errors. %d targets skipped. See above", errorCount, len(targets), skippedCount)

	}

	fmt.Fprintln(os.Stderr, "🌈 Applying Terraform succeeded 🌈")

	return nil
}

func (t *Target) DoApply() error {
	if err := t.apply(); err != nil {
		// done <- t
		return err
	}

	return nil
}

func (t *Target) plan() error {
	// fmt.Printf("planning %s\n", t.Name)
	return nil
}

func (t *Target) apply() error {
	time.Sleep(time.Second)
	// fmt.Printf("applying %s\n", t.Name)

	// if t.Name == "f" || t.Name == "c" {
	if t == nil {
		fmt.Fprintln(os.Stderr, "catch a apply target is nil")
		return nil
	}
	if t.Name == "f" {
		return fmt.Errorf("ZOMG BAD")
	}
	return nil
}

func printme(format string, s string, c color.Attribute) {
	// if c == nil {
	// 	fmt.Fprintf(os.Stderr, format, s)
	// 	return
	// }
	color.New(c).Fprintf(os.Stderr, format, s)
	color.Unset()
}
