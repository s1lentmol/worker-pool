package workerpool

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// captureStdout перехватывает весь вывод в stdout во время выполнения функции f
// и возвращает его в виде строки.
func captureStdout(f func()) string {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	f()

	w.Close()
	os.Stdout = orig

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	return buf.String()
}

// TestWorkerProcessesData проверяет, что воркер корректно обрабатывает задачу
func TestWorkerProcessesData(t *testing.T) {
	wp := NewWorkerPool(10)
	defer wp.Shutdown()

	id := wp.AddWorker()

	output := captureStdout(func() {
		wp.SendWork("task1")
		time.Sleep(200 * time.Millisecond)
	})

	expected := "Worker " + strconv.Itoa(id) + ": Обрабатываю данные: task1"
	if !strings.Contains(output, expected) {
		t.Errorf("Ожидалось, что воркер %d обработает \"task1\", но вывод был:\n%s", id, output)
	}
}

// TestRemoveWorkerStopsProcessing проверяет, что после удаления воркера он больше не обрабатывает задачи
func TestRemoveWorkerStopsProcessing(t *testing.T) {
	wp := NewWorkerPool(5)
	defer wp.Shutdown()

	id := wp.AddWorker()

	output := captureStdout(func() {
		wp.SendWork("task1")
		time.Sleep(200 * time.Millisecond)
	})

	if !strings.Contains(output, "task1") {
		t.Errorf("Ожидалось, что \"task1\" будет обработана, но вывод был:\n%s", output)
	}

	removed := wp.RemoveWorker(id)
	if !removed {
		t.Fatalf("Ожидалось, что RemoveWorker(%d) вернёт true для существующего воркера, но вернул false", id)
	}

	output2 := captureStdout(func() {
		wp.SendWork("task2")
		time.Sleep(200 * time.Millisecond)
	})

	if strings.Contains(output2, "task2") {
		t.Errorf("Ожидалось, что после удаления воркера \"task2\" не будет обработана, но вывод был:\n%s", output2)
	}
}

func TestRemoveNonExistentWorker(t *testing.T) {
	wp := NewWorkerPool(5)
	defer wp.Shutdown()

	// Пробуем удалить воркера с ID, которого нет
	removed := wp.RemoveWorker(999)
	if removed {
		t.Error("Ожидалось, что RemoveWorker для несуществующего воркера вернёт false, но вернул true")
	}
}

// TestShutdownClosesAndWaits проверяет, что Shutdown корректно закрывает пул и ждет завершения всех воркеров
func TestShutdownClosesAndWaits(t *testing.T) {
	wp := NewWorkerPool(5)
	_ = wp.AddWorker()
	_ = wp.AddWorker()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 3 {
			wp.SendWork("data")
		}
	}()

	done := make(chan struct{})
	go func() {
		wp.Shutdown()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Shutdown не завершился за ожидаемое время")
	}

	wg.Wait()
}

// TestMultipleWorkersProcessMultipleTasks проверяет, что несколько воркеров могут одновременно обрабатывать разные задачи
func TestMultipleWorkersProcessMultipleTasks(t *testing.T) {
	wp := NewWorkerPool(10)
	defer wp.Shutdown()

	id1 := wp.AddWorker()
	id2 := wp.AddWorker()

	output := captureStdout(func() {
		wp.SendWork("taskA")
		wp.SendWork("taskB")
		time.Sleep(300 * time.Millisecond)
	})

	if !strings.Contains(output, "taskA") || !strings.Contains(output, "taskB") {
		t.Errorf(
			"Ожидалось, что и \"taskA\", и \"taskB\" будут обработаны воркерами %d и %d, но вывод был:\n%s",
			id1, id2, output,
		)
	}
}
