package wpool

import (
	"context"
	"log"
	"sync"
)

type Job string

type Result string

type WorkerPool struct {
	workers      map[int]*Worker    // карта воркеров
	jobs         chan Job           // канал для задач
	results      chan Result        // канал для результатов
	workersCount int                // количество воркеров
	ctx          context.Context    // контекст
	cancel       context.CancelFunc // отмена контекста
	mu           sync.RWMutex
	wg           sync.WaitGroup // для ожидания завершения всех воркеров
}

func NewWorkerPool(wCount int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workers:      make(map[int]*Worker),
		jobs:         make(chan Job, wCount),
		results:      make(chan Result, wCount),
		workersCount: 0,
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (wp *WorkerPool) Close() {
	log.Println("Начался процеес закрытия пула воркеров...")

	// Закрываю канал задач
	close(wp.jobs)

	// Останавливаю всех воркеров
	wp.mu.Lock()
	for _, worker := range wp.workers {
		worker.Stop()
	}
	wp.mu.Unlock()

	// Ждем завершения всех воркеров
	wp.wg.Wait()

	// Отмена контекста
	wp.cancel()

	// Закрытие канала с результатами
	close(wp.results)

	log.Println("Пул воркеров закрыт")
}

func (wp *WorkerPool) SubmitJob(job Job) {
	select {
	case wp.jobs <- job:
		// задача успешно отправлена
	case <-wp.ctx.Done():
		// контекст отменен
	}
}

func (wp *WorkerPool) AddWorker() int {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	wp.workersCount++
	worker := &Worker{
		id:     wp.workersCount,
		pool:   wp,
		quit:   make(chan bool),
		active: true,
	}

	wp.workers[wp.workersCount] = worker
	wp.wg.Add(1)

	go worker.Start()

	log.Printf("Воркер %d добавлен в пул", wp.workersCount)
	return wp.workersCount
}

func (wp *WorkerPool) RemoveWorker(workerID int) bool {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	worker, exists := wp.workers[workerID]
	if !exists {
		log.Printf("Воркер %d не найден", workerID)
		return false
	}

	worker.Stop()
	delete(wp.workers, workerID)

	log.Printf("Воркер %d удален из пула", workerID)
	return true
}

func (wp *WorkerPool) GetResult() (Result, bool) {
	select {
	case result := <-wp.results:
		return result, true
	default:
		return "", false
	}
}

func (wp *WorkerPool) GetActiveWorkers() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return len(wp.workers)
}
