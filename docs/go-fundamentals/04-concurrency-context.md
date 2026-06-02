# Concurrency trong Go - Hướng Dẫn Toàn Tập

> **Mục tiêu:** Hiểu sâu concurrency model của Go, tại sao nó mạnh hơn Node.js trong nhiều scenarios, và cách sử dụng hiệu quả.
>
> **Yêu cầu:** Đã đọc `01-fundamentals.md` và `02-advanced.md`.

---

## 1. Concurrency vs Parallelism - Phân biệt cơ bản

Đây là 2 khái niệm **hoàn toàn khác nhau** mà nhiều người hay nhầm lẫn.

### Concurrency = Xử lý nhiều thứ cùng lúc (dealing with multiple things)

```
Timeline (1 core):
Task A: ██░░██░░██░░
Task B: ░░██░░██░░██
        ──────────────► time

Luân phiên giữa các task - mỗi thời điểm chỉ 1 task chạy
```

### Parallelism = Thực hiện nhiều thứ cùng lúc (doing multiple things)

```
Timeline (2 cores):
Core 1 - Task A: ████████████
Core 2 - Task B: ████████████
                 ──────────────► time

Thực sự chạy đồng thời trên nhiều core
```

### So sánh Node.js vs Go

```
┌─────────────────────────────────────────────────────────┐
│                    Node.js Model                         │
│                                                         │
│  ┌─────────────────────────────────────┐                │
│  │         Single Thread               │                │
│  │  ┌─────────────────────────────┐    │                │
│  │  │       Event Loop             │    │                │
│  │  │  ┌─────┐ ┌─────┐ ┌─────┐   │    │                │
│  │  │  │ CB1 │→│ CB2 │→│ CB3 │   │    │                │
│  │  │  └─────┘ └─────┘ └─────┘   │    │                │
│  │  └─────────────────────────────┘    │                │
│  └─────────────────────────────────────┘                │
│                                                         │
│  ✅ Concurrent (xử lý nhiều I/O cùng lúc)              │
│  ❌ NOT Parallel (không tận dụng đa core)              │
│  ❌ CPU-bound tasks sẽ BLOCK toàn bộ event loop        │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                      Go Model                           │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐                     │
│  │  OS Thread 1 │  │  OS Thread 2 │                     │
│  │ ┌──┐┌──┐┌──┐│  │ ┌──┐┌──┐┌──┐│                     │
│  │ │G1││G2││G3││  │ │G4││G5││G6││                     │
│  │ └──┘└──┘└──┘│  │ └──┘└──┘└──┘│                     │
│  └──────────────┘  └──────────────┘                     │
│                                                         │
│  ✅ Concurrent (nhiều goroutine chạy cùng lúc)         │
│  ✅ Parallel (tận dụng đa core thực sự)                │
│  ✅ CPU-bound tasks chạy song song trên nhiều core     │
└─────────────────────────────────────────────────────────┘
```

### Tại sao điều này quan trọng?

```go
// Go: CPU-bound task chạy PARALLEL - nhanh gấp N lần (N = số core)
// Mỗi goroutine chạy trên 1 core khác nhau

func heavyComputation(data []float64) float64 {
    // Tính toán nặng - chạy trên nhiều core
    sum := 0.0
    for _, v := range data {
        sum += math.Sqrt(v) * math.Sin(v)
    }
    return sum
}
```

```javascript
// Node.js: CPU-bound task BLOCK event loop
// Phải dùng Worker Threads (rườm rà) hoặc chuyển sang process khác

const { Worker } = require('worker_threads');

// Worker Threads trong Node.js rất cồng kềnh so với goroutine
const worker = new Worker('./heavy-task.js');
```

---

## 2. Node.js Event Loop Model - Ôn tập để so sánh

Trước khi hiểu Go, hãy ôn lại Node.js:

```
Node.js Event Loop (đơn giản hóa):

   ┌───────────────────────────┐
┌─>│           timers          │  setTimeout, setInterval
│  └─────────────┬─────────────┘
│  ┌─────────────┴─────────────┐
│  │     pending callbacks     │  I/O callbacks
│  └─────────────┬─────────────┘
│  ┌─────────────┴─────────────┐
│  │       idle, prepare       │  internal only
│  └─────────────┬─────────────┘
│  ┌─────────────┴─────────────┐
│  │           poll            │  incoming I/O events
│  └─────────────┬─────────────┘
│  ┌─────────────┴─────────────┐
│  │           check           │  setImmediate
│  └─────────────┬─────────────┘
│  ┌─────────────┴─────────────┐
│  │      close callbacks      │  socket.on('close')
│  └─────────────┬─────────────┘
│                │
└────────────────┘
```

### Node.js xử lý I/O như thế nào?

```javascript
// Node.js: Non-blocking I/O nhờ event loop
const fs = require('fs');

// Callback được đăng ký, event loop sẽ gọi khi I/O xong
fs.readFile('file1.txt', (err, data1) => {
    console.log(data1);
});

fs.readFile('file2.txt', (err, data2) => {
    console.log(data2);
});

// Cả 2 file đọc cùng lúc (non-blocking)
// Nhưng chỉ 1 callback chạy tại 1 thời điểm
```

### Node.js Worker Threads (cồng kén)

```javascript
// Node.js Worker Threads - rất cồng kềnh
const { Worker, isMainThread, parentPort } = require('worker_threads');

if (isMainThread) {
    const worker = new Worker(__filename);
    worker.on('message', (msg) => console.log(msg));
    worker.postMessage('start');
} else {
    parentPort.on('message', (msg) => {
        // Tính toán nặng trong worker
        const result = heavyComputation();
        parentPort.postMessage(result);
    });
}
```

```go
// Go equivalent - đơn giản hơn RẤT NHIỀU
go func() {
    result := heavyComputation()
    ch <- result
}()
```

### Node.js Cluster (multi-process)

```javascript
// Node.js Cluster - mỗi process có 1 event loop riêng
const cluster = require('cluster');
const http = require('http');
const numCPUs = require('os').cpus().length;

if (cluster.isPrimary) {
    for (let i = 0; i < numCPUs; i++) {
        cluster.fork(); // Fork process
    }
} else {
    http.createServer((req, res) => {
        res.end('Hello');
    }).listen(8000);
}
```

```go
// Go equivalent - không cần cluster, goroutine tự xử lý
http.ListenAndServe(":8000", handler)
// Go HTTP server tự động spawn goroutine cho mỗi request
// Và tự động chạy parallel trên nhiều core
```

### Node.js chạm tường khi nào?

```javascript
// ❌ BLOCK event loop - toàn bộ server bị đơ
app.get('/heavy', (req, res) => {
    let sum = 0;
    for (let i = 0; i < 1e9; i++) {  // CPU-bound
        sum += i;
    }
    res.json({ sum });
});

// Mọi request khác phải đợi task này xong
// Dù có 1000 connections, chỉ 1 request xử lý tại 1 thời điểm
```

```go
// ✅ Go: Mỗi request 1 goroutine, chạy PARALLEL trên nhiều core
func heavyHandler(w http.ResponseWriter, r *http.Request) {
    sum := 0
    for i := 0; i < 1e9; i++ { // CPU-bound
        sum += i
    }
    fmt.Fprintf(w, `{"sum": %d}`, sum)
}
// Request khác vẫn chạy bình thường trên core khác
```

---

## 3. Goroutines - Hướng dẫn toàn tập

### Goroutine là gì?

Goroutine là lightweight thread được quản lý bởi Go runtime, không phải OS thread.

```
So sánh kích thước:

OS Thread:     ~1MB stack (cố định)
Green Thread:  ~8KB-64KB (tùy ngôn ngữ)
Goroutine:     ~2KB stack (ban đầu, có thể grow/shrink)

→ Bạn có thể chạy hàng TRIỆU goroutine trên 1 machine
→ Node.js: mỗi Promise chỉ là object trong event loop, không phải thread
```

### Syntax cơ bản

```go
// Chạy function như goroutine
go myFunction()

// Chạy anonymous function như goroutine
go func() {
    fmt.Println("Hello from goroutine")
}()

// Chạy method như goroutine
go myStruct.myMethod()
```

```javascript
// Node.js equivalent - KHÔNG có syntax tương đương
// Phải dùng Promise hoặc async/await (không phải concurrency thực sự)

// Promise.all vẫn chạy tuần tự trong 1 thread
await Promise.all([
    fetchData1(),
    fetchData2(),
]);

// Nếu muốn parallel thực sự → Worker Threads (cồng kềnh)
```

### Goroutine Lifecycle

```
Goroutine Lifecycle:

  Created ──► Runnable ──► Running ──► Terminated
                    ↑          │
                    │          ▼
                    └──── Blocked (I/O, channel, syscall)

- Created: khi gặp `go` keyword
- Runnable: chờ Go scheduler gán vào Processor
- Running: đang thực thi trên OS thread
- Blocked: đang chờ (I/O, channel receive, lock, etc.)
- Terminated: function kết thúc hoặc gọi runtime.Goexit()
```

### Goroutine vs OS Thread vs Node.js Promise

