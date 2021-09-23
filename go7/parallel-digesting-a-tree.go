package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

func main() {
	// calculate a md5 sum of all files under a specified directory
	// then print the results sorted by path
	if len(os.Args) < 2 {
		fmt.Println("Error. Please pass a path to search as the first argument")
		return
	}

	m, err := md5all(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	var paths []string
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		fmt.Printf("%x %s\n", m[path], path)
	}
}

type result struct {
	path string
	sum  [md5.Size]byte
	err  error
}

// sumFiles: walk the tree and sum all the files. First stage
func sumFiles(done <-chan struct{}, root string) (<-chan result, <-chan error) {
	// for each regular file start a goroutine that sums the file and sends
	// the result on c. Send the result of the walk on errc
	c := make(chan result)
	errc := make(chan error, 1)
	go func() {
		var wg sync.WaitGroup
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			wg.Add(1)
			go func() {
				data, err := os.ReadFile(path)
				select {
				case c <- result{path, md5.Sum(data), err}:
				case <-done:
				}
				wg.Done()
			}()

			// Abort the walk if done is closed.
			select {
			case <-done:
				return errors.New("Walk Cancelled")
			default:
				return nil
			}
		})

		// walk has returned so all wait group adds are done
		go func() {
			wg.Wait()
			close(c)
		}()

		// no select needed here since errc is buffered
		errc <- err
	}()
	return c, errc
}

func md5all(root string) (map[string][md5.Size]byte, error) {
	// md5 all closes done when it returns, it may do so before getting
	// all the values from c and errc
	done := make(chan struct{})
	defer close(done)

	c, errc := sumFiles(done, root)

	m := make(map[string][md5.Size]byte)
	for r := range c {
		if r.err != nil {
			return nil, r.err
		}
		m[r.path] = r.sum
	}
	if err := <-errc; err != nil {
		return nil, err
	}
	return m, nil
}
