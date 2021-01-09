package workerpool

import (
	"sync"

	l "github.com/go-windmill/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Task is a struct providing a small wrapper around
// the func added to the Queue
// Provides an ID for tracking and namespace for
// separation
type Task struct {
	ID       uuid.UUID
	F        func()
	Namspace string
}

// Workers provides some state data for the pool
type Workers struct {
	Started int
	Stopped int
}

// Tasks provides a counter to track amount processed
type Tasks struct {
	Counter int
}

// State provides Worker and Task tracking data
type State struct {
	Workers Workers
	Tasks   Tasks
}

// WorkerPool has a namespace (for easier logging),
// State data for external tracking and access to
// WaitGroup and Mutex
// Queue chan is where tasks are added onto for
// the pool to process
// Exit handles exit signals to stop a worker
type WorkerPool struct {
	Namespace string
	State     State
	WaitGroup *sync.WaitGroup
	Mutex     *sync.Mutex
	Exit      chan bool
	Queue     chan Task
}

// Start creates a pool of `count` workers and updates state
// The `worker` gets state data passed in as well as access
// to the chan
func (w *WorkerPool) Start(count int) {
	l.Log.Debug("Starting workers", zap.Int("count", count))
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

// Stop uses the Exit chan to cause `count` workers to return
// and stop
func (w *WorkerPool) Stop(count int) {
	l.Log.Debug("Stopping workers", zap.Int("count", count))

	// exit i number of items
	for i := 0; i < count; i++ {
		w.Exit <- true
	}
}

// Running returns count of active workers
func (w *WorkerPool) Running() int {
	return w.State.Workers.Started - w.State.Workers.Stopped
}

// Add pushes the func() passed on to a chan for the worker pool
// to process
// Applies some meta data to the task (id for tracking and namespace)
func (w *WorkerPool) Add(task func()) {
	// task is real and there is a worker running
	if task != nil && w.State.Workers.Started > 0 {
		t := Task{ID: uuid.New(), F: task, Namspace: w.Namespace}
		l.Log.Debug("Adding task",
			zap.String("uuid", t.ID.String()),
			zap.String("namespace", t.Namspace))
		w.Queue <- t
	}
}

// CloseAndWait closes the queue and waits for the wait WaitGroup
// to finish
func (w *WorkerPool) CloseAndWait() {
	l.Log.Debug("Closing and waiting")
	close(w.Queue)
	w.WaitGroup.Wait()
}

// New generates new struct but does not start a worker pool
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
		l.Log.Debug("Worker starting",
			zap.String("namespace", namespace),
			zap.Int("worker", index))

		select {
		case task, ok := <-tasks:
			if !ok {
				l.Log.Warn("Worker exiting - no task",
					zap.String("namespace", namespace),
					zap.Int("worker", index))
				state.Workers.Stopped++
				return
			}
			l.Log.Info("Worker recieved task",
				zap.String("namespace", namespace),
				zap.Int("worker", index),
				zap.String("task", task.ID.String()))
			task.F()
			state.Tasks.Counter++
		case <-quit:
			l.Log.Warn("Worker quiting from notification",
				zap.String("namespace", namespace),
				zap.Int("worker", index))
			state.Workers.Stopped++
			return

		}
	}
}
