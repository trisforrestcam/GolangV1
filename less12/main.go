// ============================================
// LESSON 12: Capstone - Concurrent Web Scraper
// ============================================
// Chạy: go run ./less12
// Mục tiêu: Kết hợp TẤT CẢ kiến thức vào 1 project thực tế

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== Lesson 12: Capstone - Concurrent Web Scraper ===")
	fmt.Println()
	fmt.Println("📌 Kết hợp: Goroutines + Channels + WaitGroup + Context + Error Handling")
	fmt.Println("📌 NodeJS equivalent: Promise.all() với rate limiting")
	fmt.Println("📌 Go: Worker pool với channels - Nhanh hơn 10x!")
	fmt.Println()

	// ===== Configuration =====
	config := ScraperConfig{
		MaxWorkers:     5,
		MaxRetries:     3,
		RequestTimeout: 100 * time.Millisecond,
		RateLimit:      10, // requests per second
	}

	// ===== URLs to scrape =====
	urls := []string{
		"https://api.example.com/users/1",
		"https://api.example.com/users/2",
		"https://api.example.com/users/3",
		"https://api.example.com/users/4",
		"https://api.example.com/users/5",
		"https://api.example.com/users/6",
		"https://api.example.com/users/7",
		"https://api.example.com/users/8",
		"https://api.example.com/users/9",
		"https://api.example.com/users/10",
	}

	// ===== Run scraper =====
	start := time.Now()
	results := scrapeURLs(urls, config)
	duration := time.Since(start)

	// ===== Print results =====
	fmt.Println("\n📊 Results:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	
	successCount := 0
	errorCount := 0
	
	for _, r := range results {
		if r.Error != nil {
			fmt.Printf("❌ %s: %v (%v)\n", r.URL, r.Error, r.Duration)
			errorCount++
		} else {
			fmt.Printf("✅ %s: %s (%v)\n", r.URL, r.Data, r.Duration)
			successCount++
		}
	}

	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Printf("📈 Statistics:\n")
	fmt.Printf("   Total: %d URLs\n", len(urls))
	fmt.Printf("   Success: %d\n", successCount)
	fmt.Printf("   Errors: %d\n", errorCount)
	fmt.Printf("   Duration: %v\n", duration)
	fmt.Printf("   Avg per URL: %v\n", duration/time.Duration(len(urls)))
	fmt.Println()

	// ===== Demonstrate concurrent patterns =====
	fmt.Println("\n🔧 Concurrent Patterns Demo:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	
	// Pattern 1: Fan-out Fan-in
	fmt.Println("\n1️⃣ Fan-out Fan-in:")
	fanOutFanInDemo()

	// Pattern 2: Pipeline
	fmt.Println("\n2️⃣ Pipeline:")
	pipelineDemo()

	// Pattern 3: Worker Pool
	fmt.Println("\n3️⃣ Worker Pool:")
	workerPoolDemo()

	// Pattern 4: Timeout
	fmt.Println("\n4️⃣ Timeout:")
	timeoutDemo()

	// Pattern 5: Rate Limiter
	fmt.Println("\n5️⃣ Rate Limiter:")
	rateLimiterDemo()

	fmt.Println("\n✅ Lesson 12 hoàn thành!")
	fmt.Println("🎓 Bạn đã học xong toàn bộ Go fundamentals!")
}

// ===== TYPES =====

type ScraperConfig struct {
	MaxWorkers     int
	MaxRetries     int
	RequestTimeout time.Duration
	RateLimit      int
}

type ScrapeResult struct {
	URL      string
	Data     string
	Error    error
	Duration time.Duration
}

type PageData struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  int    `json:"status"`
}

// ===== SCRAPER IMPLEMENTATION =====

func scrapeURLs(urls []string, config ScraperConfig) []ScrapeResult {
	// Create channels
	jobs := make(chan string, len(urls))
	results := make(chan ScrapeResult, len(urls))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start workers (Fan-out)
	var wg sync.WaitGroup
	for w := 1; w <= config.MaxWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			scraperWorker(ctx, workerID, jobs, results, config)
		}(w)
	}

	// Send jobs
	go func() {
		for _, url := range urls {
			jobs <- url
		}
		close(jobs)
	}()

	// Wait for workers and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results (Fan-in)
	var allResults []ScrapeResult
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults
}

