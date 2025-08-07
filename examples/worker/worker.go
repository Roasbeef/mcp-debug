package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type WorkerPool struct {
	workers   int
	taskQueue chan Task
	wg        sync.WaitGroup
	stats     *Stats
}

type Task struct {
	ID       int
	Duration time.Duration
	Type     string
}

type Stats struct {
	mu             sync.RWMutex
	tasksCompleted int
	tasksStarted   int
	errors         int
	startTime      time.Time
}

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers:   workers,
		taskQueue: make(chan Task, 100),
		stats: &Stats{
			startTime: time.Now(),
		},
	}
}

func (wp *WorkerPool) Start() {
	log.Printf("Starting worker pool with %d workers", wp.workers)
	
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	go wp.taskGenerator()

	go wp.statsReporter()

	go wp.httpServer()
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	log.Printf("Worker %d started", id)

	for task := range wp.taskQueue {
		wp.processTask(id, task)
	}

	log.Printf("Worker %d stopped", id)
}

func (wp *WorkerPool) processTask(workerID int, task Task) {
	wp.stats.mu.Lock()
	wp.stats.tasksStarted++
	wp.stats.mu.Unlock()

	log.Printf("Worker %d processing task %d (type: %s, duration: %v)", 
		workerID, task.ID, task.Type, task.Duration)

	time.Sleep(task.Duration)

	if rand.Float32() < 0.1 {
		wp.stats.mu.Lock()
		wp.stats.errors++
		wp.stats.mu.Unlock()
		log.Printf("Worker %d: Task %d failed!", workerID, task.ID)
		return
	}

	result := wp.computeResult(task)

	wp.stats.mu.Lock()
	wp.stats.tasksCompleted++
	wp.stats.mu.Unlock()

	log.Printf("Worker %d completed task %d with result: %d", workerID, task.ID, result)
}

func (wp *WorkerPool) computeResult(task Task) int {
	sum := 0
	iterations := 1000000

	for i := 0; i < iterations; i++ {
		sum += i % 100
		if i%100000 == 0 {
			sum = sum / 2
		}
	}

	if task.Type == "complex" {
		return wp.complexCalculation(sum)
	}

	return sum
}

func (wp *WorkerPool) complexCalculation(input int) int {
	fibonacci := func(n int) int {
		if n <= 1 {
			return n
		}
		a, b := 0, 1
		for i := 2; i <= n; i++ {
			a, b = b, a+b
		}
		return b
	}

	primeCheck := func(n int) bool {
		if n <= 1 {
			return false
		}
		for i := 2; i*i <= n; i++ {
			if n%i == 0 {
				return false
			}
		}
		return true
	}

	result := fibonacci(input % 20)
	
	if primeCheck(result) {
		result *= 2
	}

	return result
}

func (wp *WorkerPool) taskGenerator() {
	taskID := 0
	taskTypes := []string{"simple", "complex", "quick"}

	for {
		taskID++
		taskType := taskTypes[rand.Intn(len(taskTypes))]
		
		var duration time.Duration
		switch taskType {
		case "quick":
			duration = time.Millisecond * time.Duration(100+rand.Intn(400))
		case "simple":
			duration = time.Second * time.Duration(1+rand.Intn(3))
		case "complex":
			duration = time.Second * time.Duration(2+rand.Intn(5))
		}

		task := Task{
			ID:       taskID,
			Duration: duration,
			Type:     taskType,
		}

		select {
		case wp.taskQueue <- task:
			log.Printf("Generated task %d (type: %s)", taskID, taskType)
		default:
			log.Printf("Task queue full, dropping task %d", taskID)
		}

		time.Sleep(time.Second * time.Duration(1+rand.Intn(3)))
	}
}

func (wp *WorkerPool) statsReporter() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		wp.stats.mu.RLock()
		uptime := time.Since(wp.stats.startTime)
		completed := wp.stats.tasksCompleted
		started := wp.stats.tasksStarted
		errors := wp.stats.errors
		wp.stats.mu.RUnlock()

		pending := started - completed - errors
		successRate := float64(completed) / float64(started) * 100

		log.Printf("=== STATS ===")
		log.Printf("Uptime: %v", uptime)
		log.Printf("Tasks - Started: %d, Completed: %d, Failed: %d, Pending: %d", 
			started, completed, errors, pending)
		log.Printf("Success rate: %.2f%%", successRate)
		log.Printf("Queue size: %d", len(wp.taskQueue))
		log.Printf("=============")
	}
}

func (wp *WorkerPool) httpServer() {
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		wp.stats.mu.RLock()
		uptime := time.Since(wp.stats.startTime)
		completed := wp.stats.tasksCompleted
		started := wp.stats.tasksStarted
		errors := wp.stats.errors
		wp.stats.mu.RUnlock()

		fmt.Fprintf(w, "Worker Pool Statistics\n")
		fmt.Fprintf(w, "=====================\n")
		fmt.Fprintf(w, "Uptime: %v\n", uptime)
		fmt.Fprintf(w, "Tasks Started: %d\n", started)
		fmt.Fprintf(w, "Tasks Completed: %d\n", completed)
		fmt.Fprintf(w, "Tasks Failed: %d\n", errors)
		fmt.Fprintf(w, "Queue Size: %d\n", len(wp.taskQueue))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK\n")
	})

	log.Println("HTTP server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("HTTP server error: %v", err)
	}
}

func (wp *WorkerPool) Stop() {
	log.Println("Stopping worker pool...")
	close(wp.taskQueue)
	wp.wg.Wait()
	log.Println("All workers stopped")
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Println("Starting long-running worker process...")
	log.Printf("Process ID: %d", os.Getpid())

	pool := NewWorkerPool(3)
	pool.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Received shutdown signal")
	pool.Stop()
}