```
┌──────────────────┬──────────────┬──────────────┬─────────────────┐
│                  │ OS Thread    │ Goroutine    │ Node.js Promise │
├──────────────────┼──────────────┼──────────────┼─────────────────┤
│ Stack size       │ ~1MB         │ ~2KB (grow)  │ N/A (event loop)│
│ Creation cost    │ ~1ms         │ ~0.3μs       │ ~0.1μs          │
│ Context switch   │ ~1-10μs      │ ~0.2μs       │ N/A             │
│ Max concurrent   │ ~10K         │ ~millions    │ ~1 (event loop) │
│ Parallel         │ ✅ Yes       │ ✅ Yes       │ ❌ No           │
│ CPU-bound        │ ✅ Good      │ ✅ Good      │ ❌ Bad          │
│ Managed by       │ OS Kernel    │ Go Runtime   │ V8 Engine       │
│ Communication    │ Shared memory│ Channels     │ Callbacks/Promises│
└──────────────────┴──────────────┴──────────────┴─────────────────┘
```

### Go Scheduler - GMP Model

Go sử dụng mô hình **GMP** để schedule goroutines:

```
GMP Model:

G = Goroutine (logic thread)
M = Machine (OS thread - thực sự chạy code)
P = Processor (context để chạy G, số lượng = GOMAXPROCS)

┌─────────────────────────────────────────────────────┐
│                   Go Runtime                         │
│                                                     │
│  ┌──────┐    ┌──────┐    ┌──────┐                  │
│  │  P0  │    │  P1  │    │  P2  │  (Processors)    │
│  │┌────┐│    │┌────┐│    │┌────┐│                   │
│  ││ G1 ││    ││ G4 ││    ││ G7 ││  Local Run Queue │
│  ││ G2 ││    ││ G5 ││    ││ G8 ││                   │
│  ││ G3 ││    ││ G6 ││    ││ G9 ││                   │
│  │└────┘│    │└────┘│    │└────┘│                   │
│  └──┬───┘    └──┬───┘    └──┬───┘                  │
│     │           │           │                       │
│  ┌──▼───┐    ┌──▼───┐    ┌──▼───┐                  │
│  │  M0  │    │  M1  │    │  M2  │  (OS Threads)    │
│  └──────┘    └──────┘    └──────┘                  │
│                                                     │
│  ┌──────────────────────────────────┐               │
│  │     Global Run Queue              │               │
│  │  [G10, G11, G12, ...]            │               │
│  └──────────────────────────────────┘               │
└─────────────────────────────────────────────────────┘

Cách hoạt động:
1. Mỗi P có local queue (lock-free, nhanh)
2. Khi local queue đầy → đẩy一半 sang global queue
3. Khi local queue rỗng → lấy từ global queue hoặc steal từ P khác
4. Khi G blocked (I/O, syscall) → M detach khỏi P, P gắn vào M khác
```

```go
// Kiểm tra số P (processor) hiện tại
fmt.Println(runtime.GOMAXPROCS(0)) // Lấy giá trị hiện tại
runtime.GOMAXPROCS(4)              // Đặt số P = 4
```

### Stack Growth và Shrinking

```go
// Goroutine stack bắt đầu nhỏ (2-8KB) và tự động grow khi cần
func recursive(n int) {
    if n <= 0 {
        return
    }
    // Stack tự động grow khi recursion sâu
    recursive(n - 1)
}

// Stack cũng tự động shrink khi không cần nữa
// → Tiết kiệm memory cho hàng triệu goroutine
```

### Cost Comparison chi tiết

```go
// Benchmark: tạo 100,000 goroutines
func BenchmarkGoroutines(b *testing.B) {
    for i := 0; i < b.N; i++ {
        var wg sync.WaitGroup
        for j := 0; j < 100000; j++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                // Minimal work
            }()
        }
        wg.Wait()
    }
}
// Kết quả: ~0.3μs per goroutine creation
// Memory: ~200MB cho 100K goroutines (2KB each)
```

```javascript
// Node.js equivalent - KHÔNG THỂ làm tương tự
// 100K Promises = 100K objects trong heap
// Nhưng vẫn chỉ 1 thread xử lý

// Nếu tạo 100K Worker Threads → hết memory ngay (~100GB)
```

### Khi nào goroutine được tạo?

```go
// 1. Explicit: dùng `go` keyword
go myFunction()

// 2. HTTP Server - mỗi request 1 goroutine (tự động)
http.ListenAndServe(":8080", handler)
// Go HTTP server tạo goroutine mới cho mỗi connection

// 3. go keyword trong loop
for i := 0; i < 10; i++ {
    go func(n int) {
        fmt.Println(n)
    }(i) // Truyền i vào, không capture
}
```

### Common Mistake: Goroutine Leaks

```go
// ❌ BAD: Goroutine leak - goroutine永远不会完成
func leakyFunction() {
    ch := make(chan int)
    go func() {
        val := 42
        ch <- val // Goroutine bị block mãi vì không ai đọc
    }()
    // Function return mà không đọc ch
    // Goroutine bị leak!
}

// ✅ GOOD: Dùng context hoặc done channel để cancel
func cleanFunction(ctx context.Context) {
    ch := make(chan int)
    go func() {
        val := 42
        select {
        case ch <- val:
        case <-ctx.Done():
            return // Clean exit
        }
    }()

    select {
    case v := <-ch:
        fmt.Println(v)
    case <-ctx.Done():
        return
    }
}
```

```javascript
// Node.js equivalent - Event loop tự xử lý
// Không có goroutine leak, nhưng có "listener leak"
// EventEmitter có maxListeners warning
```

### Detect Goroutine Leaks

```go
// Dùng runtime.NumGoroutine() để kiểm tra
func main() {
    fmt.Println("Before:", runtime.NumGoroutine()) // ~2

    for i := 0; i < 1000; i++ {
        go func() {
            time.Sleep(time.Hour) // Leak!
        }()
    }

    fmt.Println("After:", runtime.NumGoroutine()) // ~1002
    // Nếu số này tăng liên tục → có leak
}

// Dùng tool: go test -run TestLeak -race
// Hoặc dùng package: go.uber.org/goleak
```

---

## 4. Channels - Hướng dẫn hoàn chỉnh

Channel là cách Go goroutines giao tiếp với nhau. Đây là **trung tâm** của Go concurrency.

### Unbuffered Channel (đồng bộ)

```go
// Tạo unbuffered channel
ch := make(chan string)

// Unbuffered = SENDER và RECEIVER phải gặp nhau cùng lúc
// Giống như đưa đồ trực tiếp từ tay sang tay

go func() {
    time.Sleep(time.Second)
    ch <- "hello"  // Sender BLOCK cho đến khi có receiver
}()

msg := <-ch  // Receiver BLOCK cho đến khi có sender
fmt.Println(msg)
```

```
Unbuffered Channel Flow:

Sender                    Channel                    Receiver
  │                         │                           │
  │──── ch <- "hello" ─────►│                           │
  │    (BLOCKED)            │    (waiting)              │
  │                         │                           │
  │                         │◄──── <-ch ────────────────│
  │    (unblocked)          │    (data transferred)     │
  │                         │                           │
```

```javascript
// Node.js equivalent - không có concept tương đương
// Phải dùng callback hoặc Promise

// Cách 1: Callback
function sendToChannel(data, callback) {
    setTimeout(() => callback(data), 1000);
}

// Cách 2: Promise
const promise = new Promise((resolve) => {
    setTimeout(() => resolve("hello"), 1000);
});
const msg = await promise;
```

### Buffered Channel (bất đồng bộ)

```go
// Tạo buffered channel với capacity = 3
ch := make(chan int, 3)

// Có thể gửi 3 giá trị mà không block
ch <- 1  // Không block (slot 1/3)
ch <- 2  // Không block (slot 2/3)
ch <- 3  // Không block (slot 3/3)
// ch <- 4  // BLOCK! Buffer đã đầy

fmt.Println(<-ch) // 1 (slot 2/3)
fmt.Println(<-ch) // 2 (slot 1/3)
fmt.Println(<-ch) // 3 (slot 0/3)
```

```
Buffered Channel (capacity = 3):

Sender          Buffer              Receiver
  │         ┌───┬───┬───┐            │
  │── 1 ───►│ 1 │   │   │            │
  │── 2 ───►│ 1 │ 2 │   │            │
  │── 3 ───►│ 1 │ 2 │ 3 │            │
  │         └───┴───┴───┘            │
  │                                  │
  │         ┌───┬───┬───┐            │
  │         │   │ 2 │ 3 │◄── 1 ─────│
  │         └───┴───┴───┘            │
```

```javascript
// Node.js equivalent - tương tự như queue/buffer
// Không có built-in buffered channel

// Có thể dùng array như queue
const buffer = [];
const MAX_SIZE = 3;

function send(data) {
    if (buffer.length >= MAX_SIZE) {
        throw new Error("Buffer full");
    }
    buffer.push(data);
}

function receive() {
    if (buffer.length === 0) {
        throw new Error("Buffer empty");
    }
    return buffer.shift();
}
```

### Channel Direction (kiểu direction)

```go
// Bidirectional channel (default)
ch := make(chan int)

// Send-only channel (chỉ gửi)
func producer(ch chan<- int) {
    ch <- 42
    // <-ch // Compile error! Không thể receive từ send-only
}

// Receive-only channel (chỉ nhận)
func consumer(ch <-chan int) {
    msg := <-ch
    fmt.Println(msg)
    // ch <- 42 // Compile error! Không thể send vào receive-only
}

// Sử dụng:
ch := make(chan int)
go producer(ch)  // ch implicitly converted to chan<- int
go consumer(ch)  // ch implicitly converted to <-chan int
```

