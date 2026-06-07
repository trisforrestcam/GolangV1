# Lesson 12: Capstone - Concurrent Web Scraper

## 🎯 Mục tiêu
Tổng hợp **TẤT CẢ** kiến thức đã học vào một project thực tế: Concurrent Web Scraper. Đây là bài kiểm tra cuối cùng cho việc trở thành Go expert.

## 🧠 Mental Model - Góc nhìn Expert

### Go = Concurrency + Simplicity
- Bài này kết hợp: Goroutines, Channels, WaitGroup, Context, Error Handling, Structs, Methods
- **Không cần external libraries** - standard library đủ mạnh
- Go concurrency model giúp code **ngắn gọn hơn NodeJS** nhưng **performance cao hơn**

### Design Pattern: Fan-out Fan-in
```
URLs → [Worker Pool] → Results
        5 goroutines
```
- **Fan-out:** Mở rộng số workers xử lý
- **Fan-in:** Gom kết quả về một channel
- **Worker Pool:** Giới hạn concurrent requests (tránh overwhelm server)

### Context cho Lifecycle Management
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Truyền ctx xuống mọi layer
scraperWorker(ctx, ...)
```
- Context quản lý: **timeout, cancellation, deadlines**
- Khi context cancelled, tất cả goroutines nhận signal và dừng

## 📚 Kiến thức tổng hợp

### Architecture
```go
// Config
type ScraperConfig struct {
    MaxWorkers     int
    MaxRetries     int
    RequestTimeout time.Duration
    RateLimit      int
}

// Result
type ScrapeResult struct {
    URL      string
    Data     string
    Error    error
    Duration time.Duration
}
```

### Worker Pool Implementation
```go
func scrapeURLs(urls []string, config ScraperConfig) []ScrapeResult {
    jobs := make(chan string, len(urls))
    results := make(chan ScrapeResult, len(urls))
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Fan-out: Start workers
    var wg sync.WaitGroup
    for w := 1; w <= config.MaxWorkers; w++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            scraperWorker(ctx, id, jobs, results, config)
        }(w)
    }
    
    // Send jobs
    go func() {
        for _, url := range urls {
            jobs <- url
        }
        close(jobs)
    }()
    
    // Fan-in: Close results when done
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Collect results
    var allResults []ScrapeResult
    for result := range results {
        allResults = append(allResults, result)
    }
    
    return allResults
}
```

### Retry với Exponential Backoff
```go
func scrapeWithRetry(ctx context.Context, url string, config ScraperConfig) ScrapeResult {
    start := time.Now()
    
    for attempt := 1; attempt <= config.MaxRetries; attempt++ {
        data, err := scrapeURL(ctx, url, config.RequestTimeout)
        if err == nil {
            return ScrapeResult{URL: url, Data: data, Duration: time.Since(start)}
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
```

## ⚠️ Sai lầm NodeJS dev hay mắc

1. **Không dùng context**
```go
// ❌ Không thể cancel
go scraperWorker(jobs, results, config)

// ✅ Dùng context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
go scraperWorker(ctx, jobs, results, config)
```

2. **Không giới hạn workers**
```go
// ❌ Spawn goroutine cho mỗi URL
for _, url := range urls {
    go scrape(url)  // 1000 URLs = 1000 goroutines!
}

// ✅ Worker pool
for w := 0; w < 5; w++ {
    go worker(jobs, results)
}
```

3. **Quên close channels**
```go
// ❌ Goroutine leak
for result := range results {
    // Block mãi nếu không close
}

// ✅ Đóng channel
go func() {
    wg.Wait()
    close(results)
}()
```

4. **Không handle timeouts**
```go
// ❌ Request chạy mãi mãi
data, err := http.Get(url)

// ✅ Timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
data, err := http.DefaultClient.Do(req)
```

## 🏆 Best Practices (Expert Level)

### 1. Structured concurrency
```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Start scraper
    resultCh := make(chan Result)
    go scraper(ctx, urls, resultCh)
    
    // Handle Ctrl+C
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT)
    
    select {
    case result := <-resultCh:
        fmt.Println(result)
    case <-sigCh:
        cancel()  // Cancel tất cả goroutines
    }
}
```

### 2. Rate limiting
```go
type RateLimiter struct {
    ticker *time.Ticker
}

func (r *RateLimiter) Wait() {
    <-r.ticker.C
}

// Giới hạn: 10 requests/second
limiter := &RateLimiter{ticker: time.NewTicker(100 * time.Millisecond)}
```

### 3. Metrics và logging
```go
type ScrapeResult struct {
    URL      string
    Data     string
    Error    error
    Duration time.Duration
    Attempts int
}

// Collect metrics
var (
    successCount int
    errorCount   int
    totalDuration time.Duration
)

for _, r := range results {
    if r.Error != nil {
        errorCount++
        log.Printf("❌ %s: %v", r.URL, r.Error)
    } else {
        successCount++
        totalDuration += r.Duration
        log.Printf("✅ %s: %s (%v)", r.URL, r.Data, r.Duration)
    }
}
```

### 4. Graceful shutdown
```go
func main() {
    // ... setup
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    
    <-quit
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Cancel scraper
    scraperCancel()
    
    // Đợi goroutines dừng
    select {
    case <-doneCh:
        log.Println("Scraper stopped gracefully")
    case <-ctx.Done():
        log.Println("Forced shutdown")
    }
}
```

### 5. Error aggregation
```go
// Collect tất cả errors
type Result struct {
    Data  []byte
    Error error
}

results := make([]Result, len(urls))
for i, url := range urls {
    data, err := fetch(url)
    results[i] = Result{Data: data, Error: err}
}

// Report
var errors []error
for _, r := range results {
    if r.Error != nil {
        errors = append(errors, r.Error)
    }
}

if len(errors) > 0 {
    return fmt.Errorf("scraper completed with %d errors: %v", len(errors), errors)
}
```

## 🔗 Liên kết
- **Previous:** [Lesson 11: Error Handling Patterns](../less11/README.md)
- **Docs:** [go-fundamentals/04-concurrency-context.md](../../docs/go-fundamentals/04-concurrency-context.md)
- **Docs:** [go-fundamentals/07-fast-track-roadmap.md](../../docs/go-fundamentals/07-fast-track-roadmap.md)

---
*"Bài capstone này là tổng hợp của mọi thứ bạn đã học. Nếu bạn hiểu được worker pools, fan-out fan-in, context cancellation, và error handling trong bài này - bạn đã sẵn sàng viết Go production-ready."*

## 🎓 Bạn đã học xong!

Từ Lesson 1 đến Lesson 12, bạn đã nắm vững:
- ✅ Basic syntax & types
- ✅ Functions, closures, defer
- ✅ Structs, methods, composition
- ✅ Goroutines, WaitGroup, Mutex
- ✅ Channels, select, pipeline
- ✅ Real-world concurrency patterns
- ✅ HTTP API với Fiber
- ✅ JSON handling
- ✅ Interfaces, polymorphism, generics
- ✅ Error handling patterns
- ✅ Concurrent web scraper (capstone)

**Next steps:**
1. Xây dựng project thực tế với Go
2. Đọc [docs/nodejs-to-go/](../../docs/nodejs-to-go/) để hiểu sâu hơn
3. Đọc [docs/go-fundamentals/](../../docs/go-fundamentals/) để nâng cao
4. Contribute to open-source Go projects
5. Tham gia cộng đồng Go Việt Nam

**Chúc mừng bạn đã trở thành Go developer! 🎉**
