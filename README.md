# Workerpool

Concurrency enabling tools creating a queue and pool of processing workers agents.

## Installation

Install package with 

```
go get github.com/go-windmill/workerpool
```

## Details

A `WorkerPool` provides methods to start or stop `x` number of workers and way to add tasks onto the processing queue for concurrent handling.

Creating a new `WorkerPool` does not immediately generate running workers to keep a lower memory usage and more flexibilty. Instead, you need to call the `Start(int)` method afterwards to create the workers.

To utilise the concurrent pool you need to add a `func()` to via the `Add(func())` method.

This is then pushed on to the `chan` (being used as a queue) and picked up by a worker.

The size of the worker pool can be adjusted with `Start(int)` and `Stop(int)` methods to add or remove.

`Stop(int)` pushes to a `chan` used to signal `Exit` status for a worker. The worker listens to that `chan` and returns when detected.

Starting, Stopping and Adding updates the `WorkerPool` state information which tracks total number of workers started, stopped and tasks run. 

There is also a `Running()` method exposed to provide a counter of active worker pools

## Examples

Examples of using the `WorkerPool`

### Basic usage

Create two workers and use the queue to increament a counter ten times splitting the task between the pool.

```
package main

import (
    "fmt"

    "github.com/go-windmill/workerpool"
)

func main(){
    pool := workerpool.New("ns")
    // create two workers
    pool.Start(2)
    // counter
    counter := 0

    for i := 0; i < 10; i ++ {
        pool.Add(func(){
            counter++
        })
    }

    // wait for completion
    pool.CloseAndWait()
    // output the end value of the counter
    fmt.Println("Counter = ", counter)
}
```

### State data

Inspect state to ensure results align with expected.

```
package main

import (
    "fmt"

    "github.com/go-windmill/workerpool"
)

func main(){
    pool := workerpool.New("ns")
    // create two workers
    pool.Start(2)
    // counter
    counter := 0

    for i := 0; i < 10; i ++ {
        pool.Add(func(){
            counter++
        })
    }

    // wait for completion
    pool.CloseAndWait()
    // output the end value of the counter
    fmt.Println("Counter = ", counter)

    if pool.State.Tasks.Counter != 10 {
        log.Fatalln("Task counter does not match")
    }
    if pool.State.Workers.Started != 2 {
        log.Fatalln("Workers started does not match")
    }
}
```