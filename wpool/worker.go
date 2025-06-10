package wpool

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type Worker struct {
	id     int
	pool   *WorkerPool
	mu     sync.RWMutex
	active bool      // от повторного закрытия воркера
	quit   chan bool // канал для сигнала остановки
}

func (w *Worker) Start() {
	defer w.pool.wg.Done()
	defer func() {
		w.mu.Lock()
		w.active = false
		w.mu.Unlock()
	}()

	log.Printf("Воркер %d начал работу", w.id)

	for {
		select {
		case job, ok := <-w.pool.jobs:
			if !ok {
				log.Printf("Воркер %d: канал задач закрыт", w.id)
				return
			}

			w.ProcessJob(job)

		case <-w.quit:
			log.Printf("Воркер %d получил сигнал остановки, завершение работы", w.id)
			return

		case <-w.pool.ctx.Done():
			log.Printf("Воркер %d: контекст отменен", w.id)
			return
		}
	}
}

func (w *Worker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.active {
		w.active = false
		close(w.quit)
	}
}

func (w *Worker) ProcessJob(job Job) {
	w.mu.RLock()
	active := w.active
	w.mu.RUnlock()

	if !active {
		log.Printf("Воркер %d не активен, пропускает задачу: %s", w.id, job)
		return
	}

	log.Printf("Воркер %d обрабатывает: %s", w.id, job)

	// Имитация обработки задачи
	time.Sleep(time.Millisecond * 100)

	var result = Result(fmt.Sprintf("Воркер %d: %s", w.id, strings.ToUpper(string(job)))) // переводим в верхний регистр

	// Отправляем результат
	select {
	case w.pool.results <- result:
		log.Printf("Воркер %d завершил обработку: %s", w.id, job)
	default:
		log.Printf("Воркер %d: канал результатов переполнен, результат потерян", w.id)
	}
}