```
Channel Direction:

chan int     = bidirectional (gửi và nhận đều được)
chan<- int   = send-only (chỉ gửi vào)
<-chan int   = receive-only (chỉ nhận ra)

Use case:
- Producer chỉ cần send → chan<-
- Consumer chỉ cần receive → <-chan
- Prevents accidental send/receive ở sai vị trí
```

### Closing Channels

```go
ch := make(chan int, 5)

// Gửi dữ liệu
ch <- 1
ch <- 2
ch <- 3

// Đóng channel
close(ch)

// Cách 1: Comma-ok pattern để kiểm tra channel đóng chưa
val, ok := <-ch
fmt.Println(val, ok) // 1 true

// Đọc đến khi empty
for {
    val, ok := <-ch
    if !ok {
        fmt.Println("Channel closed")
        break
    }
    fmt.Println(val)
}

// Cách 2: range loop (tự động dừng khi channel đóng)
ch2 := make(chan int, 3)
ch2 <- 10
ch2 <- 20
ch2 <- 30
close(ch2)

for val := range ch2 {
    fmt.Println(val) // 10, 20, 30
}
// Loop tự động kết thúc khi channel đóng và empty
```

```
Close Channel Rules:

✅ Sender ĐƯỢC phép close channel
❌ Receiver KHÔNG NÊN close channel (có thể gửi vào closed channel → panic)
❌ Không close channel 2 lần → panic
✅ close(nil channel) → panic
✅ Send vào closed channel → panic
✅ Receive từ closed channel → zero value, ok=false
```

### nil Channel Behavior

```go
var ch chan int // nil channel

// Gửi vào nil channel → BLOCK forever
// ch <- 1 // deadlock!

// Nhận từ nil channel → BLOCK forever
// <-ch // deadlock!

// Close nil channel → PANIC
// close(ch) // panic!

// Khi nào nil channel hữu ích?
// → Trong select, có thể disable case bằng cách set channel = nil
```

```go
// nil channel trong select - disable case
func merge(channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    merged := make(chan int)

    wg.Add(len(channels))
    for _, ch := range channels {
        go func(c <-chan int) {
            defer wg.Done()
            for val := range c {
                merged <- val
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(merged)
    }()

    return merged
}
```

### Channel as Synchronization Primitive

```go
// Unbuffered channel dùng như synchronization point
done := make(chan struct{})

go func() {
    // Do work
    fmt.Println("Working...")
    time.Sleep(time.Second)
    done <- struct{}{} // Signal completion
}()

<-done // Wait for completion
fmt.Println("Done!")
```

```javascript
// Node.js equivalent
// Dùng Promise hoặc EventEmitter
const done = new Promise((resolve) => {
    setTimeout(() => {
        console.log("Working...");
        resolve();
    }, 1000);
});
await done;
console.log("Done!");
```

### select Statement (Multi-way Multiplexing)

```go
// select giống như switch, nhưng cho channels
ch1 := make(chan string)
ch2 := make(chan string)

go func() {
    time.Sleep(1 * time.Second)
    ch1 <- "one"
}()

go func() {
    time.Sleep(2 * time.Second)
    ch2 <- "two"
}()

// Chờ channel nào sẵn sàng trước
select {
case msg := <-ch1:
    fmt.Println("Received from ch1:", msg)
case msg := <-ch2:
    fmt.Println("Received từ ch2:", msg)
}
// Output: "Received from ch1: one" (vì ch1 sẵn sàng trước)
```

```
select Behavior:

1. Đánh giá TẤT CẢ channel expressions
2. Nếu có channel sẵn sàng → chọn ngẫu nhiên 1 case
3. Nếu không channel nào sẵn sàng → BLOCK chờ
4. Nếu có default → chạy default (non-blocking)

Quan trọng: select KHÔNG có fallthrough
```

### select with Default (Non-blocking)

```go
ch := make(chan int, 1)

select {
case val := <-ch:
    fmt.Println("Received:", val)
default:
    fmt.Println("No data available") // Chạy ngay nếu channel empty
}

// Tương tự với send
select {
case ch <- 42:
    fmt.Println("Sent")
default:
    fmt.Println("Channel full, skip") // Chạy ngay nếu channel full
}
```

```javascript
// Node.js equivalent - không có select
// Phải dùng event-driven hoặc polling

// Polling pattern (không efficient)
function checkChannels() {
    if (hasData(ch1)) {
        handleCh1();
    } else if (hasData(ch2)) {
        handleCh2();
    } else {
        // no data
    }
}
```

### select with Timeout

```go
ch := make(chan string)

go func() {
    time.Sleep(3 * time.Second)
    ch <- "result"
}()

select {
case msg := <-ch:
    fmt.Println("Received:", msg)
case <-time.After(2 * time.Second):
    fmt.Println("Timeout!")
}
// Output: "Timeout!" (vì ch chưa sẵn sàng trong 2s)
```

```javascript
// Node.js equivalent
const controller = new AbortController();
const timeout = setTimeout(() => controller.abort(), 2000);

try {
    const result = await fetchData({ signal: controller.signal });
    clearTimeout(timeout);
    console.log(result);
} catch (err) {
    if (err.name === 'AbortError') {
        console.log("Timeout!");
    }
}
```

### Fan-out, Fan-in Pattern

```go
// Fan-out: 1 input → nhiều workers
// Fan-in: nhiều workers → 1 output

func fanOutFanIn(input <-chan int, numWorkers int) <-chan int {
    // Fan-out: tạo nhiều workers
    workers := make([]<-chan int, numWorkers)
    for i := 0; i < numWorkers; i++ {
        workers[i] = worker(input)
    }

    // Fan-in: merge tất cả workers
    return merge(workers...)
}

func worker(input <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for val := range input {
            out <- val * 2 // Process
        }
    }()
    return out
}

func merge(channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    merged := make(chan int)

    wg.Add(len(channels))
    for _, ch := range channels {
        go func(c <-chan int) {
            defer wg.Done()
            for val := range c {
                merged <- val
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(merged)
    }()

    return merged
}
```

```
Fan-out / Fan-in Diagram:

Input ──┬──► Worker1 ──┬──
        ├──► Worker2 ──┼──► Merge ──► Output
        ├──► Worker3 ──┤
        └──► Worker4 ──┘

Fan-out: 1 source → N workers (parallel processing)
Fan-in:  N workers → 1 destination (collect results)
```

### Pipeline Pattern

```go
// Pipeline: nhiều stage nối tiếp nhau
// Stage 1 → Stage 2 → Stage 3 → Output

func pipeline() {
    // Stage 1: Generate numbers
    nums := generate(1, 2, 3, 4, 5)

    // Stage 2: Square them
    squared := square(nums)

    // Stage 3: Double them
    doubled := double(squared)

    // Output
    for val := range doubled {
        fmt.Println(val) // 2, 8, 18, 32, 50
    }
}

func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n
        }
    }()
    return out
}

func double(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * 2
        }
    }()
    return out
}
```

```
Pipeline Flow:

generate → [1,2,3,4,5] → square → [1,4,9,16,25] → double → [2,8,18,32,50]

Mỗi stage chạy trong goroutine riêng
Dữ liệu chảy qua channels
```

```javascript
// Node.js equivalent - dùng streams hoặc chaining
const { pipeline } = require('stream');

// Stream-based pipeline
const readable = createReadStream('input.txt');
const transform1 = new Transform({ transform(chunk, enc, cb) { /* ... */ } });
const transform2 = new Transform({ transform(chunk, enc, cb) { /* ... */ } });
const writable = createWriteStream('output.txt');

pipeline(readable, transform1, transform2, writable, (err) => {
    if (err) console.error('Pipeline failed');
});
```

### Worker Pool Pattern

```go
func workerPool(numWorkers int, jobs <-chan int) <-chan int {
    results := make(chan int)

    // Tạo N workers
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for job := range jobs {
                fmt.Printf("Worker %d processing job %d\n", workerID, job)
                time.Sleep(time.Second) // Simulate work
                results <- job * 2
            }
        }(i)
    }

    // Close results channel khi tất cả workers xong
    go func() {
        wg.Wait()
        close(results)
    }()

    return results
}

func main() {
    jobs := make(chan int, 100)
    results := workerPool(5, jobs)

    // Gửi jobs
    go func() {
        for i := 0; i < 20; i++ {
            jobs <- i
        }
        close(jobs)
    }()

    // Nhận results
    for result := range results {
        fmt.Println("Result:", result)
    }
}
```

```
Worker Pool:

          Jobs Channel
              │
    ┌─────────┼─────────┐
    │         │         │
    ▼         ▼         ▼
 Worker1   Worker2   Worker3
    │         │         │
    └─────────┼─────────┘
              │
       Results Channel
```

