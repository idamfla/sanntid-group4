// Use `go run foo2.go` to run your program

package main

import (
    . "fmt"
    "runtime"
)

type Increment struct{}
type Decrement struct{}
type GetValue struct{
	reply chan int
}

func runServer(cmds chan any) {
	counter := 0
	for {
		select{
		case msg := <-cmds:
			switch m := msg.(type) {
			case Increment:
				counter++
			case Decrement:
				counter--
			case GetValue:
				m.reply <-counter
			}
		}
	}
}

func main() {
    runtime.GOMAXPROCS(2)

	cmds := make(chan any)
    
	go runServer(cmds)

	go func() {
		for j := 0; j < 1_000_000; j++ {
			cmds <- Increment{}
		}
	}()

	go func() {
		for j := 0; j < 1_000_000; j++ {
			cmds <- Decrement{}
		}
	}()

	r := make(chan int)
	cmds <- GetValue{reply: r}
    Println("The magic number is:", <-r)
}