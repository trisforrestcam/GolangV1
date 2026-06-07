# Lesson 5: Concurrency - Goroutines

## 🎯 Mục tiêu
Master **Goroutines** - điểm mạnh #1 của Go. Hiểu concurrency model khác biệt so với NodeJS.

## 🧠 Mental Model - Góc nhìn Expert

### Go Concurrency ≠ NodeJS Async
| | NodeJS | Go |
|---|---|---|
| Model | Single-threaded event loop | Multi-threaded (Goroutines) |
| Async | Callbacks/Promises | Goroutines + Channels |
| CPU tasks | BLOCKS event loop | Chạy song song thực sự |
| Memory | ~1MB per thread | ~2KB per goroutine |
| Scale | Limited by event loop | Hàng triệu goroutines |

### Goroutines là "Lightweight Threads"
```go
go func() {
    // Chạy concurrently, non-blocking
}()
```
- **Không phải OS thread** - được Go runtime quản lý
- **Một OS thread chạy nhiều goroutines** (M:N threading)
- Go scheduler tự động multiplex goroutines lên OS threads

### WaitGroup = Promise.all()
```go
var wg sync.WaitGroup
for _, task := range tasks {
    wg.Add(1)
    go func(t Task) {
        defer wg.Done()
        process(t)
    }(task)
}
wg.Wait()  // Đợi tất cả hoàn thành
```

### Mutex = Mutual Exclusion
```go
var mu sync.Mutex
var counter int

// Safe concurrent access
mu.Lock()
counter++
mu.Unlock()
```

## 📚 Kiến thức cốt lõi

### Goroutine Basics
```go
// Chạy function trong background
go doWork()

// Anonymous goroutine
go func() {
    fmt.Println("Running in background")
}()
```

### WaitGroup Pattern
```go
func processItems(items []string) {
    var wg sync.WaitGroup
    results := make(chan string, len(items))
    
    for _, item := range items {
        wg.Add(1)
        go func(i string) {
            defer wg.Done()
            results <- process(i)
        }(item)
    }
    
    go func() {
        wg.Wait()
        close(results)
    }()
    
    for r := range results {
        fmt.Println(r)
    }
}
```

### Mutex vs RWMutex
```go
// sync.Mutex - Mutual exclusion
var mu sync.Mutex
mu.Lock()
// ... write/read
mu.Unlock()

// sync.RWMutex - Multiple readers, single writer
var rwmu sync.RWMutex
rwmu.RLock()    // Read lock (multiple goroutines có thể đọc)
// ... read
rwmu.RUnlock()

rwmu.Lock()     // Write lock (chỉ 1 goroutine)
// ... write
rwmu.Unlock()
```

## ⚠️ Sai lầm NodeJS dev hay mắc

1. **Tưởng goroutine = async/await**
```go
// ❌ Sai
result := go calculate()  // Không thể gán kết quả trực tiếp!

// ✅ Đúng - Dùng channel
ch := make(chan int)
go func() {
    ch <- calculate()
}()
result := <-ch
```

2. **Race condition - Quên mutex**
```go
// ❌ Race condition
counter := 0
for i := 0; i < 1000; i++ {
    go func() {
        counter++  // Không safe!
    }()
}

// ✅ Đúng
var mu sync.Mutex
counter := 0
for i := 0; i < 1000; i++ {
    go func() {
        mu.Lock()
        counter++
        mu.Unlock()
    }()
}
```

3. **Không đợi goroutines hoàn thành**
```go
// ❌ Goroutines chưa chạy xong đã exit
for _, url := range urls {
    go fetch(url)
}
// Program exit ngay!

// ✅ Đúng
var wg sync.WaitGroup
for _, url := range urls {
    wg.Add(1)
    go func(u string) {
        defer wg.Done()
        fetch(u)
    }(url)
}
wg.Wait()
```

## 🏆 Best Practices (Expert Level)

### 1. Luôn truyền parameters vào goroutine
```go
// ❌ Bug: i thay đổi trước khi goroutine chạy
for i := 0; i < 10; i++ {
    go func() {
        fmt.Println(i)  // i có thể = 10 khi chạy!
    }()
}

// ✅ Đúng
for i := 0; i < 10; i++ {
    go func(n int) {
        fmt.Println(n)
    }(i)
}
```

### 2. Dùng `go vet -race` để phát hiện race condition
```bash
go run -race ./less5   # Detect race conditions
go build -race ./...   # Build với race detector
```

### 3. Prefer channels over shared memory
```go
// ✅ Go way - Dùng channels
type Result struct {
    Value int
    Error error
}

ch := make(chan Result)
go func() {
    v, err := calculate()
    ch <- Result{Value: v, Error: err}
}()
result := <-ch

// ❌ C-style - Shared memory + mutex
var result int
var err error
var mu sync.Mutex
go func() {
    mu.Lock()
    result, err = calculate()
    mu.Unlock()
}()
```

### 4. Context cho cancellation
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

for _, task := range tasks {
    go func(t Task) {
        select {
        case <-ctx.Done():
            return
        default:
            process(t)
        }
    }(task)
}

// Cancel tất cả
cancel()
```

### 5. Worker pool pattern
```go
func workerPool(jobs <-chan Job, results chan<- Result, workers int) {
    var wg sync.WaitGroup
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- process(job)
            }
        }()
    }
    wg.Wait()
    close(results)
}
```

## 🔗 Liên kết
- **Previous:** [Lesson 4: Structs, Methods, Composition](../less4/README.md)
- **Next:** [Lesson 6: Channels, Select, Pipeline](../less6/README.md)
- **Docs:** [go-fundamentals/04-concurrency-context.md](../../docs/go-fundamentals/04-concurrency-context.md)

---
*"Đừng nghĩ goroutine là async/await nâng cao. Goroutine là true parallelism - code chạy song song thực sự trên nhiều CPU cores. Đây là lý do Go nhanh hơn NodeJS 10x trong CPU-intensive tasks."*
