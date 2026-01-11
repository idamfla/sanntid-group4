## Additional Questions

### Task 2

> We do not get 0 when we add and subtract the same amount of because of race condition and a lack of read/write locking the variable. This makes so that two functions might want to change a variable at once, unaware that somebody else is also changing the same variable. They will then not register the other functions changes and might end up overriding it.

What does `GOMAXPROCS` do? What happens if you set it to 1?

> `GOMAXPROCS` sets a max amount of CPU cores that the go routine can run at once. Multiple cores will make it possible to run the code in parallel.

What happens when we set `GOMAXPROCS` to 1?
> When setting `GOMAXPROCS` to 1, go will behave like c with mutex_lock.

### Task 3
POSIX has both mutexes (`pthread_mutex_t`) and semaphores (`sem_t`). Which one should you use? Add a comment explaining why your choice is the correct one.
> `pthread_t` is about protecting memory, a single thing. `sem_t` is about limiting the number of threads that can access a resources at the same time. For this task `pthread_t` is the most fitting.

### Task 4

How to prevent the infinite loop?
> If we do not want the infinite loop to crash the program, the sender of the data should close the channel when it is done. In this case, that is the `producer`.
