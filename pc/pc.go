package main

import (
	"fmt"
	"github.com/ChrisGora/semaphore"
	"math/rand"
	"sync"
	"time"
)

type buffer struct {
	b                 []int
	size, read, write int
}

func newBuffer(size int) buffer {
	return buffer{
		b:     make([]int, size),
		size:  size,
		read:  0,
		write: 0,
	}
}

func (buffer *buffer) get() int {
	x := buffer.b[buffer.read]
	fmt.Println("Get\t", x, "\t", buffer)
	buffer.read = (buffer.read + 1) % len(buffer.b)
	return x
}

func (buffer *buffer) put(x int) {
	buffer.b[buffer.write] = x
	fmt.Println("Put\t", x, "\t", buffer)
	buffer.write = (buffer.write + 1) % len(buffer.b)
}

func producer(buffer *buffer, spaceAvailable, workAvailable semaphore.Semaphore, mutex *sync.Mutex, start, delta int) {
	x := start
	for {

		spaceAvailable.Wait()
		mutex.Lock()
		buffer.put(x)
		mutex.Unlock()
		x = x + delta
		workAvailable.Post()

		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

	}
}

func consumer(buffer *buffer, spaceAvailable, workAvailable semaphore.Semaphore, mutex *sync.Mutex) {
	for {

		workAvailable.Wait()
		mutex.Lock()
		_ = buffer.get()
		mutex.Unlock()
		spaceAvailable.Post()
		time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
	}
}

func main() {
	buffer := newBuffer(5)
	mutex := sync.Mutex{}

	spaceAvailable := semaphore.Init(5, 5)
	workAvailable := semaphore.Init(5, 0)
	go producer(&buffer, spaceAvailable, workAvailable, &mutex, 1, 1)
	go producer(&buffer, spaceAvailable, workAvailable, &mutex, 1000, -1)

	consumer(&buffer, spaceAvailable, workAvailable, &mutex)
}
