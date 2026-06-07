# Lesson 7: Concurrency Patterns - Real-World Applications

## 🎯 Mục tiêu
Áp dụng concurrency vào **scenarios thực tế** - worker pools, timeouts, rate limiting, pub/sub, circuit breaker, graceful shutdown.

## 🧠 Mental Model - Góc nhìn Expert

### Go Concurrency là Production-Ready
- NodeJS cần thêm libraries: Bull, p-limit, Agenda
- Go có sẵn trong standard library: goroutines, channels, context, sync

### Context = Request Lifecycle
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Truyền ctx qua mọi function call
result, err := queryDatabase(ctx, sql)
```
- Context quản lý: **deadline, cancellation, request-scoped values**
- **Rule:** Luôn truyền `ctx context.Context` làm parameter đầu tiên

### Worker Pool = Controlled Concurrency
```go
// Giới hạn số goroutines chạy đồng thời
workers := 5
jobs := make(chan Job, 100)
```
- Không spawn unlimited goroutines
- Control resource usage: CPU, memory, connections

### Graceful Shutdown
```go
// Đợi requests đang xử lý hoàn thành trước khi exit
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

<-sigCh
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
server.Shutdown(ctx)
```

## 📚 Kiến thức cốt lõi

### 1. Worker Pool Pattern
```go
type Job struct {
    ID   int
    Data string
}

type Result struct {
    JobID int
    Value string
}

func workerPool(jobs <-chan Job, results chan<- Result, numWorkers int) {
    var wg sync.WaitGroup
    
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for job := range jobs {
                results <- process(job)
            }
        }(i)
    }
    
    wg.Wait()
    close(results)
}
```

### 2. Timeout Pattern
```go
func queryWithTimeout(ctx context.Context, query string) (string, error) {
    type result struct {
        data string
        err  error
    }
    
    done := make(chan result, 1)
    go func() {
        data, err := executeQuery(query)
        done <- result{data, err}
    }()
    
    select {
    case r := <-done:
        return r.data, r.err
    case <-ctx.Done():
        return "", ctx.Err()
    }
}
```

### 3. Rate Limiter
```go
type RateLimiter struct {
    ticker *time.Ticker
}

func NewRateLimiter(rps int) *RateLimiter {
    return &RateLimiter{
        ticker: time.NewTicker(time.Second / time.Duration(rps)),
    }
}

func (r *RateLimiter) Wait() {
    <-r.ticker.C
}
```

### 4. Circuit Breaker
```go
type CircuitBreaker struct {
    state      State
    failures   int
    threshold  int
    lastFailure time.Time
    timeout    time.Duration
    mu         sync.Mutex
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mu.Lock()
    if cb.state == Open && time.Since(cb.lastFailure) < cb.timeout {
        cb.mu.Unlock()
        return errors.New("circuit open")
    }
    cb.mu.Unlock()
    
    err := fn()
    
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        if cb.failures >= cb.threshold {
            cb.state = Open
        }
        return err
    }
    
    cb.failures = 0
    cb.state = Closed
    return nil
}
```

### 5. Pub/Sub Pattern
```go
type EventBus struct {
    subscribers map[string][]chan Event
    mu          sync.RWMutex
}

func (eb *EventBus) Subscribe(topic string) <-chan Event {
    ch := make(chan Event, 1)
    eb.mu.Lock()
    eb.subscribers[topic] = append(eb.subscribers[topic], ch)
    eb.mu.Unlock()
    return ch
}

func (eb *EventBus) Publish(topic string, event Event) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()
    
    for _, ch := range eb.subscribers[topic] {
        select {
        case ch <- event:
        default:  // Non-blocking
        }
    }
}
```

## ⚠️ Sai lầm NodeJS dev hay mắc

1. **Không dùng context**
```go
// ❌ Không thể cancel, timeout
result, err := queryDatabase(sql)

// ✅ Đúng
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
result, err := queryDatabase(ctx, sql)
```

2. **Không giới hạn goroutines**
```go
// ❌ Có thể crash với 1M requests
for _, req := range requests {
    go process(req)  // Spawn unlimited goroutines!
}

// ✅ Worker pool
jobs := make(chan Request, len(requests))
for i := 0; i < 100; i++ {
    go worker(jobs)
}
```

3. **Quên graceful shutdown**
```go
// ❌ Requests đang xử lý bị kill
gofunc() {
    log.Fatal(app.Listen(":3000"))
}()

// ✅ Graceful shutdown
srv := &http.Server{Addr: ":3000", Handler: handler}

go func() { srv.ListenAndServe() }()

<-sigCh
srv.Shutdown(context.Background())
```

## 🏆 Best Practices (Expert Level)

### 1. Luôn truyền context
```go
// ✅ Signature chuẩn
func processRequest(ctx context.Context, req Request) (Response, error)
func queryDatabase(ctx context.Context, sql string) (*Rows, error)
func callAPI(ctx context.Context, url string) ([]byte, error)
```

### 2. Structured concurrency
```go
// ✅ Parent goroutine quản lý child goroutines
func processAll(ctx context.Context, items []Item) error {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()
    
    errCh := make(chan error, len(items))
    var wg sync.WaitGroup
    
    for _, item := range items {
        wg.Add(1)
        go func(i Item) {
            defer wg.Done()
            if err := process(ctx, i); err != nil {
                errCh <- err
                cancel()  // Cancel tất cả
            }
        }(item)
    }
    
    wg.Wait()
    close(errCh)
    
    for err := range errCh {
        return err  // Return lỗi đầu tiên
    }
    return nil
}
```

### 3. Backpressure
```go
// ✅ Giới hạn số jobs trong queue
jobs := make(chan Job, 100)  // Buffer limit

// Client sẽ block khi queue đầy
jobs <- job  // Block nếu 100 jobs đang chờ
```

### 4. Graceful shutdown với timeout
```go
func main() {
    srv := &http.Server{Addr: ":3000"}
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    
    <-quit
    log.Println("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    log.Println("Server exiting")
}
```

## 🔗 Liên kết
- **Previous:** [Lesson 6: Channels, Select, Pipeline](../less6/README.md)
- **Next:** [Lesson 8: Fiber HTTP Server](../less8/README.md)
- **Docs:** [go-fundamentals/04-concurrency-context.md](../../docs/go-fundamentals/04-concurrency-context.md)

---
*"Concurrency patterns trong Go không phải là advanced topic - chúng là standard practice. Một Go expert phải thành thạo worker pools, rate limiting, và graceful shutdown như một NodeJS expert thành thạo async/await."*
