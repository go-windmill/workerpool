package workerpool

import (
	"log"
	"math/rand"
	"testing"
)

func testFuncWithPointer(i *int) {
	*i++
}

func Test_New_Returned(t *testing.T) {
	t.Parallel()
	// Create a new worker pool with name of testNew
	p := New("testNew")
	// close it done and wait
	p.CloseAndWait()
	// as we didnt start any workers counters should all
	// be 0
	if p.State.Workers.Started > 0 ||
		p.State.Workers.Stopped > 0 ||
		p.State.Tasks.Counter > 0 {

		log.Fatalln("No pool started or tasks added so counters should be 0, but weren't.", p.State)
	}
}

func Test_Start(t *testing.T) {
	t.Parallel()
	// create a worker pool
	p := New("testStart")
	size := rand.Intn(10-5) + 5
	// generate a least 5 workers
	p.Start(size)
	// close them all down
	p.CloseAndWait()
	// check counters
	if p.State.Workers.Started != size ||
		p.State.Workers.Stopped != size {
		log.Fatalf("Created a pool of (%v) workers but counters after closing did not match.\n%v\n", size, p.State.Workers)
	}
	if p.State.Tasks.Counter > 0 {
		log.Fatalf("Create a pool of (%v) workers without adding a task to the queue, so counter should be 0\n%v\n", size, p.State.Tasks)
	}
}

func Test_Multiple_Starts(t *testing.T) {
	t.Parallel()

	p := New("testMultiStarts")
	p.Start(5)
	p.Start(1)
	p.Start(1)

	p.CloseAndWait()

	if p.State.Workers.Started != 7 {
		log.Fatalf("Total number of workers started does not matched the expected value, expected %v, got %v\n", 7, p.State.Workers.Started)
	}
}

func Test_Multiple_StartStops(t *testing.T) {
	t.Parallel()

	p := New("testMultiStarts")
	p.Start(5)
	r1 := p.Running()
	p.Stop(1)
	r2 := p.Running()
	p.Stop(1)
	r3 := p.Running()
	p.Start(1)

	p.CloseAndWait()

	if r1 != 5 {
		log.Fatalf("Initial running count does not match expected pool size. Expected %v, got %v\n", 5, r1)
	}
	if r2 != 4 {
		log.Fatalf("Modified running count does not match expected pool size. Expected %v, got %v\n", 4, r2)
	}
	if r3 != 3 {
		log.Fatalf("Modified (again) running count does not match expected pool size. Expected %v, got %v\n", 3, r3)
	}
	if p.State.Workers.Started != 6 {
		log.Fatalf("End worker pool count does not match. Expected %v, got %v\n", 6, p.State.Workers.Started)
	}
}

func Test_Add_Simple(t *testing.T) {
	t.Parallel()
	p := New("testSimpleAdd")

	x := 0
	workers := 2
	tasks := 5

	p.Start(workers)
	for i := 0; i < tasks; i++ {
		p.Add(func() {
			testFuncWithPointer(&x)
		})
	}
	p.CloseAndWait()

	if x != tasks {
		log.Fatalf("Error running tasks, counter value incorrect. Expected %v, got %v\n", tasks, x)
	}
	if p.State.Workers.Started != workers {
		log.Fatalf("Error with worker pool size. Expected %v, got %v\n", workers, p.State.Workers.Started)
	}

}

func Test_Add_With_ExpandingPool(t *testing.T) {
	t.Parallel()
	p := New("textExpanding")

	x := 0
	workers := 2
	tasks := 50
	half := tasks / 2
	p.Start(workers)

	for i := 0; i < tasks; i++ {
		p.Add(func() {
			testFuncWithPointer(&x)
		})
		// should run twice
		if i%half == 0 {
			p.Start(1)
		}
	}

	p.CloseAndWait()

	if x != tasks {
		log.Fatalf("Error running tasks, counter value incorrect. Expected %v, got %v\n", tasks, x)
	}
	if p.State.Workers.Started != 4 {
		log.Fatalf("Pool has not expanded to the expected size. Expected %v, got %v\n", 4, p.State.Workers.Started)
	}
	if p.State.Tasks.Counter != 50 {
		log.Fatalf("Task counter does not match. Expected %v, got %v\n", tasks, p.State.Tasks.Counter)
	}

}
