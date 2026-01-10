// Compile with `gcc foo.c -Wall -std=gnu99 -lpthread`, or use the makefile
// The executable will be named `foo` if you use the makefile, or `a.out` if you use gcc directly

#include <pthread.h>
#include <stdio.h>

int i = 0;
static pthread_mutex_t i_mtx;

// Note the return type: void*
void *incrementingThreadFunction()
{
    // TODO: increment i 1_000_000 times
    for (size_t j = 0; j < 1000000; j++)
    {
        pthread_mutex_lock(&i_mtx);
        i += 1;
        pthread_mutex_unlock(&i_mtx);
    }

    return NULL;
}

void *decrementingThreadFunction()
{
    // TODO: decrement i 1_000_000 times
    for (size_t j = 0; j < 1000000; j++)
    {
        pthread_mutex_lock(&i_mtx);
        i -= 1;
        pthread_mutex_unlock(&i_mtx);
    }
    return NULL;
}

int main()
{
    // TODO:
    // start the two functions as their own threads using `pthread_create`
    // Hint: search the web! Maybe try "pthread_create example"?

    // TODO:
    // wait for the two threads to be done before printing the final result
    // Hint: Use `pthread_join`
    pthread_t thread1, thread2;
    pthread_create(&thread1, NULL, incrementingThreadFunction, &i);
    pthread_create(&thread2, NULL, decrementingThreadFunction, &i);

    pthread_join(thread1, NULL);
    pthread_join(thread2, NULL);
    printf("The magic number is: %d\n", i);
    return 0;
}
