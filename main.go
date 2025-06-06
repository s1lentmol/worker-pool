package main

import (
	"fmt"
	"log"
	"time"
	
	"worker-pool/workerpool"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	fmt.Println("Демонстрация динамического WorkerPool")

	// 1. Создаем worker-pool с буфером для 5 задач
	pool := workerpool.NewWorkerPool(5)

	// 2. Добавляем 3 начальных воркера
	fmt.Println("\n--- Добавляем начальных воркеров (3) ---")
	workerID1 := pool.AddWorker()
	fmt.Printf("Воркер #%d добавлен.\n", workerID1)
	workerID2 := pool.AddWorker()
	fmt.Printf("Воркер #%d добавлен.\n", workerID2)
	workerID3 := pool.AddWorker()
	fmt.Printf("Воркер #%d добавлен.\n", workerID3)

	// 3. Отправляем несколько задач
	fmt.Println("\n--- Отправляем первые 10 задач ---")
	for i := 1; i <= 10; i++ {
		pool.SendWork(fmt.Sprintf("Задача #%d", i))
	}
	time.Sleep(time.Second * 1) // Даем время воркерам поработать

	// 4. Добавляем еще 2 воркера
	fmt.Println("\n--- Добавляем еще 2 воркера ---")
	workerID4 := pool.AddWorker()
	fmt.Printf("Воркер #%d добавлен.\n", workerID4)
	workerID5 := pool.AddWorker()
	fmt.Printf("Воркер #%d добавлен.\n", workerID5)

	// 5. Отправляем еще задач, чтобы новые воркеры тоже получили работу
	fmt.Println("\n--- Отправляем еще 10 задач ---")
	for i := 11; i <= 20; i++ {
		pool.SendWork(fmt.Sprintf("Задача #%d", i))
	}
	time.Sleep(time.Second * 1) // Даем время поработать

	// 6. Удаляем одного из воркеров
	fmt.Printf("\n--- Удаляем воркера %d ---\n", workerID2)
	pool.RemoveWorker(workerID2)
	time.Sleep(time.Millisecond * 500) // Даем время воркеру завершиться

	// 7. Отправляем еще задач, чтобы убедиться, что остальные воркеры продолжают работать
	fmt.Println("\n--- Отправляем последние 5 задач ---")
	for i := 21; i <= 25; i++ {
		pool.SendWork(fmt.Sprintf("Задача #%d", i))
	}
	time.Sleep(time.Second * 1) // Даем время воркерам обработать последние задачи

	// 8. Попытка удалить уже удаленного воркера
	fmt.Printf("\n--- Попытка удалить воркера %d (уже удален) ---\n", workerID2)
	pool.RemoveWorker(workerID2)

	// 9. Удаляем еще одного воркера
	fmt.Printf("\n--- Удаляем воркера %d ---\n", workerID4)
	pool.RemoveWorker(workerID4)
	time.Sleep(time.Millisecond * 500)

	// 10. Завершаем работу пула
	fmt.Println("\n--- Завершение работы WorkerPool ---")
	pool.Shutdown()
	fmt.Println("Все задачи обработаны, пул завершил работу.")
}
