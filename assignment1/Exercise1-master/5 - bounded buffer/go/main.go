
// Use `go run main.go` to run your program
package main

import "fmt"
import "time"


func producer(/*TODO: parameters?*/ buf chan int){

    for i := 0; i < 10; i++ {
        time.Sleep(100 * time.Millisecond)
        fmt.Printf("[producer]: pushing %d\n", i)
        // TODO: push real value to buffer
        buf <- i
    }
    // close(buf) // the sender should close the channel when it is done
}

func consumer(/*TODO: parameters?*/ buf chan int){

    time.Sleep(1 * time.Second)
    for {
        i := <-buf // TODO: get real value from buffer
        fmt.Printf("[consumer]: %d\n", i)
        time.Sleep(50 * time.Millisecond)
    }
    
}


func main(){
    
    // TODO: make a bounded buffer
    buf := make(chan int, 5)
    
    go consumer(buf)
    go producer(buf)
    
    select {}
}