```javascript
// Node.js Worker Pool - cồng kềnh hơn nhiều
const { Worker } = require('worker_threads');

class WorkerPool {
    constructor(numWorkers, workerScript) {
        this.workers = [];
        this.freeWorkers = [];
        this.taskQueue = [];

        for (let i = 0; i < numWorkers; i++) {
            const worker = new Worker(workerScript);
            this.workers.push(worker);
            this.freeWorkers.push(worker);
        }
    }

    runTask(task) {
        return new Promise((resolve, reject) => {
            if (this.freeWorkers.length > 0) {
                const worker = this.freeWorkers.pop();
                this._run(worker, task, resolve, reject);
            } else {
                this.taskQueue.push({ task, resolve, reject });
            }
        });
    }

    _run(worker, task, resolve, reject) {
        worker.postMessage(task);
        worker.once('message', (result) => {
            resolve(result);
            this.freeWorkers.push(worker);
            if (this.taskQueue.length > 0) {
                const { task, resolve, reject } = this.taskQueue.shift();
                this._run(worker, task, resolve, reject);
            }
        });
        worker.once('error', reject);
    }
}
```

### Quit Channel Pattern

```go
func worker(quit <-chan struct{}) {
    for {
        select {
        case <-quit:
            fmt.Println("Worker shutting down")
            return
        default:
            fmt.Println("Working...")
            time.Sleep(500 * time.Millisecond)
        }
    }
}

func main() {
    quit := make(chan struct{})

    go worker(quit)

    time.Sleep(2 * time.Second)
    close(quit) // Signal worker to stop

    time.Sleep(time.Second) // Wait for worker to finish
}
```

### Done Channel Pattern

```go
func worker(done chan<- struct{}) {
    // Do work
    fmt.Println("Working...")
    time.Sleep(time.Second)
    done <- struct{}{} // Signal completion
}

func main() {
    done := make(chan struct{})

    go worker(done)

    <-done // Wait for completion
    fmt.Println("Worker done!")
}
```

---

## 5. sync Package - Hoàn chỉnh

### sync.Mutex và sync.RWMutex

```go
// Mutex (Mutual Exclusion) - lock khi ghi/đọc shared state
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Get() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}

// RWMutex - cho phép nhiều readers, 1 writer
type SafeCache struct {
    mu    sync.RWMutex
    data  map[string]string
}

func (c *SafeCache) Get(key string) (string, bool) {
    c.mu.RLock()         // Read lock - nhiều goroutine có thể đọc cùng lúc
    defer c.mu.RUnlock()
    val, ok := c.data[key]
    return val, ok
}

func (c *SafeCache) Set(key, value string) {
    c.mu.Lock()          // Write lock - exclusive access
    defer c.mu.Unlock()
    c.data[key] = value
}
```

```
Mutex vs RWMutex:

sync.Mutex:
  - Lock()    → exclusive access (không goroutine khác được truy cập)
  - Unlock()  → release lock

sync.RWMutex:
  - RLock()    → read lock (nhiều goroutine có thể đọc cùng lúc)
  - RUnlock()  → release read lock
  - Lock()     → write lock (exclusive, không ai đọc/ghi được)
  - Unlock()   → release write lock

Khi nào dùng RWMutex?
→ Khi read nhiều hơn write rất nhiều lần
→ Ví dụ: config cache, session store
```

```javascript
// Node.js equivalent - KHÔNG CẦN mutex
// Node.js single-threaded nên không có race condition cho synchronous code
// Nhưng async operations vẫn có thể có race condition

// Ví dụ race condition trong Node.js:
let counter = 0;
async function increment() {
    const val = counter;    // Read
    await delay(1);         // Context switch (async)
    counter = val + 1;      // Write → RACE CONDITION!
}

// Node.js solution: dùng lock library hoặc serialize operations
const { Mutex } = require('async-mutex');
const mutex = new Mutex();

async function safeIncrement() {
    const release = await mutex.acquire();
    try {
        counter++;
    } finally {
        release();
    }
}
```

### sync.WaitGroup

```go
var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)  // Tăng counter
    go func(id int) {
        defer wg.Done()  // Giảm counter khi xong
        fmt.Printf("Worker %d done\n", id)
        time.Sleep(time.Second)
    }(i)
}

wg.Wait()  // Block cho đến khi counter = 0
fmt.Println("All workers done!")
```

```javascript
// Node.js equivalent
const promises = [];
for (let i = 0; i < 5; i++) {
    promises.push(
        (async (id) => {
            console.log(`Worker ${id} done`);
            await delay(1000);
        })(i)
    );
}
await Promise.all(promises);
console.log("All workers done!");
```

### sync.Once

```go
var once sync.Once
var config *Config

func GetConfig() *Config {
    once.Do(func() {
        // Chỉ chạy 1 lần duy nhất, thread-safe
        config = loadConfig()
        fmt.Println("Config loaded!")
    })
    return config
}

// Gọi nhiều lần từ nhiều goroutine:
// → Config chỉ load 1 lần
// → Các goroutine khác đợi cho đến khi load xong
```

```javascript
// Node.js equivalent - module caching
// Node.js modules chỉ load 1 lần (built-in behavior)
const config = require('./config'); // Cached after first require

// Hoặc dùng singleton pattern
class Config {
    static instance = null;
    static getInstance() {
        if (!this.instance) {
            this.instance = new Config();
        }
        return this.instance;
    }
}
```

### sync.Map

```go
// sync.Map - concurrent-safe map
// Dùng khi: nhiều goroutine đọc/ghi, keys ít thay đổi
var m sync.Map

// Store
m.Store("key1", "value1")
m.Store("key2", "value2")

// Load
val, ok := m.Load("key1")
if ok {
    fmt.Println(val) // "value1"
}

// Load or Store (atomic)
val, loaded := m.LoadOrStore("key3", "value3")
// loaded = false vì key3 chưa tồn tại

// Delete
m.Delete("key1")

// Range (iterate)
m.Range(func(key, value interface{}) bool {
    fmt.Printf("%v: %v\n", key, value)
    return true // tiếp tục iteration
})
```

```
sync.Map vs map + mutex:

sync.Map:
  ✅ Optimized cho 2 cases:
    1. Entry chỉ viết 1 lần nhưng đọc nhiều lần (read-heavy)
    2. Nhiều goroutine đọc/ghi disjoint sets of keys
  ❌ Không tối ưu cho cases khác
  ❌ Không có len() - không biết bao nhiêu entries

map + mutex:
  ✅ Linh hoạt hơn
  ✅ Có thể dùng len()
  ✅ Tốt cho balanced read/write
  ❌ Phải tự quản lý lock
```

### sync.Pool

```go
// sync.Pool - object reuse pool
// Giảm garbage collection pressure

var bufPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func process(data []byte) {
    // Lấy buffer từ pool (hoặc tạo mới nếu pool empty)
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()        // Clear buffer
        bufPool.Put(buf)   // Trả lại pool
    }()

    buf.Write(data)
    // Use buf...
}
```

```javascript
// Node.js equivalent - object pooling
class Pool {
    constructor(factory, maxSize = 10) {
        this.factory = factory;
        this.pool = [];
        this.maxSize = maxSize;
    }

    acquire() {
        if (this.pool.length > 0) {
            return this.pool.pop();
        }
        return this.factory();
    }

    release(obj) {
        if (this.pool.length < this.maxSize) {
            this.pool.push(obj);
        }
    }
}

const bufferPool = new Pool(() => Buffer.alloc(1024));
```

### sync.Cond

```go
// sync.Cond - condition variable
// Dùng để signal/broadcast giữa goroutines

type BoundedQueue struct {
    mu       sync.Mutex
    cond     *sync.Cond
    items    []interface{}
    maxSize  int
}

func NewBoundedQueue(maxSize int) *BoundedQueue {
    q := &BoundedQueue{maxSize: maxSize}
    q.cond = sync.NewCond(&q.mu)
    return q
}

func (q *BoundedQueue) Put(item interface{}) {
    q.mu.Lock()
    defer q.mu.Unlock()

    // Đợi cho đến khi có chỗ
    for len(q.items) >= q.maxSize {
        q.cond.Wait() // Release lock và đợi signal
    }

    q.items = append(q.items, item)
    q.cond.Signal() // Signal 1 waiter
}

func (q *BoundedQueue) Get() interface{} {
    q.mu.Lock()
    defer q.mu.Unlock()

    // Đợi cho đến khi có item
    for len(q.items) == 0 {
        q.cond.Wait()
    }

    item := q.items[0]
    q.items = q.items[1:]
    q.cond.Signal()
    return item
}
```

```
sync.Cond Methods:

Wait()    → Release lock, đợi signal, re-acquire lock
Signal()  → Wake up 1 waiter
Broadcast() → Wake up ALL waiters

Lưu ý: Luôn dùng Wait() trong loop (for), không dùng if
→ Vì spurious wakeup (có thể wakeup mà không có signal)
```

### Atomic Operations (sync/atomic)

```go
import "sync/atomic"

var counter int64

// Atomic increment - thread-safe, không cần mutex
atomic.AddInt64(&counter, 1)

// Atomic load
val := atomic.LoadInt64(&counter)

// Atomic store
atomic.StoreInt64(&counter, 42)

// Compare and Swap (CAS)
swapped := atomic.CompareAndSwapInt64(&counter, 42, 100)
// swapped = true nếu counter == 42, đổi thành 100
// swapped = false nếu counter != 42

// Atomic Value (any type)
var config atomic.Value
config.Store(map[string]string{"key": "value"})
current := config.Load().(map[string]string)
```

