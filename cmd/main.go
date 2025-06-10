package main

import (
	"fmt"
	"time"

	"github.com/Noviiich/vk-worker-pool/wpool"
)

func main() {
	// Создаем пул с буфером на 10 задач
	pool := wpool.NewWorkerPool(10)
	defer pool.Close()

	// Добавляем начальных воркеров
	worker1 := pool.AddWorker()
	worker2 := pool.AddWorker()

	jobs := []wpool.Job{
		"task1", "task2", "task3", "task4", "task5",
		"task6", "task7", "task8", "task9", "task10", "task11", "task12",
	}

	go func() {
		for _, job := range jobs {
			pool.SubmitJob(job)
			time.Sleep(time.Millisecond * 50) // небольшая задержка между задачами
		}
	}()

	// Пример динамического управления воркерами
	go func() {
		time.Sleep(time.Millisecond * 200)
		fmt.Printf("Активных воркеров: %d\n", pool.GetActiveWorkers())

		pool.RemoveWorker(worker1)
		fmt.Printf("Удаление воркеров, активных: %d\n", pool.GetActiveWorkers())

		worker3 := pool.AddWorker()
		worker4 := pool.AddWorker()
		fmt.Printf("Добавлены воркеры, активных: %d\n", pool.GetActiveWorkers())

		time.Sleep(time.Millisecond * 300)

		pool.RemoveWorker(worker3)
		fmt.Printf("Удаление воркеров, активных: %d\n", pool.GetActiveWorkers())

		time.Sleep(time.Millisecond * 200)
		pool.AddWorker()
		fmt.Printf("Добавлен воркер, активных: %d\n", pool.GetActiveWorkers())

		pool.RemoveWorker(worker2)
		pool.RemoveWorker(worker4)
	}()

	// Собираем результаты
	go func() {
		processedCount := 0
		for processedCount < len(jobs) {
			if result, ok := pool.GetResult(); ok {
				fmt.Printf("Результат: %s\n", result)
				processedCount++
			} else {
				time.Sleep(time.Millisecond * 10)
			}
		}
	}()

	// Даем время на обработку
	time.Sleep(time.Second * 3)

	fmt.Printf("Финальное количество активных воркеров перед завершением: %d\n", pool.GetActiveWorkers())
}
