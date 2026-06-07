# Lesson 6: Channels - Communication giữa Goroutines

## 🎯 Mục tiêu
Master **Channels** - cơ chế giao tiếp type-safe giữa các goroutines. Hiểu "Do not communicate by sharing memory; instead, share memory by communicating."

## 🧠 Mental Model - Góc nhìn Expert

### Channels = Typed Communication Pipes
```go
ch := make(chan int)      // Unbuffered
ch := make(chan int, 10)  // Buffered (capacity = 10)
```
- **Unbuffered:** Synchronous - sender block cho đến khi receiver nhận
- **Buffered:** Asynchronous - sender block khi buffer đầy

### "Share Memory by Communicating"
- NodeJS: Shared state + locks (mutex)
- Go: Pass data qua channels (message passing)
- Go proverb: **"Do not communicate by sharing memory; instead, share memory by communicating."**

### Channel Direction
```go
func producer(ch chan<- string) {  // Send-only
    ch <- "data"
}

func consumer(ch <-chan string) string {  // Receive-only
    return <-ch
}
```
- **Compile-time safety** - Không thể nhận từ send-only channel

### Select = Multiplexing Channels
```go
select {
case msg1 := <-ch1:
    // ch1 ready trước
case msg2 := <-ch2:
    // ch2 ready trường hợpcase <-timeout:
    // Timeout
}
```
- **Pseudo-random** nếu nhiều channel ready cùng lúc
- Không deterministic về thứ tự

## 📚 Kiến thức cốt lõi

### Channel Operations
```go
ch <- v      // Send v vào channel
v := <-ch    // Receive từ channel
close(ch)    // Đóng channel
```

### Range over Channel
```go
for v := range ch {
    // Tự động dừng khi channel đóng
}
```

### Select Patterns
```go
// 1. Multiplexing
select {
case v := <-ch1:
case v := <-ch2:
}

// 2. Timeout
select {
case result := <-ch:
case <-time.After(5 * time.Second):
    // Timeout
}

// 3. Non-blocking
select {
case v := <-ch:
    // Có data
default:
    // Không có data, không block
}

// 4. Random selection
for {
    select {
    case v := <-ch1:
        // Random nếu cả 2 ready
    case v := <-ch2:
    }
}
```

### Pipeline Pattern
```go
// Stage 1: Generate
func generator(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

// Stage 2: Process
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

// Connect pipeline
pipeline := square(generator(1, 2, 3, 4, 5))
for result := range pipeline {
    fmt.Println(result)
}
```

## ⚠️ Sai lầm NodeJS dev hay mắc

1. **Gửi vào channel đã đóng**
```go
close(ch)
ch <- 42  // ❌ Panic: send on closed channel

// ✅ Check trước khi send
select {
case ch <- 42:
default:
    // Channel đóng hoặc buffer đầy
}
```

2. **Nhận từ nil channel (block forever)**
```go
var ch chan int  // nil channel
v := <-ch       // ❌ Block forever!

// ✅ Khởi tạo channel
ch := make(chan int)
```

3. **Không đóng channel**
```go
// ❌ Goroutine leak - range không bao giờ dừng
ch := make(chan int)
go func() {
    for v := range ch {  // Block mãi mãi
        fmt.Println(v)
    }
}()
// Quên close(ch)

// ✅ Đóng channel khi xong
go func() {
    for v := range ch {
        fmt.Println(v)
    }
}()
// ... send data
close(ch)
```

4. **Dùng channel thay cho mutex**
```go
// ❌ Không nên - channel quá nặng cho simple counter
type Counter struct {
    ch chan int
    count int
}

// ✅ Đơn giản hơn với sync/atomic
var counter int64
atomic.AddInt64(&counter, 1)
```

## 🏆 Best Practices (Expert Level)

### 1. Channel ownership
```go
// ✅ Owner là goroutine gửi data - owner đóng channel
func producer() <-chan int {
    ch := make(chan int)
    go func() {
        defer close(ch)
        for i := 0; i < 10; i++ {
            ch <- i
        }
    }()
    return ch
}

// Consumer chỉ nhận, không đóng
func consumer(ch <-chan int) {
    for v := range ch {
        fmt.Println(v)
    }
}
```

### 2. Done channel pattern
```go
func worker(done <-chan struct{}) {
    for {
        select {
        case <-done:
            return
        case job := <-jobs:
            process(job)
        }
    }
}
```

### 3. Buffered channel cho fan-out
```go
// ✅ Buffered channel giảm blocking
jobs := make(chan Job, 100)
results := make(chan Result, 100)
```

### 4. Select với done channel
```go
select {
case result := <-ch:
    return result, nil
case <-ctx.Done():
    return nil, ctx.Err()
}
```

### 5. Semaphore pattern
```go
// Giới hạn số goroutines chạy đồng thời
sem := make(chan struct{}, 10)  // Max 10 concurrent

for _, job := range jobs {
    sem <- struct{}{}  // Acquire
    go func(j Job) {
        defer func() { <-sem }()  // Release
        process(j)
    }(job)
}

// Đợi tất cả hoàn thành
for i := 0; i < cap(sem); i++ {
    sem <- struct{}{}
}
```

## 🔗 Liên kết
- **Previous:** [Lesson 5: Goroutines, WaitGroup, Mutex](../less5/README.md)
- **Next:** [Lesson 7: Real-world Concurrency Patterns](../less7/README.md)
- **Docs:** [go-fundamentals/04-concurrency-context.md](../../docs/go-fundamentals/04-concurrency-context.md)

---
*"Channels là superpower của Go. Nếu bạn vẫn dùng mutex cho mọi thứ, bạn chưa hiểu Go concurrency. Channels giúp code dễ đọc hơn, dễ maintain hơn, và tránh race condition tự nhiên."*