```
Atomic vs Mutex:

Atomic:
  ✅ Nhanh hơn (lock-free)
  ✅ Chỉ cho phép operations đơn giản (add, load, store, CAS)
  ❌ Không cho phép complex operations

Mutex:
  ✅ Cho phép bất kỳ code nào trong critical section
  ❌ Chậm hơn (context switch khi blocked)

Khi nào dùng atomic?
→ Counter, flags, simple state changes
→ Khi operation chỉ là read-modify-write đơn giản
```

### Khi nào dùng Channels vs Mutex?

```
Channels:
  ✅ Truyền dữ liệu giữa goroutines
  ✅ Synchronization (rendezvous point)
  ✅ Ownership transfer (ai gửi channel owns data)
  ✅ Fan-out, fan-in, pipeline patterns

Mutex:
  ✅ Bảo vệ shared state (cache, counter)
  ✅ Simple locking scenarios
  ✅ Khi bạn chỉ cần protect data, không cần communicate

Rule of thumb:
  "Don't communicate by sharing memory; share memory by communicating."
  → Ưu tiên channels, dùng mutex khi channels quá cồng kềnh
```

---

## 6. Context - Deep Dive

### context.Context là gì?

Context là đối tượng mang theo:
1. **Cancellation signal** - hủy operation
2. **Deadline/Timeout** - thời gian tối đa
3. **Request-scoped values** - metadata theo request

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key interface{}) interface{}
}
```

### context.Background() và context.TODO()

```go
// context.Background() - root context, không bao giờ bị cancel
ctx := context.Background()

// context.TODO() - placeholder khi chưa biết dùng context nào
ctx := context.TODO()

// Cả 2 đều trả về empty context
// Khác nhau về ngữ nghĩa:
// Background: dùng cho main, init, tests
// TODO: dùng khi bạn chưa implement context propagation
```

### context.WithCancel

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())

    go worker(ctx, "worker-1")
    go worker(ctx, "worker-2")

    time.Sleep(2 * time.Second)
    cancel() // Cancel tất cả workers

    time.Sleep(time.Second) // Đợi workers cleanup
}

func worker(ctx context.Context, name string) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("%s: cancelled\n", name)
            return
        default:
            fmt.Printf("%s: working\n", name)
            time.Sleep(500 * time.Millisecond)
        }
    }
}
```

```
Cancel Propagation:

context.Background()
    │
    ▼
ctx, cancel := context.WithCancel(parent)
    │
    ├──► worker-1 (đợi ctx.Done())
    │
    └──► worker-2 (đợi ctx.Done())

cancel() được gọi → ctx.Done() đóng → tất cả workers nhận signal
```

```javascript
// Node.js equivalent - AbortController
const controller = new AbortController();
const { signal } = controller;

async function worker(name, signal) {
    while (!signal.aborted) {
        console.log(`${name}: working`);
        await delay(500);
    }
    console.log(`${name}: cancelled`);
}

worker('worker-1', signal);
worker('worker-2', signal);

setTimeout(() => controller.abort(), 2000);
```

### context.WithTimeout

```go
func fetchData(ctx context.Context) (string, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel() // Luôn gọi defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)
    if err != nil {
        return "", err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(body), nil
}
```

```javascript
// Node.js equivalent
async function fetchData(signal) {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 5000);

    try {
        const resp = await fetch('https://api.example.com/data', {
            signal: controller.signal
        });
        clearTimeout(timeout);
        return await resp.text();
    } catch (err) {
        clearTimeout(timeout);
        throw err;
    }
}
```

### context.WithDeadline

```go
// Deadline = thời điểm cụ thể (absolute time)
// Timeout = khoảng thời gian (relative)

deadline := time.Now().Add(5 * time.Second)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()

// Tương đương với:
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

### context.WithValue (và khi KHÔNG nên dùng)

```go
// context.WithValue - gắn metadata vào context
ctx := context.WithValue(context.Background(), "userID", "12345")
ctx = context.WithValue(ctx, "requestID", "abc-123")

// Lấy value
userID := ctx.Value("userID").(string)
requestID := ctx.Value("requestID").(string)
```

```
⚠️ Khi KHÔNG nên dùng context.WithValue:

❌ Truyền required parameters (dùng function arguments)
❌ Truyền optional parameters (dùng struct options)
❌ Truyền data mà bất kỳ function nào cũng cần

✅ NÊN dùng cho:
  - Request ID, Trace ID
  - Authentication token
  - User ID (nếu cần ở nhiều middleware)
  - Deadline/timeout metadata

Rule: Values phải là request-scoped và immutable
```

### Context Propagation trong HTTP Handlers

```go
func main() {
    http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
        // r.Context() - context từ HTTP request
        // Tự động cancel khi client disconnect
        ctx := r.Context()

        // Thêm metadata
        ctx = context.WithValue(ctx, "requestID", generateRequestID())

        // Truyền context xuống các layers
        data, err := fetchFromDB(ctx)
        if err != nil {
            if ctx.Err() == context.Canceled {
                fmt.Println("Client disconnected")
                return
            }
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        json.NewEncoder(w).Encode(data)
    })

    http.ListenAndServe(":8080", nil)
}

func fetchFromDB(ctx context.Context) ([]Data, error) {
    // Truyền context vào database query
    rows, err := db.QueryContext(ctx, "SELECT * FROM data")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    // ... process rows
}
```

```javascript
// Node.js equivalent - không có built-in context propagation
// Phải tự implement hoặc dùng library

// Express middleware pattern
app.use((req, res, next) => {
    req.requestId = generateRequestId();
    next();
});

app.get('/api/data', async (req, res) => {
    try {
        const data = await fetchFromDB(req.requestId);
        res.json(data);
    } catch (err) {
        if (err.name === 'AbortError') {
            console.log('Client disconnected');
            return;
        }
        res.status(500).json({ error: err.message });
    }
});
```

### Context trong Database Queries

```go
func getUser(ctx context.Context, userID string) (*User, error) {
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    query := "SELECT id, name, email FROM users WHERE id = ?"
    row := db.QueryRowContext(ctx, query, userID)

    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return nil, fmt.Errorf("query timeout for user %s", userID)
        }
        return nil, err
    }

    return &user, nil
}
```

### Context trong Goroutines

```go
func processRequest(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    errCh := make(chan error, 1)

    go func() {
        // Goroutine sử dụng context
        result, err := heavyComputation(ctx)
        if err != nil {
            errCh <- err
            return
        }
        // Process result...
        errCh <- nil
    }()

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        return ctx.Err() // Timeout hoặc cancelled
    }
}
```

### Cancellation Propagation Tree

```
Context Propagation Tree:

main context (Background)
    │
    ├── HTTP Request 1 context
    │       │
    │       ├── DB Query context (WithTimeout 3s)
    │       │
    │       └── External API context (WithTimeout 5s)
    │
    ├── HTTP Request 2 context
    │       │
    │       └── ...
    │
    └── Background Job context (WithCancel)
            │
            ├── Worker 1 context
            │
            └── Worker 2 context

Khi parent context bị cancel → tất cả child contexts cũng bị cancel
```

### Context vs AbortController (Node.js)

```
┌─────────────────────┬─────────────────────┬─────────────────────┐
│ Feature             │ Go context.Context  │ Node.js AbortController│
├─────────────────────┼─────────────────────┼─────────────────────┤
│ Cancellation        │ ✅ Built-in         │ ✅ Built-in          │
│ Timeout             │ ✅ Built-in         │ ⚠️ Manual (setTimeout)│
│ Deadline            │ ✅ Built-in         │ ❌ Not built-in      │
│ Value propagation   │ ✅ Built-in         │ ❌ Not built-in      │
│ Hierarchical        │ ✅ Tree structure   │ ❌ Flat              │
│ Stdlib integration  │ ✅ All stdlib       │ ⚠️ Partial           │
│ API consistency     │ ✅ 1 interface      │ ❌ Different APIs    │
│ Learn once          │ ✅ Yes              │ ❌ No                │
└─────────────────────┴─────────────────────┴─────────────────────┘
```

---

## 7. Common Concurrency Patterns

### Fan-out/Fan-in Pattern

```go
// Fan-out: phân phối work cho nhiều workers
// Fan-in: kết hợp kết quả từ nhiều workers

func processURLs(urls []string) []Result {
    input := make(chan string, len(urls))
    for _, url := range urls {
        input <- url
    }
    close(input)

    // Fan-out: 10 workers
    numWorkers := 10
    workers := make([]<-chan Result, numWorkers)
    for i := 0; i < numWorkers; i++ {
        workers[i] = fetchURL(input)
    }

    // Fan-in: merge results
    return collectResults(workers)
}

func fetchURL(input <-chan string) <-chan Result {
    out := make(chan Result)
    go func() {
        defer close(out)
        for url := range input {
            resp, err := http.Get(url)
            out <- Result{URL: url, Response: resp, Err: err}
        }
    }()
    return out
}

