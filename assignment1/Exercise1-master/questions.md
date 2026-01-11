Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> The difference between *concurrency* and *parallelism* is that concurrency runs on the same CPU core, while the parallelism runs on multiple. Concurrency gives the illusion of parallelism by switching between multiple tasks very quickly.

What is the difference between a *race condition* and a *data race*? 
> *race condition* is when the timing of the execution affect the behavior of the program. This term are commenly used when there are many possible exection orders which all lead to different outcomes.
> *data race* is when two threads try to access the same memory at the same time, and at least one of these are writing to the memory. This happens when there is no propper lock or guard around the accessed memory that prevent others from accessing it while it is being written to.
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> A schedualer decides the priority of different tasks, and it can preempt running tasks in favor of one with higher priority. It ensures that deadlines are met. Priority can be decided by deadline, execution time or other.


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> We would use multiple threads to simulate multiple CPU cores, and it make it possible to run multiple tasks "parallel".

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> *Fiber* (*green threads*)/*coroutines* are lightweight threads managed in user space. This means that the user decides when the switching between fibers are suppose to occur instead of the operating system controlling it.
> We would rather use these if we want to be able to control when the program switch between different threads.

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> This depends on the task at hand. For simple tasks there might be overkill to implement concurrenty, but for more complex ones, that need to give a real-time-feel, concurrency mgiht be a necessity. For the simple task, it will be a bad trade-off to implement all the complexity of threads if it is possiple to achive the same result without it.

What do you think is best - *shared variables* or *message passing*?
> We think that for simple operations *message passing* is the better choice because it prevents race conditions. However, for more complex systems with a lot of functionality and many variables, it might cause the code to become unorgianized. Also if only one function is allowed to change a variable, you will end up with a lot of logic inside that one function which might make the whole process a bit slow.
