package workerpool

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Task is
type Task struct {
	ID       uuid.UUID
	F        func()
	Namspace string
}

// Workers is
type Workers struct {
	Started int
	Stopped int
}

// Tasks is
type Tasks struct {
	Counter int
}

// State is
type State struct {
	Workers Workers
	Tasks   Tasks
}

// WorkerPool is
type WorkerPool struct {
	Namespace string
	State     State
	WaitGroup *sync.WaitGroup
	Mutex     *sync.Mutex
	Exit      chan bool
	Queue     chan Task
}

// Start is
func (w *WorkerPool) Start(count int) {
	fmt.Println("Starting workers", count)
	for i := 0; i < count; i++ {
		w.WaitGroup.Add(1)
		w.State.Workers.Started++
		go worker(
			w.Namespace,
			i,
			w.Queue,
			w.Exit,
			&w.State,
			w.WaitGroup,
			w.Mutex)
	}
}

// Stop is
func (w *WorkerPool) Stop(count int) {
	fmt.Println("Stopping workers", count)

	// exit i number of items
	for i := 0; i < count; i++ {
		w.Exit <- true
	}
}

// Running is
func (w *WorkerPool) Running() int {
	return w.State.Workers.Started - w.State.Workers.Stopped
}

// Add is
func (w *WorkerPool) Add(task func()) {
	// task is real and there is a worker running
	if task != nil && w.State.Workers.Started > 0 {
		t := Task{ID: uuid.New(), F: task, Namspace: w.Namespace}
		fmt.Println("Adding a task", t)
		w.Queue <- t
	}
}

// CloseAndWait is
func (w *WorkerPool) CloseAndWait() {
	close(w.Queue)
	w.WaitGroup.Wait()
}

// New is
func New(namespace string) *WorkerPool {
	var wg sync.WaitGroup

	wp := WorkerPool{
		Namespace: namespace,
		WaitGroup: &wg,
		Mutex:     &sync.Mutex{},
		Exit:      make(chan bool),
		Queue:     make(chan Task, 1),
		State: State{
			Workers{Started: 0, Stopped: 0},
			Tasks{Counter: 0}}}

	return &wp
}

func worker(
	namespace string,
	index int,
	tasks <-chan Task,
	quit <-chan bool,
	state *State,
	wg *sync.WaitGroup,
	mu *sync.Mutex) {

	defer wg.Done()
	for {
		fmt.Printf("[%s:%d] worker waiting\n", namespace, index)
		select {
		case task, ok := <-tasks:
			if !ok {
				fmt.Printf("[%s:%d] no task, quiting\n", namespace, index)
				state.Workers.Stopped++
				return
			}
			fmt.Printf("[%s:%d][%s:%s] worker has recieved task\n", namespace, index, task.Namspace, task.ID)
			task.F()
			state.Tasks.Counter++
		case <-quit:
			fmt.Printf("[%s:%d] worker has quit\n", namespace, index)
			state.Workers.Stopped++
			return

		}
	}
}