func collectResults(channels []<-chan Result) []Result {
    var wg sync.WaitGroup
    merged := make(chan Result)

    wg.Add(len(channels))
    for _, ch := range channels {
        go func(c <-chan Result) {
            defer wg.Done()
            for result := range c {
                merged <- result
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(merged)
    }()

    var results []Result
    for result := range merged {
        results = append(results, result)
    }
    return results
}
```

### Pipeline Pattern

```go
// Pipeline: dữ liệu chảy qua nhiều stages
// Mỗi stage là 1 goroutine, connected bằng channels

type Stage func(<-chan interface{}) <-chan interface{}

func pipeline(stages ...Stage) Stage {
    return func(input <-chan interface{}) <-chan interface{} {
        current := input
        for _, stage := range stages {
            current = stage(current)
        }
        return current
    }
}

// Stage implementations
func filterStage(predicate func(interface{}) bool) Stage {
    return func(input <-chan interface{}) <-chan interface{} {
        output := make(chan interface{})
        go func() {
            defer close(output)
            for item := range input {
                if predicate(item) {
                    output <- item
                }
            }
        }()
        return output
    }
}

func mapStage(transform func(interface{}) interface{}) Stage {
    return func(input <-chan interface{}) <-chan interface{} {
        output := make(chan interface{})
        go func() {
            defer close(output)
            for item := range input {
                output <- transform(item)
            }
        }()
        return output
    }
}

// Usage
func main() {
    input := make(chan interface{})

    // Build pipeline
    process := pipeline(
        filterStage(func(item interface{}) bool {
            return item.(int)%2 == 0 // Chỉ lấy số chẵn
        }),
        mapStage(func(item interface{}) interface{} {
            return item.(int) * item.(int) // Bình phương
        }),
    )

    // Run pipeline
    go func() {
        for i := 0; i < 10; i++ {
            input <- i
        }
        close(input)
    }()

    for result := range process(input) {
        fmt.Println(result) // 0, 4, 16, 36, 64
    }
}
```

### Worker Pool Pattern (detailed)

```go
type Pool struct {
    jobs    chan Job
    results chan Result
    quit    chan struct{}
    wg      sync.WaitGroup
}

type Job struct {
    ID   int
    Data interface{}
}

type Result struct {
    JobID  int
    Output interface{}
    Err    error
}

func NewPool(numWorkers int, bufferSize int) *Pool {
    p := &Pool{
        jobs:    make(chan Job, bufferSize),
        results: make(chan Result, bufferSize),
        quit:    make(chan struct{}),
    }

    // Start workers
    p.wg.Add(numWorkers)
    for i := 0; i < numWorkers; i++ {
        go p.worker(i)
    }

    return p
}

func (p *Pool) worker(id int) {
    defer p.wg.Done()
    for {
        select {
        case job, ok := <-p.jobs:
            if !ok {
                return // Channel closed
            }
            result := p.process(job)
            p.results <- result
        case <-p.quit:
            return
        }
    }
}

func (p *Pool) process(job Job) Result {
    // Process job
    output := fmt.Sprintf("Processed job %d", job.ID)
    return Result{JobID: job.ID, Output: output}
}

func (p *Pool) Submit(job Job) {
    p.jobs <- job
}

func (p *Pool) Results() <-chan Result {
    return p.results
}

func (p *Pool) Shutdown() {
    close(p.jobs)
    p.wg.Wait()
    close(p.results)
}
```

### Rate Limiting Pattern

```go
// Rate limiter sử dụng ticker
func rateLimitedWorker(requests <-chan Request, rate int) {
    ticker := time.NewTicker(time.Second / time.Duration(rate))
    defer ticker.Stop()

    for req := range requests {
        <-ticker.C // Đợi cho đến khi tick
        process(req)
    }
}

// Token bucket rate limiter
type RateLimiter struct {
    ticker   *time.Ticker
    tokens   chan struct{}
}

func NewRateLimiter(rate int) *RateLimiter {
    rl := &RateLimiter{
        ticker: time.NewTicker(time.Second / time.Duration(rate)),
        tokens: make(chan struct{}, rate),
    }

    // Fill tokens
    go func() {
        for range rl.ticker.C {
            select {
            case rl.tokens <- struct{}{}:
            default:
            }
        }
    }()

    return rl
}

func (rl *RateLimiter) Wait() {
    <-rl.tokens
}
```

```javascript
// Node.js rate limiting
class RateLimiter {
    constructor(rate) {
        this.rate = rate;
        this.tokens = rate;
        this.lastRefill = Date.now();
    }

    async wait() {
        this.refill();
        if (this.tokens <= 0) {
            await new Promise(resolve => setTimeout(resolve, 1000 / this.rate));
            this.refill();
        }
        this.tokens--;
    }

    refill() {
        const now = Date.now();
        const elapsed = now - this.lastRefill;
        this.tokens = Math.min(this.rate, this.tokens + elapsed * this.rate / 1000);
        this.lastRefill = now;
    }
}
```

### Debounce Pattern

```go
func debounce(interval time.Duration, f func()) func() {
    var timer *time.Timer
    var mu sync.Mutex

    return func() {
        mu.Lock()
        defer mu.Unlock()

        if timer != nil {
            timer.Stop()
        }

        timer = time.AfterFunc(interval, f)
    }
}

// Usage
func main() {
    search := debounce(300*time.Millisecond, func() {
        fmt.Println("Searching...")
    })

    // Gọi nhiều lần, chỉ execute sau 300ms không gọi nữa
    search()
    search()
    search()
    time.Sleep(time.Second) // "Searching..." in ra 1 lần
}
```

```javascript
// Node.js debounce
function debounce(fn, delay) {
    let timer;
    return function (...args) {
        clearTimeout(timer);
        timer = setTimeout(() => fn.apply(this, args), delay);
    };
}

const search = debounce(() => console.log('Searching...'), 300);
search();
search();
search();
// "Searching..." in ra 1 lần sau 300ms
```

### Timeout Pattern

```go
func withTimeout(timeout time.Duration, f func() error) error {
    done := make(chan error, 1)

    go func() {
        done <- f()
    }()

    select {
    case err := <-done:
        return err
    case <-time.After(timeout):
        return fmt.Errorf("operation timed out after %v", timeout)
    }
}

// Usage
err := withTimeout(5*time.Second, func() error {
    return heavyComputation()
})
```

### Heartbeat Pattern

```go
func workerWithHeartbeat(ctx context.Context, heartbeatInterval time.Duration) {
    heartbeat := time.NewTicker(heartbeatInterval)
    defer heartbeat.Stop()

    for {
        select {
        case <-ctx.Done():
            fmt.Println("Worker shutting down")
            return
        case <-heartbeat.C:
            fmt.Println("Worker alive") // Heartbeat signal
        default:
            // Do work
            doWork()
        }
    }
}
```

### Graceful Shutdown Pattern

```go
func main() {
    srv := &http.Server{Addr: ":8080", Handler: handler}

    // Channel để nhận signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    // Start server trong goroutine
    go func() {
        fmt.Println("Server starting on :8080")
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Đợi signal
    <-quit
    fmt.Println("Shutting down server...")

    // Graceful shutdown với timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    fmt.Println("Server stopped")
}
```

### errGroup Pattern

```go
import "golang.org/x/sync/errgroup"

func fetchData(ctx context.Context) (*Data, error) {
    var (
        userData    *User
        orderData   *Order
        productData *Product
    )

    g, ctx := errgroup.WithContext(ctx)

    // Fetch user data
    g.Go(func() error {
        var err error
        userData, err = fetchUser(ctx)
        return err
    })

    // Fetch order data
    g.Go(func() error {
        var err error
        orderData, err = fetchOrders(ctx)
        return err
    })

    // Fetch product data
    g.Go(func() error {
        var err error
        productData, err = fetchProducts(ctx)
        return err
    })

    // Wait for all, hoặc fail fast nếu có error
    if err := g.Wait(); err != nil {
        return nil, err
    }

    return &Data{
        User:    userData,
        Orders:  orderData,
        Products: productData,
    }, nil
}
```

```javascript
// Node.js equivalent
async function fetchData() {
    try {
        const [userData, orderData, productData] = await Promise.all([
            fetchUser(),
            fetchOrders(),
            fetchProducts(),
        ]);

        return { user: userData, orders: orderData, products: productData };
    } catch (err) {
        throw err;
    }
}
```

---

## 8. Concurrency Anti-patterns và Pitfalls

### Goroutine Leaks

```go
// ❌ BAD: Goroutine leak - goroutine bị block mãi
func leakyRequest() {
    ch := make(chan Response)
    go func() {
        resp, _ := http.Get("https://api.example.com/data")
        ch <- resp // Block nếu không ai đọc
    }()
    // Function return mà không đọc ch
    // Goroutine bị leak!
}

// ✅ GOOD: Dùng buffered channel hoặc context
func cleanRequest(ctx context.Context) error {
    ch := make(chan Response, 1) // Buffered channel
    go func() {
        resp, err := http.Get("https://api.example.com/data")
        if err != nil {
            ch <- Response{Err: err}
            return
        }
        ch <- Response{Data: resp}
    }()

    select {
    case result := <-ch:
        return result.Err
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### Race Conditions

```go
// ❌ BAD: Race condition
var counter int

func increment() {
    counter++ // NOT thread-safe!
}

// Chạy 1000 goroutines cùng increment
// Kết quả có thể < 1000

// ✅ GOOD: Sử dụng mutex
var mu sync.Mutex
var safeCounter int

func safeIncrement() {
    mu.Lock()
    defer mu.Unlock()
    safeCounter++
}

// ✅ BETTER: Sử dụng atomic
var atomicCounter int64

func atomicIncrement() {
    atomic.AddInt64(&atomicCounter, 1)
}
```

```bash
# Detect race conditions
go run -race main.go
go test -race ./...

# Output khi có race:
# WARNING: DATA RACE
# Read at 0x00000064 by goroutine 8:
#   main.increment()
#       /path/to/main.go:10 +0x38
# Previous write at 0x00000064 by goroutine 7:
#   main.increment()
#       /path/to/main.go:10 +0x50
```

### Deadlocks

```go
// ❌ BAD: Deadlock - 2 goroutines chờ nhau
func deadlock() {
    ch1 := make(chan int)
    ch2 := make(chan int)

    go func() {
        <-ch1 // Đợi ch1
        ch2 <- 1
    }()

    go func() {
        <-ch2 // Đợi ch2
        ch1 <- 1
    }()

    // Cả 2 goroutine đều block → deadlock
}

// ✅ GOOD: Đảm bảo có sender/receiver
func noDeadlock() {
    ch1 := make(chan int)
    ch2 := make(chan int)

    go func() {
        ch1 <- 1 // Gửi trước
        <-ch2
    }()

    go func() {
        <-ch1
        ch2 <- 1
    }()
}
```

### For-loop Variable Capture

```go
// ❌ BAD: Capture loop variable
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i) // Có thể in 5,5,5,5,5
    }()
}
time.Sleep(time.Second)

// ✅ GOOD: Pass as parameter
for i := 0; i < 5; i++ {
    go func(n int) {
        fmt.Println(n) // In 0,1,2,3,4
    }(i)
}

// ✅ BETTER (Go 1.22+): Loop variable scope changes
// Go 1.22+ loop variables có scope per-iteration
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i) // Go 1.22+: OK
    }()
}
```

### Closing Channel từ Receiver Side

```go
// ❌ BAD: Close channel từ receiver
func badClose() {
    ch := make(chan int)
    go func() {
        ch <- 1
        ch <- 2
    }()

    fmt.Println(<-ch)
    close(ch) // Có thể close trước khi sender gửi xong → panic
    fmt.Println(<-ch)
}

// ✅ GOOD: Close từ sender side
func goodClose() {
    ch := make(chan int)
    go func() {
        ch <- 1
        ch <- 2
        close(ch) // Sender close
    }()

    for val := range ch {
        fmt.Println(val)
    }
}
```

### Sending to Closed Channel

```go
// ❌ BAD: Send vào closed channel → PANIC
func sendToClosed() {
    ch := make(chan int, 1)
    close(ch)
    ch <- 1 // panic: send on closed channel
}

// ✅ GOOD: Check channel status
func safeSend(ch chan int, val int) (sent bool) {
    defer func() {
        if r := recover(); r != nil {
            sent = false
        }
    }()

    ch <- val
    return true
}
```

### Select with nil Channel

```go
// ❌ BAD: Select với nil channel → block forever
func selectNil() {
    var ch chan int // nil channel

    select {
    case <-ch: // Block forever!
        fmt.Println("received")
    }
}

// ✅ GOOD: Use nil channel to disable select case
func dynamicSelect(channels []<-chan int) {
    for len(channels) > 0 {
        var merged <-chan int
        cases := make([]reflect.SelectCase, len(channels))
        for i, ch := range channels {
            cases[i] = reflect.SelectCase{
                Dir:  reflect.SelectRecv,
                Chan: reflect.ValueOf(ch),
            }
        }

        chosen, value, ok := reflect.Select(cases)
        if !ok {
            // Channel closed, remove from slice
            channels = append(channels[:chosen], channels[chosen+1:]...)
            continue
        }
        fmt.Printf("Received %v from channel %d\n", value.Interface(), chosen)
    }
}
```

---

## 9. Go Concurrency vs Node.js Patterns - So sánh song song

### Promise.all vs WaitGroup

```javascript
// Node.js: Promise.all
const [user, posts, comments] = await Promise.all([
    fetchUser(userId),
    fetchPosts(userId),
    fetchComments(userId),
]);
```

```go
// Go: WaitGroup
var wg sync.WaitGroup
var (
    user     *User
    posts    []Post
    comments []Comment
    errs     []error
)

wg.Add(3)
go func() { defer wg.Done(); user, errs = append(errs, fetchUser(userId)) }()
go func() { defer wg.Done(); posts, errs = append(errs, fetchPosts(userId)) }()
go func() { defer wg.Done(); comments, errs = append(errs, fetchComments(userId)) }()
wg.Wait()

// Hoặc dùng errgroup (tốt hơn)
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { var err error; user, err = fetchUser(userId); return err })
g.Go(func() error { var err error; posts, err = fetchPosts(userId); return err })
g.Go(func() error { var err error; comments, err = fetchComments(userId); return err })
if err := g.Wait(); err != nil { /* handle error */ }
```

### Promise.race vs select

```javascript
// Node.js: Promise.race
const result = await Promise.race([
    fetchFromAPI1(),
    fetchFromAPI2(),
    timeout(5000),
]);
```

```go
// Go: select
func fetchFastest(ctx context.Context) (string, error) {
    ch1 := make(chan string, 1)
    ch2 := make(chan string, 1)

    go func() { result, _ := fetchFromAPI1(ctx); ch1 <- result }()
    go func() { result, _ := fetchFromAPI2(ctx); ch2 <- result }()

    select {
    case result := <-ch1:
        return result, nil
    case result := <-ch2:
        return result, nil
    case <-ctx.Done():
        return "", ctx.Err()
    }
}
```

### async/await vs goroutine+channel

```javascript
// Node.js: async/await
async function processData() {
    const data = await fetchData();
    const processed = await process(data);
    await save(processed);
    return processed;
}
```

```go
// Go: goroutine + channel (parallel)
func processData(ctx context.Context) error {
    errCh := make(chan error, 3)

    go func() {
        data, err := fetchData(ctx)
        if err != nil {
            errCh <- err
            return
        }
        processed, err := process(ctx, data)
        if err != nil {
            errCh <- err
            return
        }
        errCh <- save(ctx, processed)
    }()

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### EventEmitter vs channel

```javascript
// Node.js: EventEmitter
const EventEmitter = require('events');
const emitter = new EventEmitter();

emitter.on('data', (data) => {
    console.log('Received:', data);
});

emitter.emit('data', 'hello');
```

```go
// Go: channel
func main() {
    dataCh := make(chan string)

    // Consumer
    go func() {
        for data := range dataCh {
            fmt.Println("Received:", data)
        }
    }()

    // Producer
    dataCh <- "hello"
    close(dataCh)
}
```

### Worker Threads vs goroutines

```javascript
// Node.js: Worker Threads
const { Worker } = require('worker_threads');

const worker = new Worker('./task.js');
worker.postMessage({ data: 'input' });
worker.on('message', (result) => {
    console.log(result);
});
worker.on('error', (err) => {
    console.error(err);
});
```

```go
// Go: goroutine + channel
func main() {
    input := make(chan string, 1)
    output := make(chan string, 1)
    errCh := make(chan error, 1)

    go func() {
        defer close(output)
        defer close(errCh)
        for data := range input {
            result, err := process(data)
            if err != nil {
                errCh <- err
                return
            }
            output <- result
        }
    }()

    input <- "input"
    close(input)

    select {
    case result := <-output:
        fmt.Println(result)
    case err := <-errCh:
        fmt.Println(err)
    }
}
```

### Cluster vs GOMAXPROCS

```javascript
// Node.js: Cluster
const cluster = require('cluster');
const numCPUs = require('os').cpus().length;

if (cluster.isPrimary) {
    for (let i = 0; i < numCPUs; i++) {
        cluster.fork();
    }
    cluster.on('exit', (worker) => {
        cluster.fork(); // Restart worker
    });
} else {
    // Worker process
    require('./server.js');
}
```

```go
// Go: Không cần cluster
// Go HTTP server tự động chạy parallel trên nhiều core
// GOMAXPROCS mặc định = số CPU cores

import "runtime"

func main() {
    // Optional: set GOMAXPROCS (mặc định đã là numCPU)
    runtime.GOMAXPROCS(runtime.NumCPU())

    http.ListenAndServe(":8080", handler)
    // Tự động spawn goroutine cho mỗi request
    // Tự động schedule trên nhiều OS threads
}
```

---

## 10. Real-world Examples

### Concurrent HTTP Client

```go
type HTTPClient struct {
    client    *http.Client
    semaphore chan struct{} // Rate limiting
}

func NewHTTPClient(maxConcurrent int) *HTTPClient {
    return &HTTPClient{
        client: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        maxConcurrent,
                MaxIdleConnsPerHost: maxConcurrent,
                IdleConnTimeout:     90 * time.Second,
            },
        },
        semaphore: make(chan struct{}, maxConcurrent),
    }
}

func (c *HTTPClient) Get(ctx context.Context, url string) ([]byte, error) {
    // Rate limiting
    select {
    case c.semaphore <- struct{}{}:
        defer func() { <-c.semaphore }()
    case <-ctx.Done():
        return nil, ctx.Err()
    }

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}

// Concurrent fetch multiple URLs
func (c *HTTPClient) FetchAll(ctx context.Context, urls []string) (map[string][]byte, error) {
    results := make(map[string][]byte)
    var mu sync.Mutex
    var wg sync.WaitGroup
    errCh := make(chan error, len(urls))

    for _, url := range urls {
        wg.Add(1)
        go func(u string) {
            defer wg.Done()
            data, err := c.Get(ctx, u)
            if err != nil {
                errCh <- err
                return
            }
            mu.Lock()
            results[u] = data
            mu.Unlock()
        }(url)
    }

    // Wait and check errors
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        close(errCh)
        for err := range errCh {
            return nil, err
        }
        return results, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

### Rate-limited API Caller

```go
type RateLimitedCaller struct {
    limiter *time.Ticker
    client  *http.Client
    mu      sync.Mutex
}

func NewRateLimitedCaller(requestsPerSecond int) *RateLimitedCaller {
    return &RateLimitedCaller{
        limiter: time.NewTicker(time.Second / time.Duration(requestsPerSecond)),
        client:  &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *RateLimitedCaller) Call(ctx context.Context, url string) ([]byte, error) {
    // Wait for rate limit token
    select {
    case <-c.limiter.C:
    case <-ctx.Done():
        return nil, ctx.Err()
    }

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}

// Usage
func main() {
    caller := NewRateLimitedCaller(10) // 10 requests/second

    urls := []string{
        "https://api.example.com/1",
        "https://api.example.com/2",
        // ... 1000 URLs
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    results := make(chan Result, len(urls))
    var wg sync.WaitGroup

    for _, url := range urls {
        wg.Add(1)
        go func(u string) {
            defer wg.Done()
            data, err := caller.Call(ctx, u)
            results <- Result{URL: u, Data: data, Err: err}
        }(url)
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    for result := range results {
        if result.Err != nil {
            fmt.Printf("Error fetching %s: %v\n", result.URL, result.Err)
            continue
        }
        fmt.Printf("Fetched %s: %d bytes\n", result.URL, len(result.Data))
    }
}
```

### Log Processor Pipeline

```go
type LogEntry struct {
    Timestamp time.Time
    Level     string
    Message   string
    Source    string
}

// Stage 1: Parse log lines
func parseStage(ctx context.Context, lines <-chan string) <-chan LogEntry {
    out := make(chan LogEntry)
    go func() {
        defer close(out)
        for line := range lines {
            select {
            case <-ctx.Done():
                return
            default:
            }

            entry, err := parseLogLine(line)
            if err != nil {
                continue // Skip invalid lines
            }
            select {
            case out <- entry:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Stage 2: Filter by level
func filterStage(ctx context.Context, entries <-chan LogEntry, level string) <-chan LogEntry {
    out := make(chan LogEntry)
    go func() {
        defer close(out)
        for entry := range entries {
            select {
            case <-ctx.Done():
                return
            default:
            }

            if entry.Level == level {
                select {
                case out <- entry:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    return out
}

// Stage 3: Aggregate
func aggregateStage(ctx context.Context, entries <-chan LogEntry) <-chan map[string]int {
    out := make(chan map[string]int)
    go func() {
        defer close(out)
        counts := make(map[string]int)
        for entry := range entries {
            select {
            case <-ctx.Done():
                return
            default:
            }
            counts[entry.Source]++
        }
        select {
        case out <- counts:
        case <-ctx.Done():
        }
    }()
    return out
}

// Run pipeline
func processLogs(ctx context.Context, logFile string) (map[string]int, error) {
    lines, err := readLines(ctx, logFile)
    if err != nil {
        return nil, err
    }

    entries := parseStage(ctx, lines)
    errors := filterStage(ctx, entries, "ERROR")
    counts := aggregateStage(ctx, errors)

    select {
    case result := <-counts:
        return result, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

### Concurrent File Processor

```go
type FileProcessor struct {
    numWorkers int
    bufferSize int
}

func NewFileProcessor(numWorkers, bufferSize int) *FileProcessor {
    return &FileProcessor{
        numWorkers: numWorkers,
        bufferSize: bufferSize,
    }
}

func (fp *FileProcessor) Process(ctx context.Context, dir string) error {
    files, err := filepath.Glob(filepath.Join(dir, "*.txt"))
    if err != nil {
        return err
    }

    jobs := make(chan string, fp.bufferSize)
    errs := make(chan error, len(files))

    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < fp.numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for file := range jobs {
                select {
                case <-ctx.Done():
                    return
                default:
                }

                if err := fp.processFile(ctx, file); err != nil {
                    errs <- fmt.Errorf("worker %d: %w", workerID, err)
                }
            }
        }(i)
    }

    // Send jobs
    go func() {
        defer close(jobs)
        for _, file := range files {
            select {
            case jobs <- file:
            case <-ctx.Done():
                return
            }
        }
    }()

    // Wait for completion
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        close(errs)
        for err := range errs {
            if err != nil {
                return err
            }
        }
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (fp *FileProcessor) processFile(ctx context.Context, file string) error {
    data, err := os.ReadFile(file)
    if err != nil {
        return err
    }

    // Process file content
    result := processContent(data)

    // Write result
    output := strings.Replace(file, ".txt", ".processed", 1)
    return os.WriteFile(output, result, 0644)
}
```

### Web Scraper with Worker Pool

```go
type Scraper struct {
    client      *http.Client
    maxWorkers  int
    maxRetries  int
    rateLimiter *time.Ticker
}

func NewScraper(maxWorkers, maxRetries, requestsPerSecond int) *Scraper {
    return &Scraper{
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
        maxWorkers: maxWorkers,
        maxRetries: maxRetries,
        rateLimiter: time.NewTicker(time.Second / time.Duration(requestsPerSecond)),
    }
}

type ScrapeResult struct {
    URL     string
    Title   string
    Links   []string
    Error   error
}

func (s *Scraper) Scrape(ctx context.Context, urls []string) []ScrapeResult {
    jobs := make(chan string, len(urls))
    results := make(chan ScrapeResult, len(urls))

    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < s.maxWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for url := range jobs {
                select {
                case <-ctx.Done():
                    return
                case <-s.rateLimiter.C:
                    result := s.scrapeWithRetry(ctx, url)
                    results <- result
                }
            }
        }()
    }

    // Send jobs
    go func() {
        defer close(jobs)
        for _, url := range urls {
            select {
            case jobs <- url:
            case <-ctx.Done():
                return
            }
        }
    }()

    // Wait for completion
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

func (s *Scraper) scrapeWithRetry(ctx context.Context, url string) ScrapeResult {
    var lastErr error

    for attempt := 0; attempt < s.maxRetries; attempt++ {
        select {
        case <-ctx.Done():
            return ScrapeResult{URL: url, Error: ctx.Err()}
        default:
        }

        result, err := s.scrape(ctx, url)
        if err == nil {
            return result
        }
        lastErr = err

        // Exponential backoff
        backoff := time.Duration(attempt*attempt) * time.Second
        select {
        case <-time.After(backoff):
        case <-ctx.Done():
            return ScrapeResult{URL: url, Error: ctx.Err()}
        }
    }

    return ScrapeResult{URL: url, Error: fmt.Errorf("max retries exceeded: %w", lastErr)}
}

func (s *Scraper) scrape(ctx context.Context, url string) (ScrapeResult, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return ScrapeResult{}, err
    }

    resp, err := s.client.Do(req)
    if err != nil {
        return ScrapeResult{}, err
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return ScrapeResult{}, err
    }

    title := doc.Find("title").Text()
    var links []string
    doc.Find("a[href]").Each(func(_ int, sel *goquery.Selection) {
        href, exists := sel.Attr("href")
        if exists {
            links = append(links, href)
        }
    })

    return ScrapeResult{
        URL:   url,
        Title: title,
        Links: links,
    }, nil
}
```

---

## Tóm tắt

```
┌─────────────────────────────────────────────────────────────────┐
│                    Go Concurrency Cheat Sheet                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Goroutine:    go func()                                       │
│  WaitGroup:    wg.Add(1) / wg.Done() / wg.Wait()              │
│  Channel:      ch := make(chan T) / make(chan T, buffer)       │
│  Send:         ch <- value                                     │
│  Receive:      value := <-ch                                   │
│  Close:        close(ch)                                       │
│  Select:       select { case ...: }                            │
│  Context:      ctx, cancel := context.WithCancel(parent)       │
│  Mutex:        mu.Lock() / mu.Unlock()                         │
│  RWMutex:      mu.RLock() / mu.RUnlock() / mu.Lock()          │
│  Atomic:       atomic.AddInt64(&counter, 1)                    │
│  Once:         once.Do(func() { ... })                         │
│  errgroup:     g.Go(func() error { ... })                      │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│  Node.js → Go Mapping:                                         │
│  ─────────────────────                                          │
│  Promise.all()      → errgroup.Group + WaitGroup               │
│  Promise.race()     → select {}                                │
│  async/await        → goroutine + channel                      │
│  EventEmitter       → channel                                  │
│  Worker Threads     → goroutine                                │
│  Cluster            → GOMAXPROCS (automatic)                   │
│  AbortController    → context.Context                          │
│  setTimeout         → time.After / context.WithTimeout         │
│  mutex (async)      → sync.Mutex                               │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│  Rules of thumb:                                                │
│  ────────────────                                               │
│  1. "Don't communicate by sharing memory;                       │
│      share memory by communicating."                            │
│  2. Prefer channels over mutexes                                │
│  3. Always use context for cancellation/timeout                 │
│  4. Always defer cancel() after WithCancel/WithTimeout          │
│  5. Use errGroup for concurrent error handling                  │
│  6. Use -race flag during development                           │
│  7. Avoid goroutine leaks (use context or done channels)        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

> **Tiếp theo:** Đọc `05-stdlib-packages.md` để tìm hiểu về standard library packages quan trọng.
