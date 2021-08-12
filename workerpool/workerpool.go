package workerpool

import (
	"go-worker/logger"
	"sync"
)

// closeChan - close channel for worker pool
var closeChan chan bool

// wg - wait group for worker pool
var wg sync.WaitGroup

// WorkerPool - holds worker list
type WorkerPool struct {
	workerList []Worker
}

// New - creates a pool of workers
func New(numWorkers int) (*WorkerPool, error) {
	closeChan = make(chan bool, 1)
	workerList := spawnWorkers(numWorkers)
	pool := &WorkerPool{
		workerList: workerList,
	}
	logger.Log.Info("Successfully started worker pool")
	return pool, nil
}

// spawnWorkers - initializes workers based on worker count
func spawnWorkers(numWorkers int) []Worker {
	var workerList []Worker
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		workerID := i + 1
		worker := NewWorker(workerID)
		worker.Init()
		workerList = append(workerList, *worker)
		logger.Log.Infof("Worker %d initialized successfully", workerID)
	}
	return workerList
}

// Close - stop all the workers
func (wp *WorkerPool) Close() {
	close(closeChan)
	wg.Wait()
	logger.Log.Info("Successfully closed all workers")
}