func scraperWorker(ctx context.Context, workerID int, jobs <-chan string, results chan<- ScrapeResult, config ScraperConfig) {
	for url := range jobs {
		select {
		case <-ctx.Done():
			results <- ScrapeResult{
				URL:   url,
				Error: ctx.Err(),
			}
			continue
		default:
		}

		// Rate limiting
		time.Sleep(time.Second / time.Duration(config.RateLimit))

		// Scrape with retry
		result := scrapeWithRetry(ctx, url, config)
		results <- result
	}
}

func scrapeWithRetry(ctx context.Context, url string, config ScraperConfig) ScrapeResult {
	start := time.Now()

	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		data, err := scrapeURL(ctx, url, config.RequestTimeout)
		if err == nil {
			return ScrapeResult{
				URL:      url,
				Data:     data,
				Duration: time.Since(start),
			}
		}

		if attempt < config.MaxRetries {
			time.Sleep(time.Duration(attempt*50) * time.Millisecond)
		}
	}

	return ScrapeResult{
		URL:      url,
		Error:    fmt.Errorf("max retries exceeded"),
		Duration: time.Since(start),
	}
}

func scrapeURL(ctx context.Context, url string, timeout time.Duration) (string, error) {
	// Simulate HTTP request
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(time.Duration(rand.Intn(50)+20) * time.Millisecond):
		// Simulate random errors (20% chance)
		if rand.Float32() < 0.2 {
			return "", fmt.Errorf("connection timeout")
		}

		// Simulate response
		page := PageData{
			Title:   fmt.Sprintf("Page %s", url),
			Content: fmt.Sprintf("Content from %s", url),
			Status:  200,
		}
		jsonData, _ := json.Marshal(page)
		return string(jsonData), nil
	}
}

// ===== CONCURRENT PATTERNS DEMO =====

// Pattern 1: Fan-out Fan-in
func fanOutFanInDemo() {
	numbers := make(chan int, 10)
	results := make(chan string, 10)

	// Fan-out: 3 workers
	for w := 1; w <= 3; w++ {
		go func(id int) {
			for n := range numbers {
				time.Sleep(10 * time.Millisecond)
				results <- fmt.Sprintf("Worker %d processed %d", id, n)
			}
		}(w)
	}

	// Send numbers
	for i := 1; i <= 9; i++ {
		numbers <- i
	}
	close(numbers)

	// Fan-in: collect results
	time.Sleep(50 * time.Millisecond)
	close(results)
	for r := range results {
		fmt.Printf("   %s\n", r)
	}
}

// Pattern 2: Pipeline
func pipelineDemo() {
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

	// Pipeline: generate → square → double
	pipeline := double(square(generator(1, 2, 3, 4, 5)))
	for result := range pipeline {
		fmt.Printf("   Result: %d\n", result)
	}
}

// Pattern 3: Worker Pool
func workerPoolDemo() {
	jobs := make(chan int, 20)
	results := make(chan int, 20)

	// Start 3 workers
	for w := 1; w <= 3; w++ {
		go func(id int) {
			for job := range jobs {
				time.Sleep(10 * time.Millisecond)
				results <- job * 2
			}
		}(w)
	}

	// Send 10 jobs
	for j := 1; j <= 10; j++ {
		jobs <- j
	}
	close(jobs)

	// Collect results
	time.Sleep(50 * time.Millisecond)
	close(results)
	for r := range results {
		fmt.Printf("   Result: %d\n", r)
	}
}

// Pattern 4: Timeout
func timeoutDemo() {
	slowOperation := func(ch chan<- string) {
		time.Sleep(200 * time.Millisecond)
		ch <- "slow result"
	}

	ch := make(chan string, 1)
	go slowOperation(ch)

	select {
	case result := <-ch:
		fmt.Printf("   Got: %s\n", result)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("   ⏰ Timeout!")
	}
}

// Pattern 5: Rate Limiter
func rateLimiterDemo() {
	limiter := time.NewTicker(100 * time.Millisecond)
	defer limiter.Stop()

	for i := 1; i <= 5; i++ {
		<-limiter.C
		fmt.Printf("   Request %d at %v\n", i, time.Now().Format("15:04:05.000"))
	}
}
