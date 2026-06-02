// ============================================
// LESSON 6: Channels - Communication giữa Goroutines
// ============================================
// Chạy: go run ./less6
// Mục tiêu: Hiểu Channels - cách goroutines giao tiếp với nhau

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== Lesson 6: Channels ===")
	fmt.Println()
	fmt.Println("📌 Channels = Typed communication pipes giữa goroutines")
	fmt.Println("   NodeJS: EventEmitter, Promises, Callbacks")
	fmt.Println("Go Channels: Type-safe, blocking, compile-time checked")
	fmt.Println()

	// ===== 1. Basic Channel =====
	fmt.Println("\n--- 1. Basic Channel ---")
	
	ch := make(chan string)  // Unbuffered channel
	
	go func() {
		time.Sleep(50 * time.Millisecond)
		ch <- "Hello from goroutine!"  // Send
	}()
	
	msg := <-ch  // Receive (blocking until data available)
	fmt.Printf("Received: %s\n", msg)

	// ===== 2. Buffered Channel =====
	fmt.Println("\n--- 2. Buffered Channel ---")
	fmt.Println("NodeJS: Không có equivalent (cần custom queue)")
	fmt.Println()

	buffered := make(chan int, 3)  // Buffer size = 3
	
	// Non-blocking send (nếu buffer chưa đầy)
	buffered <- 1
	buffered <- 2
	buffered <- 3
	// buffered <- 4  // ⚠️ BLOCK vì buffer đầy!
	
	fmt.Printf("Received: %d\n", <-buffered)
	fmt.Printf("Received: %d\n", <-buffered)
	fmt.Printf("Received: %d\n", <-buffered)

	// ===== 3. Channel Direction (Send-only, Receive-only) =====
	fmt.Println("\n--- 3. Channel Direction ---")
	
	producer := func(ch chan<- string) {  // Send-only
		ch <- "data from producer"
	}
	
	consumer := func(ch <-chan string) string {  // Receive-only
		return <-ch
	}
	
	ch2 := make(chan string)
	go producer(ch2)
	fmt.Printf("Consumed: %s\n", consumer(ch2))

	// ===== 4. Range over Channel =====
	fmt.Println("\n--- 4. Range over Channel ---")
	fmt.Println("NodeJS equivalent: for await (const item of stream)")
	fmt.Println()

	numCh := make(chan int)
	
	go func() {
		for i := 1; i <= 5; i++ {
			numCh <- i
			time.Sleep(30 * time.Millisecond)
		}
		close(numCh)  // Đóng channel khi xong
	}()
	
	// Range sẽ tự động dừng khi channel đóng
	for num := range numCh {
		fmt.Printf("  Received: %d\n", num)
	}

	// ===== 5. Select - Multiplexing Channels =====
	fmt.Println("\n--- 5. Select - Multiplexing Channels ===")
	fmt.Println("NodeJS equivalent: Promise.race([...])")
	fmt.Println()

	ch1 := make(chan string)
	ch3 := make(chan string)

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch1 <- "Result from API 1"
	}()

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch3 <- "Result from API 2"
	}()

	// Select đợi channel nào ready trước
	select {
	case msg := <-ch1:
		fmt.Printf("  ch1: %s\n", msg)
	case msg := <-ch3:
		fmt.Printf("  ch2: %s\n", msg)
	}

	// ===== 6. Select with Timeout =====
	fmt.Println("\n--- 6. Select with Timeout ---")
	fmt.Println("NodeJS equivalent: Promise.race([promise, timeout])")
	fmt.Println()

	slowCh := make(chan string)
	
	go func() {
		time.Sleep(200 * time.Millisecond)
		slowCh <- "slow result"
	}()

	select {
	case result := <-slowCh:
		fmt.Printf("  Got: %s\n", result)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  ⏰ Timeout!")
	}

	// ===== 7. Select with Default (Non-blocking) =====
	fmt.Println("\n--- 7. Non-blocking Select ---")
	
	ch4 := make(chan string, 1)
	
	select {
	case msg := <-ch4:
		fmt.Printf("  Got: %s\n", msg)
	default:
		fmt.Println("  No data available (non-blocking)")
	}
	
	ch4 <- "now have data"
	
	select {
	case msg := <-ch4:
		fmt.Printf("  Got: %s\n", msg)
	default:
		fmt.Println("  No data available")
	}

	// ===== 8. Done Channel Pattern =====
	fmt.Println("\n--- 8. Done Channel Pattern ---")
	fmt.Println("NodeJS: AbortController")
	fmt.Println("Go: done channel")
	fmt.Println()

	done := make(chan struct{})  // struct{} không tốn memory
	
	go func() {
		for {
			select {
			case <-done:
				fmt.Println("  Worker: Received done signal, stopping...")
				return
			default:
				fmt.Println("  Worker: Working...")
				time.Sleep(30 * time.Millisecond)
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)
	close(done)  // Signal tất cả goroutines đang đợi
	time.Sleep(10 * time.Millisecond)

	// ===== 9. Pipeline Pattern =====
	fmt.Println("\n--- 9. Pipeline Pattern ---")
	fmt.Println("NodeJS: Streams, Transform streams")
	fmt.Println("Go: Channels connected in pipeline")
	fmt.Println()

	// Stage 1: Generate numbers
	generator := func(nums ...int) <-chan int {
		out := make(chan int)
		go func() {
			for _, n := range nums {
				out <- n
			}
			close(out)
		}()
		return out
	}

	// Stage 2: Square numbers
	square := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			for n := range in {
				out <- n * n
			}
			close(out)
		}()
		return out
	}

	// Stage 3: Double numbers
	double := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			for n := range in {
				out <- n * 2
			}
			close(out)
		}()
		return out
	}

	// Connect pipeline: generate → square → double
	pipeline := double(square(generator(1, 2, 3, 4, 5)))
	
	for result := range pipeline {
		fmt.Printf("  Pipeline result: %d\n", result)
	}

	// ===== 10. Fan-out Fan-in với Channels =====
	fmt.Println("\n--- 10. Fan-out Fan-in ---")
	
	jobs := make(chan int, 10)
	results := make(chan int, 10)

	// Start 3 workers
	for w := 1; w <= 3; w++ {
		go fanWorker(w, jobs, results)
	}

	// Send 9 jobs
	for j := 1; j <= 9; j++ {
		jobs <- j
	}
	close(jobs)

	// Collect results
	for r := 1; r <= 9; r++ {
		fmt.Printf("  Result: %d\n", <-results)
	}

	// ===== 11. Channel of Channels =====
	fmt.Println("\n--- 11. Channel of Channels ---")
	
	chanOfChans := make(chan chan string)
	
	go func() {
		innerCh := make(chan string)
		chanOfChans <- innerCh
		innerCh <- "Hello from inner channel!"
	}()
	
	innerCh := <-chanOfChans
	fmt.Printf("  Received: %s\n", <-innerCh)

	// ===== 12. Semaphore Pattern =====
	fmt.Println("\n--- 12. Semaphore Pattern ---")
	fmt.Println("Giới hạn số goroutines chạy cùng lúc")
	fmt.Println()

	semaphore := make(chan struct{}, 3)  // Max 3 concurrent
	var semWg sync.WaitGroup
	
	for i := 1; i <= 10; i++ {
		semWg.Add(1)
		go func(id int) {
			defer semWg.Done()
			
			semaphore <- struct{}{}  // Acquire (blocking nếu đầy)
			defer func() { <-semaphore }()  // Release
			
			fmt.Printf("  Task %d running (concurrent)\n", id)
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		}(i)
	}
	
	semWg.Wait()
	fmt.Println("All tasks completed!")

	fmt.Println("\n✅ Lesson 6 hoàn thành!")
}

func fanWorker(id int, jobs <-chan int, results chan<- int) {
	for job := range jobs {
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		results <- job * 2
	}
}
