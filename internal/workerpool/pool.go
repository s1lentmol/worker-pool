package workerpool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// WorkerPool представляет собой динамический пул воркеров
type WorkerPool struct {
	inputChan       chan string
	workers         map[int]context.CancelFunc
	workerIDCounter int
	mu              sync.Mutex
	wg              sync.WaitGroup

	rootCtx    context.Context    // Корневой контекст для всего пула
	rootCancel context.CancelFunc // Функция отмены для корневого контекста
}

// NewWorkerPool создает новый экземпляр WorkerPool.
// bufferSize определяет размер буфера для входного канала.
func NewWorkerPool(bufferSize int) *WorkerPool {
	rootCtx, rootCancel := context.WithCancel(context.Background())
	wp := &WorkerPool{
		inputChan:       make(chan string, bufferSize),
		workers:         make(map[int]context.CancelFunc),
		workerIDCounter: 0,
		rootCtx:         rootCtx,
		rootCancel:      rootCancel,
	}
	log.Printf("WorkerPool: Создан с буфером канала %d.\n", bufferSize)
	return wp
}

// AddWorker добавляет нового воркера в пул.
// Возвращает ID нового воркера.
func (wp *WorkerPool) AddWorker() int {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	wp.workerIDCounter++
	workerID := wp.workerIDCounter

	workerCtx, cancel := context.WithCancel(wp.rootCtx)
	wp.workers[workerID] = cancel

	wp.wg.Add(1)
	go wp.worker(workerCtx, workerID)

	log.Printf("WorkerPool: Воркер %d добавлен.\n", workerID)
	return workerID
}

// RemoveWorker удаляет воркера с заданным ID.
// Возвращает true, если воркер найден и удален, false в противном случае.
func (wp *WorkerPool) RemoveWorker(workerID int) bool {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if cancelFunc, ok := wp.workers[workerID]; ok {
		cancelFunc()
		delete(wp.workers, workerID)
		log.Printf("WorkerPool: Воркер %d помечен на удаление (контекст отменен).\n", workerID)
		return true
	}
	log.Printf("WorkerPool: Воркер %d не найден для удаления.\n", workerID)
	return false
}

// SendWork отправляет данные во входной канал пула.
func (wp *WorkerPool) SendWork(data string) {
	select {
	case wp.inputChan <- data:
		// Данные успешно отправлены
	case <-wp.rootCtx.Done():
		log.Printf("WorkerPool: Не удалось отправить данные '%s', пул завершает работу.\n", data)
	default:
		log.Printf("WorkerPool: Канал входных данных полон, не удалось отправить '%s' немедленно.\n", data)
		wp.inputChan <- data // Это заблокирует, пока место не освободится
	}
}

// Shutdown завершает работу всего пула.
// Все воркеры будут остановлены, и все незавершенные задачи будут обработаны (если канал не будет закрыт немедленно).
func (wp *WorkerPool) Shutdown() {
	log.Println("WorkerPool: Запуск процедуры завершения работы.")
	wp.rootCancel()
	close(wp.inputChan)
	log.Println("WorkerPool: Входной канал закрыт.")
	wp.wg.Wait()
	log.Println("WorkerPool: Все воркеры завершили работу. Пул остановлен.")
}

// worker - это горутина, представляющая отдельный воркер.
func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()
	log.Printf("Worker %d: Запущен.\n", id)

	for {
		select {
		case data, ok := <-wp.inputChan:
			if !ok {
				// Канал закрыт и пуст, воркер может завершить работу.
				log.Printf("Worker %d: Входной канал закрыт. Выход.\n", id)
				return
			}
			// Обработка данных
			fmt.Printf("Worker %d: Обрабатываю данные: %s\n", id, data)
			time.Sleep(time.Millisecond * 100) // Имитация обработки данных
		case <-ctx.Done():
			// Контекст отменен (пул или конкретный воркер).
			// Воркер должен завершить работу.
			log.Printf("Worker %d: Контекст отменен. Выход.\n", id)
			return
		}
	}
}
