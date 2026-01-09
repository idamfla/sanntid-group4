// Use `go run foo.go` to run your program

package main

import (
    . "fmt"
    "runtime"
    "time"
)

var i = 0

func incrementing() {
    //TODO: increment i 1000000 times
    for j := 0; j < 1000000; j++ {
        i++
    }
}

func decrementing() {
    //TODO: decrement i 1000000 times
    for j := 0; j < 1000000; j++ {
        i--
    }
}

func main() {
    // What does GOMAXPROCS do? What happens if you set it to 1?
    runtime.GOMAXPROCS(1)
    /* maximum number of cpu cores go can use at the same time.
    if it sets to 1 then you get it to be concurrent instead of parallel*/    
	
    // TODO: Spawn both functions as goroutines
    go incrementing()
    go decrementing()
	
    // We have no direct way to wait for the completion of a goroutine (without additional sc := make(chan int)  // create a channel for integersynchronization of some sort)
    // We will do it properly with channels soon. For now: Sleep.
    time.Sleep(500*time.Millisecond)
    Println("The magic number is:", i)
}
