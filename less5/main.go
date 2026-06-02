// ============================================
// LESSON 5: Concurrency - Goroutines (ĐIỂM MẠNH #1 CỦA GO)
// ============================================
// Chạy: go run ./less5
// Mục tiêu: Hiểu Goroutines - tại sao Go vượt trội hơn NodeJS về concurrency

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== Lesson 5: Goroutines - Concurrency trong Go ===")
	fmt.Println()

	// ===== SO SÁNH: NodeJS vs Go =====
	fmt.Println("📌 NodeJS: Single-threaded event loop")
	fmt.Println("   - async/await: Non-blocking nhưng CHẠY TRÊN 1 THREAD")
	fmt.Println("   - CPU-intensive tasks sẽ BLOCK toàn bộ server")
	fmt.Println()
	fmt.Println("📌 Go: Multi-threaded với Goroutines")
	fmt.Println("   - goroutine: Lightweight thread (~2KB RAM)")
	fmt.Println("   - Có thể chạy hàng TRIỆU goroutines cùng lúc")
	fmt.Println("   - TRUE parallelism trên multi-core CPU")
	fmt.Println()

	// ===== 1. Basic Goroutine =====
	fmt.Println("\n--- 1. Basic Goroutine ---")
	fmt.Println("NodeJS equivalent: setTimeout(() => {}, 0)")
	fmt.Println()

	// Sequential (giống NodeJS)
	start := time.Now()
	task("Task-A", 100)
	task("Task-B", 100)
	task("Task-C", 100)
	fmt.Printf("Sequential: %v\n", time.Since(start))

	// Concurrent với Goroutines
	start = time.Now()
	go task("Task-1", 100)  // Chạy trong background
	go task("Task-2", 100)  // Chạy trong background
	go task("Task-3", 100)  // Chạy trong background
	time.Sleep(150 * time.Millisecond) // Đợi goroutines hoàn thành
	fmt.Printf("Concurrent: %v\n", time.Since(start))

	// ===== 2. WaitGroup - Đợi tất cả goroutines =====
	fmt.Println("\n--- 2. WaitGroup - Đợi tất cả goroutines ---")
	fmt.Println("NodeJS equivalent: await Promise.all([...])")
	fmt.Println()

	var wg sync.WaitGroup
	
	tasks := []string{"Download", "Process", "Upload", "Notify"}
	
	for _, t := range tasks {
		wg.Add(1)  // Tăng counter
		go func(taskName string) {
			defer wg.Done()  // Giảm counter khi hoàn thành
			
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			fmt.Printf("  ✅ %s completed\n", taskName)
		}(t)
	}
	
	wg.Wait() // Đợi counter = 0
	fmt.Println("All tasks completed!")

	// ===== 3. Goroutine với Return Value (Channels) =====
	fmt.Println("\n--- 3. Goroutines với Return Values ---")
	fmt.Println("NodeJS equivalent: const result = await promise")
	fmt.Println()

	// Tạo channel để nhận kết quả
	resultCh := make(chan string)
	
	go func() {
		time.Sleep(50 * time.Millisecond)
		resultCh <- "Hello from goroutine!"  // Gửi kết quả
	}()
	
	result := <-resultCh  // Nhận kết quả (blocking)
	fmt.Printf("Received: %s\n", result)

	// ===== 4. Multiple Return Values =====
	fmt.Println("\n--- 4. Multiple Goroutines trả kết quả ---")
	fmt.Println("NodeJS equivalent: const [a, b, c] = await Promise.all([...])")
	fmt.Println()

	// Tạo channels cho từng kết quả
	ch1 := make(chan string)
	ch2 := make(chan string)
	ch3 := make(chan string)

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch1 <- "Result from API 1"
	}()
	go func() {
		time.Sleep(50 * time.Millisecond)
		ch2 <- "Result from API 2"
	}()
	go func() {
		time.Sleep(75 * time.Millisecond)
		ch3 <- "Result from API 3"
	}()

	// Nhận kết quả (thứ tự không đảm bảo)
	fmt.Printf("  %s\n", <-ch1)
	fmt.Printf("  %s\n", <-ch2)
	fmt.Printf("  %s\n", <-ch3)

	// ===== 5. Fan-out Fan-in Pattern =====
	fmt.Println("\n--- 5. Fan-out Fan-in Pattern ---")
	fmt.Println("NodeJS equivalent: Worker pool với Bull queue")
	fmt.Println()

	jobs := make(chan int, 10)
	results := make(chan int, 10)

	// Fan-out: 3 workers xử lý cùng lúc
	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	// Gửi 9 jobs
	for j := 1; j <= 9; j++ {
		jobs <- j
	}
	close(jobs)

	// Fan-in: Nhận tất cả kết quả
	for r := 1; r <= 9; r++ {
		fmt.Printf("  Result: %d\n", <-results)
	}

	// ===== 6. Race Condition Demo =====
	fmt.Println("\n--- 6. Race Condition - TẠI SAO CẦN CẨN THẬN ---")
	fmt.Println()

	// ⚠️ RACE CONDITION - Không an toàn!
	fmt.Println("⚠️ Without mutex (race condition):")
	counter := 0
	var wg2 sync.WaitGroup
	
	for i := 0; i < 1000; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			counter++  // ⚠️ RACE CONDITION!
		}()
	}
	wg2.Wait()
	fmt.Printf("  Counter (expected 1000): %d\n", counter)  // Có thể < 1000

	// ✅ AN TOÀN với Mutex
	fmt.Println("✅ With mutex (safe):")
	var mu sync.Mutex
	safeCounter := 0
	var wg3 sync.WaitGroup
	
	for i := 0; i < 1000; i++ {
		wg3.Add(1)
		go func() {
			defer wg3.Done()
			mu.Lock()
			safeCounter++
			mu.Unlock()
		}()
	}
	wg3.Wait()
	fmt.Printf("  Counter (expected 1000): %d\n", safeCounter)  // Luôn = 1000

	// ===== 7. Goroutine Leak Prevention =====
	fmt.Println("\n--- 7. Goroutine Leak Prevention ---")
	fmt.Println("NodeJS: Memory leak từ unresolved Promises")
	fmt.Println("Go: Goroutine leak nếu không cancel")
	fmt.Println()

	// Tạo goroutine với timeout
	done := make(chan bool)
	
	go func() {
		time.Sleep(200 * time.Millisecond)
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("  Goroutine completed!")
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  Timeout! Goroutine leaked (but we continue)")
	}

	// ===== Performance Comparison =====
	fmt.Println("\n--- Performance Comparison ---")
	
	// Benchmark: 10,000 goroutines
	start = time.Now()
	var wgPerf sync.WaitGroup
	
	for i := 0; i < 10000; i++ {
		wgPerf.Add(1)
		go func() {
			defer wgPerf.Done()
			// Simulate work
			time.Sleep(time.Microsecond)
		}()
	}
	wgPerf.Wait()
	
	fmt.Printf("10,000 goroutines completed in: %v\n", time.Since(start))
	fmt.Println("(NodeJS với 10,000 async calls sẽ chậm hơn nhiều)")

	fmt.Println("\n✅ Lesson 5 hoàn thành!")
}

func task(name string, duration int) {
	time.Sleep(time.Duration(duration) * time.Millisecond)
	fmt.Printf("  %s done\n", name)
}

func worker(id int, jobs <-chan int, results chan<- int) {
	for job := range jobs {
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		results <- job * 2  // Process job
	}
}
