package main

import (
	"fmt"
	"sync"
)

func main() {
	sum := 0
	var wg sync.WaitGroup
	var mutex = &sync.Mutex{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			mutex.Lock()
			sum = sum + 1
			wg.Done()
			mutex.Unlock()
		}()
	}

	wg.Wait()
	fmt.Println(sum)
}
