package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

// For thread-safe get ID for tasks
type SafeCounter struct {
	mu sync.Mutex
	id int
}

func (c *SafeCounter) GetID() int {
	result_id := 0
	c.mu.Lock()
	c.id++
	result_id = c.id
	c.mu.Unlock()
	return result_id
}

// Ttype represents a task with its id, creation time, execution time, and status.
type Ttype struct {
	id         int
	cT         time.Time // Creation Time
	fT         time.Time // Execution Time
	taskStatus int
}

const (
	STATUS_ERROR_TASK       = iota
	STATUS_READY_TO_PROCESS = iota
	STATUS_DONE             = iota
	STATUS_UNDONE           = iota
)

func main() {
	// Init safe counter for task ids
	safeCounter := SafeCounter{}

	// Set the total execution time to 3 seconds.
	end_time := time.Now().Add(3 * time.Second)

	// Create a channel to receive tasks.
	taskChan := make(chan Ttype, 10)
	// Channels for collecting results
	doneTasks := make(chan Ttype, 10)
	undoneTasks := make(chan error, 10)

	// Final result of the execution
	results := make(map[int]Ttype)
	errors := []error{}

	// Use a wait group to wait for all goroutines to finish.
	var wg sync.WaitGroup
	var worker_wg sync.WaitGroup

	task_creturer := func() {
		defer wg.Done()

		// Create tasks for the total execution time.
		for time.Now().Before(end_time) {
			status := STATUS_READY_TO_PROCESS // ready to process

			// Simulation of creating a task with an error
			if rand.Intn(10) == 0 { // 1/10
				status = STATUS_ERROR_TASK // Status indicating that an error has occurred
			}

			// Create a new task.
			task := Ttype{
				id:         safeCounter.GetID(),
				cT:         time.Now(),
				fT:         time.Now(),
				taskStatus: status,
			}

			// Send the task to the channel.
			taskChan <- task
		}

		// Close the channel when all tasks have been sent.
		close(taskChan)
	}

	task_sorter := func(t Ttype) {
		switch t.taskStatus {
		case STATUS_DONE:
			doneTasks <- t
			break
		case STATUS_UNDONE:
			undoneTasks <- fmt.Errorf("Task id %d time %s, error %s", t.id, t.cT, "Task completion failed")
			break
		case STATUS_ERROR_TASK:
			undoneTasks <- fmt.Errorf("Task id %d time %s, error %s", t.id, t.cT, "The task was with an error")
			break
		default:
			break
		}
	}

	task_worker := func() {
		defer worker_wg.Done()

		// Receive tasks from the channel until it is closed.
		for task := range taskChan {
			task.fT = time.Now()

			if task.taskStatus == STATUS_READY_TO_PROCESS { // task ready to process
				if rand.Float64() > 0.1 {
					task.taskStatus = STATUS_DONE // Success
				} else {
					task.taskStatus = STATUS_UNDONE // Something went wrong.
				}
			}

			task_sorter(task)
		}

	}

	// Run main logic

	// Start a goroutine to create tasks and send them to the channel.
	wg.Add(1)
	go task_creturer()

	// Start a goroutine to receive tasks from the channel and process them.
	go func() {
		for i := 0; i < runtime.NumCPU(); i++ {
			worker_wg.Add(1)
			go task_worker()
		}
		worker_wg.Wait()
		// Сlose channels as soon as the workers complete their work.
		close(doneTasks)
		close(undoneTasks)
	}()

	// Collect data from channels and write them to result
	{
		wg.Add(2)
		go func() {
			defer wg.Done()
			for task := range doneTasks {
				results[task.id] = task
			}
		}()
		go func() {
			defer wg.Done()

			for task_err := range undoneTasks {
				errors = append(errors, task_err)
			}
		}()
	}

	// Wait for all goroutines to finish.
	wg.Wait()

	// Print successful tasks and errors
	fmt.Println("Successful tasks:")
	println(len(results), "tasks")
	fmt.Println("Errors:")
	println(len(errors), "tasks")
}
