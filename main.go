package main

import (
	"fmt"
	"sync"
)

func gen(start int, end int) <-chan int {
	out := make(chan int)
	go func() {
		for i := start; i <= end; i++ {
			out <- i
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

func even(in <-chan int) <-chan int {
	return filter(in, func(n int) bool { return (n%2 == 0) })
}

func odd(in <-chan int) <-chan int {
	return filter(in, func(n int) bool { return (n%2 != 0) })
}

func filter(in <-chan int, fn func(int) bool) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			if fn(n) {
				out <- n
			}
		}
		close(out)
	}()
	return out
}

func merge(ins ...<-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		var wg sync.WaitGroup
		wg.Add(len(ins))

		for _, in := range ins {
			go func(in <-chan int) {
				defer wg.Done()
				for n := range in {
					out <- n
				}
			}(in)
		}
		wg.Wait()
	}()
	return out
}

func broadcast(in <-chan int, out ...chan<- int) {
	go func() {
		var wg sync.WaitGroup
		for n := range in {
			wg.Add(len(out))
			for _, och := range out {
				go func(o chan<- int, n int) {
					o <- n
					wg.Done()
				}(och, n)
			}
		}
		wg.Wait()
		for _, och := range out {
			close(och)
		}
	}()
}

func distribute(in <-chan int, out ...chan<- int) {
	go func() {
		var wg sync.WaitGroup
		for n := range in {
			next := make(chan struct{})
			wg.Add(len(out))
			for _, och := range out {
				go func(o chan<- int, n int, next chan struct{}) {
					select {
					case <-next:
					case o <- n:
						next <- struct{}{}
					}
					wg.Done()
				}(och, n, next)
			}
		}
		wg.Wait()
		for _, och := range out {
			close(och)
		}
	}()
}

func main() {
	out1 := make(chan int)
	out2 := make(chan int)
	distribute(gen(1, 2), out1, out2)
	out := merge(sq(out1), even(out2))

	for n := range out {
		fmt.Println(n)
	}
}
