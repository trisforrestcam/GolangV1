// ============================================
// LESSON 7: Concurrency Patterns - Real-World Applications
// ============================================
// Chạy: go run ./less7
// Mục tiêu: Áp dụng concurrency trong các scenarios thực tế

package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== Lesson 7: Real-World Concurrency Patterns ===")
	fmt.Println()

	// ===== 1. Worker Pool - API Rate Limiting =====
	fmt.Println("--- 1. Worker Pool (API Rate Limiting) ---")
	fmt.Println("NodeJS: Bull queue, p-limit")
	fmt.Println("Go: Goroutines + Channels")
	fmt.Println()

	type APIRequest struct {
		ID     int
		Method string
		URL    string
	}

	type APIResponse struct {
		RequestID int
		Status    int
		Body      string
		Duration  time.Duration
	}

	// Worker pool với 3 workers (giới hạn concurrent requests)
	requests := make(chan APIRequest, 10)
	responses := make(chan APIResponse, 10)

	// Start 3 workers
	var wg sync.WaitGroup
	for w := 1; w <= 3; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for req := range requests {
				start := time.Now()
				// Simulate API call
				time.Sleep(time.Duration(rand.Intn(50)+20) * time.Millisecond)
				
				responses <- APIResponse{
					RequestID: req.ID,
					Status:    200,
					Body:      fmt.Sprintf("Response for %s %s", req.Method, req.URL),
					Duration:  time.Since(start),
				}
			}
		}(w)
	}

	// Send requests
	go func() {
		for i := 1; i <= 10; i++ {
			requests <- APIRequest{
				ID:     i,
				Method: "GET",
				URL:    fmt.Sprintf("/api/users/%d", i),
			}
		}
		close(requests)
	}()

	// Collect responses
	go func() {
		wg.Wait()
		close(responses)
	}()

	for resp := range responses {
		fmt.Printf("  Request %d: %d (%v)\n", resp.RequestID, resp.Status, resp.Duration)
	}

	// ===== 2. Timeout Pattern - Database Query =====
	fmt.Println("\n--- 2. Timeout Pattern ---")
	fmt.Println("NodeJS: Promise.race([db.query(), timeout()])")
	fmt.Println("Go: context.WithTimeout")
	fmt.Println()

	queryWithTimeout := func(ctx context.Context, query string) (string, error) {
		resultCh := make(chan string, 1)
		errCh := make(chan error, 1)

		go func() {
			// Simulate slow database query
			time.Sleep(200 * time.Millisecond)
			resultCh <- fmt.Sprintf("Result for: %s", query)
		}()

		select {
		case result := <-resultCh:
			return result, nil
		case err := <-errCh:
			return "", err
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// Query với timeout 100ms (sẽ timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	result, err := queryWithTimeout(ctx, "SELECT * FROM users")
	if err != nil {
		fmt.Printf("  Query 1: %v\n", err)
	} else {
		fmt.Printf("  Query 1: %s\n", result)
	}
	cancel()

	// Query với timeout 300ms (sẽ thành công)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 300*time.Millisecond)
	result2, err2 := queryWithTimeout(ctx2, "SELECT * FROM orders")
	if err2 != nil {
		fmt.Printf("  Query 2: %v\n", err2)
	} else {
		fmt.Printf("  Query 2: %s\n", result2)
	}
	cancel2()

	// ===== 3. Rate Limiter =====
	fmt.Println("\n--- 3. Rate Limiter ---")
	fmt.Println("NodeJS: Bottleneck, express-rate-limit")
	fmt.Println("Go: Ticker + Channel")
	fmt.Println()

	type RateLimiter struct {
		ticker *time.Ticker
		done   chan struct{}
	}

	newRateLimiter := func(rps int) *RateLimiter {
		return &RateLimiter{
			ticker: time.NewTicker(time.Second / time.Duration(rps)),
			done:   make(chan struct{}),
		}
	}

	limiter := newRateLimiter(3) // 3 requests per second
	defer limiter.ticker.Stop()

	for i := 1; i <= 10; i++ {
		<-limiter.ticker.C
		fmt.Printf("  Request %d allowed at %v\n", i, time.Now().Format("15:04:05.000"))
	}

	// ===== 4. Pub/Sub Pattern =====
	fmt.Println("\n--- 4. Pub/Sub Pattern ---")
	fmt.Println("NodeJS: EventEmitter, Redis Pub/Sub")
	fmt.Println("Go: Channel of channels")
	fmt.Println()

	type EventBus struct {
		subscribers map[string][]chan string
		mu          sync.RWMutex
	}

	newEventBus := func() *EventBus {
		return &EventBus{
			subscribers: make(map[string][]chan string),
		}
	}

	bus := newEventBus()

	// Subscribe
	subscribe := func(event string) <-chan string {
		bus.mu.Lock()
		defer bus.mu.Unlock()
		
		ch := make(chan string, 1)
		bus.subscribers[event] = append(bus.subscribers[event], ch)
		return ch
	}

	// Publish
	publish := func(event, data string) {
		bus.mu.RLock()
		defer bus.mu.RUnlock()
		
		for _, ch := range bus.subscribers[event] {
			ch <- data
		}
	}

	// Create subscribers
	userCh := subscribe("user.created")
	orderCh := subscribe("order.created")
	userCh2 := subscribe("user.created")

	// Publish events (async)
	go func() {
		time.Sleep(50 * time.Millisecond)
		publish("user.created", "User #1 created")
		publish("order.created", "Order #100 created")
		publish("user.created", "User #2 created")
	}()

	// Listen for events
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("  User subscriber 1: %s\n", <-userCh)
	fmt.Printf("  Order subscriber: %s\n", <-orderCh)
	fmt.Printf("  User subscriber 2: %s\n", <-userCh2)

	// ===== 5. Circuit Breaker =====
	fmt.Println("\n--- 5. Circuit Breaker ---")
	fmt.Println("NodeJS: opossum")
	fmt.Println("Go: Channel + State machine")
	fmt.Println()

	type CircuitState int

	const (
		StateClosed CircuitState = iota
		StateOpen
		StateHalfOpen
	)

	type CircuitBreaker struct {
		state       CircuitState
		failures    int
		threshold   int
		timeout     time.Duration
		lastFailure time.Time
		mu          sync.Mutex
	}

	newCircuitBreaker := func(threshold int, timeout time.Duration) *CircuitBreaker {
		return &CircuitBreaker{
			state:     StateClosed,
			threshold: threshold,
			timeout:   timeout,
		}
	}

	cb := newCircuitBreaker(3, 1*time.Second)

	execute := func(fn func() error) error {
		cb.mu.Lock()
		
		if cb.state == StateOpen {
			if time.Since(cb.lastFailure) > cb.timeout {
				cb.state = StateHalfOpen
				fmt.Println("  Circuit: Half-Open")
			} else {
				cb.mu.Unlock()
				return fmt.Errorf("circuit open")
			}
		}
		cb.mu.Unlock()

		err := fn()
		
		cb.mu.Lock()
		defer cb.mu.Unlock()
		
		if err != nil {
			cb.failures++
			cb.lastFailure = time.Now()
			if cb.failures >= cb.threshold {
				cb.state = StateOpen
				fmt.Println("  Circuit: Open")
			}
			return err
		}
		
		cb.failures = 0
		cb.state = StateClosed
		return nil
	}

	// Test circuit breaker
	for i := 1; i <= 5; i++ {
		err := execute(func() error {
			if rand.Float32() < 0.7 {
				return fmt.Errorf("service unavailable")
			}
			return nil
		})
		if err != nil {
			fmt.Printf("  Request %d: %v\n", i, err)
		} else {
			fmt.Printf("  Request %d: Success\n", i)
		}
		time.Sleep(200 * time.Millisecond)
	}

	// ===== 6. Graceful Shutdown =====
	fmt.Println("\n--- 6. Graceful Shutdown ---")
	fmt.Println("NodeJS: process.on('SIGTERM', ...)")
	fmt.Println("Go: signal.NotifyContext")
	fmt.Println()

	ctx, cancel = context.WithCancel(context.Background())

	// Start workers
	workerCtx, workerCancel := context.WithCancel(ctx)
	
	go func() {
		for {
			select {
			case <-workerCtx.Done():
				fmt.Println("  Worker: Shutting down gracefully...")
				return
			default:
				fmt.Println("  Worker: Processing...")
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()

	time.Sleep(120 * time.Millisecond)
	fmt.Println("  Main: Sending shutdown signal...")
	workerCancel()
	time.Sleep(60 * time.Millisecond)
	cancel()

	fmt.Println("\n✅ Lesson 7 hoàn thành!")
}
