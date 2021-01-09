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

// WorkerPool is
type WorkerPool struct {
	Namespace  string
	MaxWorkers int
	Exit       chan bool

	WaitGroup *sync.WaitGroup
	Mutex     *sync.Mutex

	Tasks chan Task
}

// Start is
func (w *WorkerPool) Start(count int) {
	for i := 0; i < count; i++ {
		w.WaitGroup.Add(1)
		fmt.Println("New worker", i)
		go worker(
			w.Namespace,
			i,
			w.Tasks,
			w.Exit,
			w.WaitGroup,
			w.Mutex)
	}
}

// Stop is
func (w *WorkerPool) Stop(count int) {
	for i := 0; i < count; i++ {
		<-w.Exit
	}
}

// Add is
func (w *WorkerPool) Add(task func()) {

	if task != nil {
		t := Task{ID: uuid.New(), F: task, Namspace: w.Namespace}
		fmt.Println("Adding a task", t)
		w.Tasks <- t
	}
}

// CloseAndWait is
func (w *WorkerPool) CloseAndWait() {
	close(w.Tasks)
	w.WaitGroup.Wait()
}

// New is
func New(namespace string, count int) *WorkerPool {
	var wg sync.WaitGroup

	wp := WorkerPool{
		Namespace:  namespace,
		MaxWorkers: count,
		Exit:       make(chan bool),
		WaitGroup:  &wg,
		Mutex:      &sync.Mutex{},
		Tasks:      make(chan Task, 1)}

	wp.Start(count)

	return &wp
}

func worker(
	namespace string,
	index int,
	tasks <-chan Task,
	quit <-chan bool,
	wg *sync.WaitGroup,
	mu *sync.Mutex) {

	defer wg.Done()
	for {
		fmt.Printf("[%s:%d] worker waiting\n", namespace, index)
		select {
		case task, ok := <-tasks:
			if !ok {
				fmt.Printf("[%s:%d] worker failed to retrieve task: \n%v\n", namespace, index, ok)
				return
			}
			fmt.Printf("[%s:%d][%s:%s] worker has recieved task\n", namespace, index, task.Namspace, task.ID)
			task.F()
		case <-quit:
			fmt.Printf("[%s:%d] worker has quit\n", namespace, index)
			return

		}
	}
}
