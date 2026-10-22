# Go Mental Model Cho Senior Node.js Developer

> Dành cho developer có 4 năm kinh nghiệm Node.js muốn hiểu Go một cách sâu sắc, không phải tutorial surface-level.

---

## 1. Deep Mental Model Mapping

Mỗi concept lớn trong Go map sang Node.js tương đương, nhưng mapping nào cũng có "mismatch" — phần mismatch mới là thứ quan trọng cần hiểu.

### 1.1 `package main` + `func main()` vs `index.js` Entrypoint

```go
// main.go
package main

import "fmt"

func main() {
    fmt.Println("Hello from Go")
}
```

```js
// index.js
console.log("Hello from Node");
```

**Giống nhau:** Cả hai đều là entrypoint của chương trình.

**Khác biệt sâu:**
- Trong Node, `index.js` chỉ là file JS bình thường. `node index.js` chạy nó, nhưng bạn cũng có thể `require('./index')` từ file khác.
- Trong Go, `package main` là **duy nhất** đánh dấu executable entrypoint. Compiler tìm `func main()` trong `package main` để build binary. Bạn **không thể** import `package main` từ package khác.
- Go có thể có nhiều file trong `package main`, tất cả share cùng namespace. Node mỗi file là module riêng với scope riêng.

```go
// File: main.go và helper.go cùng package main
// main.go
package main

func main() {
    sayHello() // gọi trực tiếp, không cần import
}

// helper.go
package main

import "fmt"

func sayHello() {
    fmt.Println("Hello")
}
```

**Mental model:** `package main` = "đây là entrypoint executable, không phải reusable library". Nếu bạn đang viết library, tạo `package mylib` riêng.

### 1.2 `go.mod` vs `package.json` (Modules vs npm)

```go
// go.mod
module github.com/yourname/yourproject

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    go.uber.org/zap v1.26.0
)
```

```json
// package.json
{
  "name": "your-project",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.18.0",
    "winston": "^3.11.0"
  }
}
```

**Khác biệt cốt lõi:**
- `go.mod` **không có version của project bạn**. Module path là identifier, không phải versioned package. Go modules chỉ version dependencies, không version chính nó theo cách npm làm.
- `go.sum` = checksum file, tương tự `package-lock.json` nhưng chỉ chứa cryptographic hashes cho integrity verification.
- Go **không có registry central** như npm. Dependencies được fetch trực tiếp từ source (GitHub, GitLab, etc.) và cache local.
- **Không có `node_modules`**. Dependencies được download vào `$GOPATH/pkg/mod`. Binary compile xong chứa tất cả code cần thiết.
- `go mod tidy` tự động thêm dependencies thiếu và xóa dependencies không dùng. Không có equivalent npm script nào tiện bằng.
- **Version runtime:** Node cần `nvm` + `.nvmrc`; từ Go 1.21 bạn chỉ cần ghi `go 1.22` trong `go.mod` là mọi lệnh `go` **tự tải và switch đúng toolchain** (`GOTOOLCHAIN=auto`). Chi tiết 4 cách (goenv, g, asdf/mise): [[06-tooling-testing]] mục 1.9.

```bash
# Node equivalent workflow
npm init → npm install express → npm install -D typescript

# Go equivalent workflow
go mod init github.com/you/project → go get github.com/gin-gonic/gin → go mod tidy
```

**Mental model:** Go modules nhẹ hơn npm rất nhiều. Không có `node_modules` hell, không có phantom dependencies, không có hoisting issues.

### 1.3 `import` Compile-Time vs `require/import` Runtime

```go
// Go: compile-time resolution
import (
    "fmt"
    "net/http"
    "github.com/gin-gonic/gin"
)
```

```js
// Node: runtime resolution
const express = require('express');       // CommonJS
import express from 'express';            // ESM
```

**Đây là khác biệt ARCHITECTURAL quan trọng:**
- Go import xảy ra **khi compile**. Compiler resolve tất cả dependencies, check types, link code. Nếu thiếu import → compile error. Nếu import mà không dùng → **compile error**.
- Node require/import xảy ra **khi runtime**. Module được resolve và execute khi dòng code đó chạy lần đầu. Circular dependencies tạo partial objects.

```go
// Go: Unused import = compile error. Phải explicit.
import (
    "fmt"
    "os" // ERROR: imported and not used
)

func main() {
    fmt.Println("hello")
}
```

```js
// Node: Unused import chỉ là warning ở best
const fs = require('fs');  // Không dùng? Không sao. Silent.
```

**Tại sao Go làm vậy:**
- Compile-time resolution cho phép dead code elimination (binary nhỏ hơn)
- Không có dynamic import tricks → binary có thể optimize hoàn toàn
- Force developer clean code: không import bừa bãi

**Mental model:** Go import = "khai báo dependency cho compiler". Node import = "lazy load module ở runtime".

### 1.4 `struct` vs Class/Object Literal

```go
// Go: struct
type User struct {
    ID    int
    Name  string
    Email string
    Age   int
}

// Constructor convention (không phải keyword)
func NewUser(name, email string, age int) *User {
    return &User{
        Name:  name,
        Email: email,
        Age:   age,
    }
}

// Method
func (u *User) IsAdult() bool {
    return u.Age >= 18
}
```

```js
// Node: class hoặc object literal
class User {
    constructor(name, email, age) {
        this.id = null;
        this.name = name;
        this.email = email;
        this.age = age;
    }

    isAdult() {
        return this.age >= 18;
    }
}

// Hoặc object literal
const user = {
    id: null,
    name: "John",
    email: "john@example.com",
    age: 25,
    isAdult() { return this.age >= 18; }
};
```

**Khác biệt sâu:**
- Go `struct` **chỉ chứa data**. Methods được define riêng với receiver.
- **Không có inheritance**. Go dùng composition (embedding).
- `struct` có zero value tự động: `var u User` → tất cả fields đều zero value, không phải `undefined` hay `null`.
- `struct` là **value type**. Copy `u1 = u2` tạo bản sao hoàn toàn, không share reference.

```go
// Composition thay vì inheritance
type BaseUser struct {
    ID    int
    Name  string
}

type Admin struct {
    BaseUser              // embedded, không phải inheritance
    Permissions []string
}

admin := Admin{
    BaseUser:    BaseUser{ID: 1, Name: "Root"},
    Permissions: []string{"read", "write", "delete"},
}
fmt.Println(admin.Name) // truy cập trực tiếp từ BaseUser
```

```js
// JS: inheritance
class BaseUser {
    constructor(id, name) {
        this.id = id;
        this.name = name;
    }
}

class Admin extends BaseUser {
    constructor(id, name, permissions) {
        super(id, name);
        this.permissions = permissions;
    }
}
```

**Mental model:** `struct` = "plain data container, không có behavior attached". Behavior nằm ở methods. Composition > Inheritance, luôn luôn.

### 1.5 Interface vs TypeScript Interface (Implicit vs Explicit)

```go
// Go: implicit interface satisfaction
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// Bất kỳ type nào có method Read([]byte) (int, error) đều satisfies Reader
// Không cần declare "implements Reader"
type FileReader struct{ file *os.File }

func (fr FileReader) Read(p []byte) (int, error) {
    return fr.file.Read(p)
}
// FileReader tự động satisfies Reader. Không cần keyword nào.
```

```ts
// TypeScript: explicit implementation
interface Reader {
    read(buffer: Uint8Array): { bytesRead: number; buffer: Uint8Array };
}

class FileReader implements Reader {
    read(buffer: Uint8Array): { bytesRead: number; buffer: Uint8Array } {
        // ...
    }
}
```

**Đây là khác biệt DESIGN PHILOSOPHY lớn:**
- Go interface = **implicit contract**. Consumer define interface, không phải producer. Bạn không cần biết type `X` implement interface `Y` — chỉ cần nó có đúng method signatures.
- TypeScript interface = **explicit declaration**. Producer phải biết và declare implements.
- Go interface thường **rất nhỏ** (1-3 methods). `io.Reader` chỉ có 1 method. `fmt.Stringer` chỉ có `String() string`.

```go
// Interface nhỏ, composition lớn
type Stringer interface {
    String() string
}

type Reader interface {
    Read(p []byte) (n int, err error)
}

type ReadWriter interface {
    Reader
    Writer
}

// Interface acceptance cực kỳ linh hoạt
func process(r io.Reader) error {
    // Nhận được: File, Network connection, Buffer, Pipe, ...
    // Bất kỳ thứ gì có Read()
    buf := make([]byte, 1024)
    n, err := r.Read(buf)
    // ...
}
```

**Mental model:** Go interface = "capability declaration ở consumer side". Bạn define interface khi bạn cần abstraction, không phải khi bạn tạo type. Đây gọi là **"Accept interfaces, return structs"**.

### 1.6 Goroutine vs Promise/async-await (FUNDAMENTALLY DIFFERENT!)

```go
// Go: goroutine = lightweight OS thread multiplexed onto real threads
go func() {
    fmt.Println("Running concurrently")
}()
```

```js
// Node: Promise = future value in event loop
new Promise((resolve) => {
    console.log("Running in event loop tick");
    resolve();
});

// async/await = syntactic sugar cho Promises
async function fetchData() {
    const res = await fetch('https://api.example.com/data');
    return res.json();
}
```

**Đây là khác biệt LỚN NHẤT. Đừng bỏ qua phần này.**

| Aspect | Goroutine | Promise/async-await |
|--------|-----------|-------------------|
| Mô hình | Real concurrent execution (multi-core) | Cooperative concurrency (single thread) |
| Parallelism | **Có** — nhiều goroutines chạy đồng thời trên nhiều CPU cores | **Không** — chỉ có 1 thread, event loop multiplexing |
| Blocking | Goroutine có thể block mà không ảnh hưởng others | Blocking 1 operation = block toàn bộ event loop |
| Memory | ~2KB stack ban đầu, grow tự động | ~microtask allocation, nhưng overhead từ closure chains |
| Lifecycle | Explicit lifecycle management qua channels/context | Automatic cleanup khi Promise settle |
| Creation cost | Rất rẻ (nanoseconds) | Rất rẻ |
| Scaling | Scale tới hàng triệu goroutines | Scale tới hàng trăm nghìn pending promises |

```go
// Goroutine chạy TRULY PARALLEL
func main() {
    // Cả 4 goroutines này có thể chạy đồng thời trên 4 CPU cores
    for i := 0; i < 4; i++ {
        go func(id int) {
            // CPU-intensive work
            sum := 0
            for j := 0; j < 1_000_000_000; j++ {
                sum += j
            }
            fmt.Printf("Worker %d: %d\n", id, sum)
        }(i)
    }
    time.Sleep(10 * time.Second)
}
```

```js
// Node: dù tạo 4 promises, vẫn chỉ 1 thread chạy CPU work
async function main() {
    const promises = [];
    for (let i = 0; i < 4; i++) {
        promises.push((async (id) => {
            // CPU-intensive work BLOCK event loop
            let sum = 0;
            for (let j = 0; j < 1_000_000_000; j++) {
                sum += j;
            }
            console.log(`Worker ${id}: ${sum}`);
        })(i));
    }
    await Promise.all(promises); // Chạy tuần tự thực tế!
}
```

**Mental model:** Goroutine = "cheap thread bạn tạo thoải mái". Promise = "callback trong event loop". Goroutine cho bạn **true parallelism**. Promise chỉ cho bạn **concurrency** (chuyển đổi context, không chạy đồng thời).

### 1.7 Channel vs EventEmitter/MessagePort

```go
// Go: channel = typed, synchronous/asynchronous message pipe
ch := make(chan string)    // unbuffered, synchronous
ch := make(chan string, 5) // buffered, async up to 5 messages

// Send
ch <- "hello"

// Receive
msg := <-ch

// Close
close(ch)
```

```js
// Node: EventEmitter
const EventEmitter = require('events');
const emitter = new EventEmitter();

emitter.on('message', (msg) => console.log(msg));
emitter.emit('message', 'hello');

// MessagePort (Worker threads)
const { MessageChannel } = require('worker_threads');
const { port1, port2 } = new MessageChannel();
port1.on('message', (msg) => console.log(msg));
port2.postMessage('hello');
```

**Khác biệt cốt lõi:**
- Channel là **typed** và **directional** (`chan<-` send-only, `<-chan` receive-only).
- Channel có thể **block**: unbuffered channel block sender cho đến khi có receiver. Đây là synchronization mechanism, không chỉ message passing.
- Channel là **first-class value** trong Go. Truyền channel vào function, return từ function, lưu trong struct.
- **Không có event name** — channel itself là identifier. Không có string-based event dispatch.

```go
// Pattern: Fan-out, Fan-in
func fanOut(input <-chan int, workers int) []<-chan int {
    channels := make([]<-chan int, workers)
    for i := 0; i < workers; i++ {
        channels[i] = worker(input)
    }
    return channels
}

func fanIn(channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    merged := make(chan int)
    for _, ch := range channels {
        wg.Add(1)
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

**Mental model:** Channel = "typed message pipe với built-in synchronization". EventEmitter = "string-keyed observer pattern". Channel mạnh hơn vì nó enforce types, directionality, và synchronization.

### 1.8 `context.Context` vs `AbortController`

```go
// Go: context cho cancellation, timeout, và value propagation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := http.NewRequestWithContext(ctx, "GET", url, nil)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        // Timeout
    }
}
```

```js
// Node: AbortController
const controller = new AbortController();
const timeoutId = setTimeout(() => controller.abort(), 5000);

try {
    const res = await fetch(url, { signal: controller.signal });
} catch (err) {
    if (err.name === 'AbortError') {
        // Aborted
    }
} finally {
    clearTimeout(timeoutId);
}
```

**Go context mạnh hơn AbortController rất nhiều:**
- Context có thể **carry values** qua call chain (request ID, auth token, etc.)
- Context có thể **tree hierarchy** — parent cancel → tất cả children cancel.
- Context là **idiomatic first parameter** trong Go: `func DoWork(ctx context.Context, ...) error`
- **Không có try-catch** để cleanup — context + defer là pattern thay thế.

```go
// Context propagation trong real-world API handler
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // context từ HTTP request

    // Thêm value vào context
    ctx = context.WithValue(ctx, "requestID", uuid.New().String())

    // Truyền context xuống các layer
    user, err := userService.FindByID(ctx, userID)
    if err != nil {
        if ctx.Err() != nil {
            // Client disconnected hoặc timeout
            return
        }
        http.Error(w, err.Error(), 500)
        return
    }
}
```

**Mental model:** Context = "cancellation + timeout + value propagation qua call chain". AbortController chỉ là cancellation. Context là backbone của Go request lifecycle.

### 1.9 Error as Value vs throw/try-catch

```go
// Go: error là giá trị bình thường, return từ function
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doSomething failed: %w", err)
}
```

```js
// Node: exception flow
try {
    const result = await doSomething();
} catch (err) {
    throw new Error(`doSomething failed: ${err.message}`);
}
```

**Đây là khác biệt có ảnh hưởng lớn nhất đến code style:**

```go
// Go: mỗi error point phải explicit check
func ProcessOrder(ctx context.Context, orderID string) error {
    order, err := repo.FindOrder(ctx, orderID)  // check 1
    if err != nil {
        return fmt.Errorf("find order: %w", err)
    }

    user, err := repo.FindUser(ctx, order.UserID) // check 2
    if err != nil {
        return fmt.Errorf("find user: %w", err)
    }

    err = payment.Charge(ctx, user, order.Amount) // check 3
    if err != nil {
        return fmt.Errorf("charge payment: %w", err)
    }

    err = notification.Send(ctx, user.Email, "Order confirmed") // check 4
    if err != nil {
        // Log nhưng không return — notification fail không phải critical
        log.Printf("notification failed: %v", err)
    }

    return nil // success
}
```

```js
// Node: exception tự propagate lên
async function processOrder(orderId) {
    const order = await repo.findOrder(orderId);      // throw nếu fail
    const user = await repo.findUser(order.userId);    // throw nếu fail
    await payment.charge(user, order.amount);          // throw nếu fail
    try {
        await notification.send(user.email, 'Order confirmed');
    } catch (err) {
        console.error('notification failed:', err);    // swallow non-critical
    }
}
```

**Go error handling xem như verbose, nhưng thực tế cho bạn control rõ ràng hơn:**
- Bạn **thấy chính xác** error xảy ra ở đâu (không cần stack trace cho simple errors).
- Bạn **quyết định** error nào propagate, error nào wrap, error nào swallow.
- Error wrapping (`%w`) cho phép unwrap chain: `errors.Is(err, context.DeadlineExceeded)` traverse qua toàn bộ chain.
- **Không có unchecked exception** — không bao giờ surprise error từ tầng nào đó.

### 1.10 `defer` vs `try-finally`

```go
func readFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close() // Chạy khi function return, dù error hay không

    buf := make([]byte, 1024)
    _, err = f.Read(buf)
    if err != nil {
        return err
    }

    fmt.Println(string(buf))
    return nil
}
```

```js
// Node: try-finally hoặc using
async function readFile(path) {
    let fh;
    try {
        fh = await fs.promises.open(path, 'r');
        const buf = Buffer.alloc(1024);
        await fh.read(buf, 0, 1024, 0);
        console.log(buf.toString());
    } finally {
        if (fh) await fh.close();
    }
}

// Node 18+: using (Explicit Resource Management)
{
    await using fh = await fs.promises.open(path, 'r');
    // auto close khi scope exit
}
```

**Đặc điểm defer:**
- LIFO order (stack): defer đầu tiên chạy cuối cùng.
- **Argument evaluation ngay khi defer statement execute**, không phải khi defer chạy.

```go
func example() {
    x := 10
    defer fmt.Println("deferred:", x) // x được eval = 10 ngay lúc này
    x = 20
    fmt.Println("normal:", x)
}
// Output:
// normal: 20
// deferred: 10  (không phải 20!)
```

**Mental model:** `defer` = "register cleanup ngay sau khi resource được acquire". Nhẹ hơn try-finally vì bạn không cần nesting. Pattern `defer f.Close()` ngay sau `os.Open()` là idiomatic Go.

### 1.11 Zero Value vs `undefined`/`null`

```go
// Go: mọi type có zero value rõ ràng
var s string   // "" (empty string, không phải nil)
var n int      // 0
var f float64  // 0.0
var b bool     // false
var p *int     // nil
var sl []int   // nil (nhưng len(sl) = 0, append vẫn hoạt động)
var m map[string]int // nil (nhưng đọc key không có → zero value, không panic)
var ch chan int // nil (nhưng send/recv vào nil channel → block forever)
var fn func()  // nil
var iface error // nil
var t time.Time // time.Time{} (zero time, January 1, year 1, 00:00:00 UTC)
```

```js
// Node: undefined và null là 2 khái niệm khác nhau
let s;           // undefined
let n = null;    // null
const obj = {};
obj.foo;         // undefined
JSON.parse('{"a":null}').a; // null
```

**Tại sao zero value quan trọng:**
- **Không bao giờ nil pointer panic** khi dùng value types (string, int, etc.)
- **Không cần optional chaining** cho value types: `var s string` → `s` luôn usable, chỉ là empty string.
- **Zero value có ý nghĩa**: `var count int` → count = 0, hợp lệ để dùng.
- **Chỉ pointer, slice, map, channel, interface, func** mới có thể nil.

```go
// Zero value cho phép safe usage không cần initialization
type Config struct {
    Port    int    // zero = 0, có thể dùng default
    Host    string // zero = "", check emptiness
    Debug   bool   // zero = false
    Timeout time.Duration // zero = 0, có thể dùng default
}

func NewConfig() Config {
    var c Config    // tất cả zero values, hoàn toàn usable
    c.Port = 8080  // override nếu cần
    return c
}

// Map zero value: nil nhưng safe
var m map[string]int
fmt.Println(m["key"]) // 0, không panic
fmt.Println(len(m))   // 0, không panic
// m["key"] = 1        // PANIC! nil map không thể write
m = make(map[string]int) // phải init trước khi write
```

**Mental model:** Go không có "null/undefined confusion". Mọi thứ có zero value predictable. Chỉ pointer/reference types mới nil. Đây là source of bugs lớn trong JS nhưng gần như không tồn tại trong Go.

---

## 2. Go is Explicit, Node is Implicit

### 2.1 Error Handling: Forced Check vs Exception Bubbling

```go
// Go: BẮT BUỘC check mọi error
func GetUser(ctx context.Context, id string) (*User, error) {
    row := db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = $1", id)

    var u User
    err := row.Scan(&u.ID, &u.Name, &u.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrUserNotFound // custom sentinel error
        }
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }

    return &u, nil
}

// Caller PHẢI check error
user, err := GetUser(ctx, "123")
if err != nil {
    if errors.Is(err, ErrUserNotFound) {
        // handle not found
    }
    // handle other errors
}
```

```js
// Node: exception tự propagate, có thể quên catch
async function getUser(id) {
    const row = await db.query('SELECT * FROM users WHERE id = $1', [id]);
    if (row.length === 0) {
        throw new UserNotFoundError(`User ${id} not found`);
    }
    return row[0];
}

// Caller có thể quên try-catch → unhandled rejection
const user = await getUser("123"); // nếu throw, rejection propagate
// Hoặc wrap trong try-catch nếu muốn handle
```

**Go cho bạn EXPLICIT control flow:**
- Mọi error path đều visible trong code.
- Không có "exception invisibly propagate qua 10 layers".
- Error wrapping cho context: `fmt.Errorf("query user %s: %w", id, err)` → bạn biết error xảy ra ở step nào.
- Sentinel errors + `errors.Is()` cho programmatic error matching.

### 2.2 Không Có `null`/`undefined` Distinction

```js
// Node: null vs undefined confusion
function greet(name) {
    if (name === undefined) {
        return "Hello stranger";
    }
    if (name === null) {
        return "Hello nobody"; // ???
    }
    return `Hello ${name}`;
}

greet();          // "Hello stranger" — name is undefined
greet(null);      // "Hello nobody" — name is null
greet(undefined); // "Hello stranger" — same as first
```

```go
// Go: chỉ có nil cho pointer/reference types
func Greet(name *string) string {
    if name == nil {
        return "Hello stranger"
    }
    return fmt.Sprintf("Hello %s", *name)
}

// Hoặc dùng zero value pattern
func GreetV2(name string) string {
    if name == "" {
        return "Hello stranger"
    }
    return fmt.Sprintf("Hello %s", name)
}
```

**Trong Go, bạn không bao giờ cần phân biệt "no value" vs "empty value" cho value types.** Zero value đã cover "empty" case.

### 2.3 Không Cần Optional Chaining (Vì Zero Values)

```js
// Node: optional chaining để tránh TypeError
const city = user?.address?.city?.toUpperCase();
const firstTag = user?.tags?.[0];
const result = user?.getName?.();
```

```go
// Go: zero value patterns
// Nếu User là struct (value type):
city := user.Address.City // zero value nếu bất kỳ field nào empty. Không panic.

// Nếu cần check nil (pointer types):
if user != nil && user.Address != nil {
    city := user.Address.City
}

// Idiomatic Go: defensive copy
func GetCity(user *User) string {
    if user == nil {
        return ""
    }
    if user.Address == nil {
        return ""
    }
    return user.Address.City
}
```

**Go code nhìn verbose hơn, nhưng bạn thấy MỌI possible path.** Không có hidden `TypeError: Cannot read property 'city' of undefined`.

### 2.4 Type System: Go Compile-Time vs TypeScript Runtime Erasure

```ts
// TypeScript: types bị erase ở runtime
interface User {
    id: number;
    name: string;
}

function process(user: User) {
    console.log(user.name);
}

// TypeScript compile → JS, types biến mất
process({ id: 1, name: "John" }); // OK
process({ id: "1", name: "John" }); // TS error nhưng runtime vẫn chạy
process({ id: 1 }); // TS error nhưng runtime: user.name = undefined
```

```go
// Go: types exist at runtime
type User struct {
    ID   int
    Name string
}

func process(user User) {
    fmt.Println(user.Name)
}

process(User{ID: 1, Name: "John"}) // OK
process(User{ID: 1})               // Compile error: missing Name field
process(User{ID: "1", Name: "John"}) // Compile error: cannot use string as int
```

**Go type system là REAL, không phải annotation:**
- Compile-time type checking prevents type errors at runtime.
- Không có `any` equivalent dễ dãi (Go có `interface{}` nhưng dùng rất ít).
- Không có runtime type assertion surprises (Go type assertion có `, ok` pattern).
- Generics trong Go 1.18+ thêm flexibility nhưng vẫn compile-time checked.

---

## 3. Memory Model Differences

### 3.1 Stack vs Heap Allocation

```go
// Go: compiler quyết định stack hay heap (escape analysis)
func createUser() *User {
    u := User{Name: "John"} // có thể stack...
    return &u               // ...nhưng return pointer → escape to heap
}

func processUser() {
    u := User{Name: "John"} // stack allocation (không escape)
    fmt.Println(u.Name)
    // u bị GC khi function return (hoặc ngay khi không dùng nếu stack)
}
```

```js
// Node/V8: mọi object allocation trên heap
function createUser() {
    const u = { name: "John" }; // always heap
    return u;
}

// V8 optimize: hidden classes, inline caching
// Nhưng fundamentally vẫn là heap allocation + GC
```

**Go compiler thực hiện escape analysis:**
- Nếu variable không "escape" function scope → stack allocation (zero cost).
- Nếu variable escape (return pointer, stored in global, passed to goroutine) → heap allocation.
- Stack allocation = trivially cheap (move stack pointer). Heap allocation = cần GC.

```bash
# Xem escape analysis
go build -gcflags="-m" main.go
# Output: "moved to heap: u" nếu u escape
```

### 3.2 Value Semantics vs Reference Semantics

```go
// Go: struct là value type — copy khi assign hoặc pass to function
type Point struct {
    X, Y float64
}

func movePoint(p Point) Point {
    p.X += 10 // modify copy, không ảnh hưởng original
    return p
}

func main() {
    p1 := Point{X: 1, Y: 2}
    p2 := movePoint(p1)
    fmt.Println(p1) // {1 2} — unchanged
    fmt.Println(p2) // {11 2}
}
```

```js
// Node: object là reference type
function movePoint(p) {
    p.x += 10; // mutate original!
    return p;
}

const p1 = { x: 1, y: 2 };
const p2 = movePoint(p1);
console.log(p1); // { x: 11, y: 2 } — MUTATED!
console.log(p2); // { x: 11, y: 2 } — same reference

// Phải explicit clone nếu muốn copy
const p3 = { ...p1 };
```

**Value semantics = predictable behavior. Không có "accidental mutation" bugs.**

Khi cần reference semantics, dùng pointer:

```go
func movePoint(p *Point) {
    p.X += 10 // modify original qua pointer
}

p1 := Point{X: 1, Y: 2}
movePoint(&p1)
fmt.Println(p1) // {11 2} — modified
```

### 3.3 Tại Sao Go Không Cần GC Tuning Như Node/V8

- Go GC là **concurrent, low-latency** — typical pause time < 1ms.
- Stack allocation giảm pressure trên GC significantly.
- Không có "nursery" hay "generational" GC tuning.
- V8 GC phức tạp hơn: nursery, old generation, mark-sweep, compaction. Node developer phải hiểu heap limits, `--max-old-space-size`.
- Go: hầu như **không bao giờ** cần tune GC. `GOGC` environment variable là rare optimization.

### 3.4 Pointer Semantics Cho JS Dev

```go
// Pointer trong Go đơn giản hơn C/C++ rất nhiều
func increment(p *int) {
    *p++ // dereference và modify
}

func main() {
    x := 5
    increment(&x)  // &x = address of x
    fmt.Println(x)  // 6
}
```

**Mental model cho JS dev:**
- `&x` = "get reference to x" (giống tạo object wrapper `{value: x}` trong JS).
- `*p` = "dereference p, get the value" (giống `wrapper.value` trong JS).
- Pointer = cho phép function modify caller's variable.
- Không có pointer arithmetic như C.
- Không có null pointer exception nếu bạn check nil trước khi use.

```go
// Idiomatic Go: receiver type quyết định mutation
type User struct {
    Name string
    Age  int
}

// Value receiver: không mutate
func (u User) Greet() string {
    return fmt.Sprintf("Hi, I'm %s", u.Name)
}

// Pointer receiver: có thể mutate
func (u *User) SetAge(age int) {
    u.Age = age // modify original
}
```

**Rule of thumb:** Method nào cần modify state → pointer receiver. Method nào chỉ read → value receiver. Khi unsure → pointer receiver (an toàn hơn).

---

## 4. Compilation vs Interpretation

### 4.1 Go Compile to Single Binary vs Node Runtime

```bash
# Go: compile to standalone binary
go build -o myapp main.go
./myapp  # Chạy trực tiếp, không cần Go installed

# Node: cần Node.js runtime
node myapp.js  # Cần Node.js installed trên target machine
```

**Deployment implications:**

```dockerfile
# Go: multi-stage build → tiny image
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM scratch
COPY --from=builder /app/server /server
ENTRYPOINT ["/server"]
# Final image: ~10-15MB (chỉ binary)
```

```dockerfile
# Node: cần Node runtime → larger image
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
CMD ["node", "server.js"]
# Final image: ~100-200MB (Node runtime + node_modules)
```

### 4.2 Cross-Compilation

```bash
# Go: cross-compile cực dễ
GOOS=linux GOARCH=amd64 go build -o server-linux .
GOOS=darwin GOARCH=arm64 go build -o server-mac .
GOOS=windows GOARCH=amd64 go build -o server.exe .
# Không cần Docker, không cần target machine

# Node: "cross-platform" nhưng thực tế chỉ là interpreter
# Phải install Node trên target, hoặc bundle với pkg/nexe
```

### 4.3 Build Time vs Startup Time Tradeoffs

| Aspect | Go | Node |
|--------|-----|------|
| Compilation time | Seconds to minutes | N/A (interpreted) |
| Startup time | **Near-instant** (microseconds) | 50-500ms (module loading, JIT warmup) |
| Runtime warmup | None (compiled to native) | JIT needs time to optimize hot paths |
| Cold start (serverless) | **~1-10ms** | **~100-5000ms** |

**Tại sao startup time quan trọng:**
- **Serverless/Lambda:** Go cold start gần như instant. Node cold start có thể 1-5s.
- **CLI tools:** Go binary chạy ngay. Node cần interpreter boot.
- **Container orchestration:** Go containers start nhanh hơn, healthcheck pass nhanh hơn.

### 4.4 Deployment Advantages

```bash
# Go deployment: copy binary là xong
scp server user@prod:/usr/local/bin/
ssh user@prod 'systemctl restart myapp'

# Node deployment: cần runtime + dependencies + version management
scp -r . user@prod:/app/
ssh user@prod 'cd /app && npm ci && pm2 restart myapp'
```

---

## 5. Concurrency Philosophy

### 5.1 Node.js: Single-Threaded Event Loop

```js
// Node: single thread, non-blocking I/O
const http = require('http');

const server = http.createServer(async (req, res) => {
    // I/O operations: non-blocking, handled by libuv
    const data = await fs.promises.readFile('file.txt');
    const dbResult = await db.query('SELECT * FROM users');

    // CPU-intensive: BLOCKS event loop!
    const hash = crypto.createHash('sha256').update(hugeData).digest();
    // Mọi requests khác phải đợi

    res.end('OK');
});
```

```
Event Loop Visualization (Node.js):
┌───────────────────────────┐
│         Event Loop        │
│                           │
│  ┌─────────┐ ┌─────────┐ │
│  │  Timers  │ │  Poll   │ │ ──→ I/O callbacks
│  └─────────┘ └─────────┘ │
│  ┌─────────┐ ┌─────────┐ │
│  │  Check  │ │ Close   │ │ ──→ Microtasks
│  └─────────┘ └─────────┘ │
│                           │
│  Single thread: chỉ 1     │
│  execution context active │
└───────────────────────────┘
```

### 5.2 Go: Multi-Threaded by Default

```go
// Go: multiple OS threads + goroutine scheduler
func main() {
    // GOMAXPROCS defaults to number of CPU cores
    // runtime.GOMAXPROCS(4) // nếu cần set explicit

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // I/O operations: non-blocking
        data, _ := os.ReadFile("file.txt")
        row := db.QueryRow("SELECT * FROM users")

        // CPU-intensive: KHÔNG BLOCK goroutines khác
        hash := sha256.Sum256(hugeData)

        w.Write(data)
    })

    http.ListenAndServe(":8080", nil)
}
```

```
Go Runtime Visualization:
┌─────────────────────────────────────┐
│           Go Runtime                │
│  ┌───────────┐  ┌───────────┐      │
│  │  Thread 1  │  │  Thread 2  │     │
│  │  ┌──────┐ │  │  ┌──────┐ │     │
│  │  │  G1  │ │  │  │  G3  │ │     │
│  │  │  G2  │ │  │  │  G4  │ │     │
│  │  └──────┘ │  │  └──────┘ │     │
│  └───────────┘  └───────────┘      │
│                                     │
│  M:N scheduling: M goroutines       │
│  multiplexed onto N OS threads      │
└─────────────────────────────────────┘
```

### 5.3 CSP (Communicating Sequential Processes) vs Event Loop

```go
// Go CSP: goroutines communicate qua channels
func main() {
    ch := make(chan string)

    // Producer goroutine
    go func() {
        for i := 0; i < 5; i++ {
            ch <- fmt.Sprintf("message %d", i)
            time.Sleep(100 * time.Millisecond)
        }
        close(ch)
    }()

    // Consumer goroutine (main)
    for msg := range ch {
        fmt.Println("Received:", msg)
    }
}
```

```js
// Node: event-driven hoặc async iteration
const { EventEmitter } = require('events');
const emitter = new EventEmitter();

// Producer
setTimeout(() => {
    for (let i = 0; i < 5; i++) {
        emitter.emit('message', `message ${i}`);
    }
    emitter.emit('done');
}, 0);

// Consumer
emitter.on('message', (msg) => console.log('Received:', msg));
emitter.on('done', () => console.log('Done'));
```

### 5.4 "Don't Communicate By Sharing Memory; Share Memory By Communicating"

```go
// SAI (mutex-based, giống Node/JS style):
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

// ĐÚNG (channel-based, Go style):
func counter(done <-chan struct{}) <-chan int {
    ch := make(chan int)
    go func() {
        count := 0
        for {
            select {
            case <-done:
                close(ch)
                return
            case ch <- count:
                count++
            }
        }
    }()
    return ch
}
```

**Mental model:**
- Node: "Tôi có data chia sẻ, tôi cần lock để bảo vệ nó" (mutex).
- Go: "Tôi gửi data qua channel, mỗi goroutine sở hữu data riêng" (CSP).
- Channels enforce synchronization: sender và receiver phải "gặp nhau" tại cùng thời điểm (unbuffered) hoặc buffer space available (buffered).

```go
// Real-world pattern: worker pool với channels
func workerPool(jobs <-chan Job, results chan<- Result, workers int) {
    var wg sync.WaitGroup
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for job := range jobs {
                result := process(job)
                results <- result
            }
        }(i)
    }
    wg.Wait()
    close(results)
}
```

---

## 6. Common Mistakes Node Devs Make in Go

### 6.1 Viết JS-Style Code Trong Go

```go
// SAI: JS style — callback hell trong Go
func getUser(id string, callback func(*User, error)) {
    // ... async style trong sync language
}

// ĐÚNG: Go style — return error
func GetUser(ctx context.Context, id string) (*User, error) {
    // direct return, không callback
}
```

```go
// SAI: treating Go like dynamic language
func process(data interface{}) interface{} {
    // Too much interface{}, mất type safety
    return data
}

// ĐÚNG: use concrete types hoặc generics
func process[T any](data T) T {
    return data
}
```

### 6.2 Over-Using Interfaces

```go
// SAI: interface cho mọi thứ (Java/TS style)
type UserServiceInterface interface {
    FindByID(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter Filter) ([]*User, error)
}

// ĐÚNG: chỉ define interface khi CẦN abstraction
// Accept interfaces, return structs
type UserFinder interface {
    FindByID(ctx context.Context, id string) (*User, error)
}

func ProcessUser(ctx context.Context, finder UserFinder, id string) error {
    user, err := finder.FindByID(ctx, id)
    // ...
}
```

**Rule:** Interface define ở consumer side. Không define interface "just in case".

### 6.3 Không Hiểu Value vs Pointer Semantics

```go
// SAI: không hiểu mutation
func addTag(user User, tag string) {
    user.Tags = append(user.Tags, tag)
    // user.Tags bị modify trong function nhưng caller không thấy!
    // Vì User là value type, append tạo slice mới nhưng backing array có thể shared
}

// ĐÚNG: dùng pointer receiver
func (u *User) AddTag(tag string) {
    u.Tags = append(u.Tags, tag)
}
```

```go
// SAI: pointer không cần thiết
func GetName(user *User) string {
    // chỉ read, không modify → value receiver đủ
    return user.Name
}

// ĐÚNG: value receiver cho read-only
func (u User) GetName() string {
    return u.Name
}
```

### 6.4 Treat Goroutines Like Promises

```go
// SAI: fire-and-forget như Promise không await
go doSomething()   // goroutine có thể bị GC nếu main() exit
go doSomethingElse() // không có lifecycle management

// ĐÚNG: track goroutine lifecycle
var wg sync.WaitGroup
wg.Add(2)
go func() {
    defer wg.Done()
    doSomething()
}()
go func() {
    defer wg.Done()
    doSomethingElse()
}()
wg.Wait() // đợi cả hai hoàn thành
```

```go
// SAI: goroutine leak (như Promise leak)
func leaky() {
    ch := make(chan int)
    go func() {
        ch <- 42 // goroutine block forever nếu không ai receive
    }()
    // ch bị GC, goroutine vẫn block → leak
}

// ĐÚNG: context cancellation cho goroutine lifecycle
func notLeaky(ctx context.Context) {
    ch := make(chan int)
    go func() {
        select {
        case ch <- 42:
        case <-ctx.Done():
            return // clean exit
        }
    }()
}
```

### 6.5 Ignoring Error Returns

```go
// SAI: bỏ qua error
json.Unmarshal(data, &user) // error bị bỏ qua
resp, _ := http.Get(url)    // error bị ignore
file.Write(data)             // write error bị ignore

// ĐÚNG: handle mọi error
if err := json.Unmarshal(data, &user); err != nil {
    return fmt.Errorf("unmarshal user: %w", err)
}
```

### 6.6 Building Too Many Abstractions Too Early

```go
// SAI: abstract quá sớm (JS/TS framework style)
type Repository[T any] interface {
    Find(ctx context.Context, id string) (*T, error)
    Create(ctx context.Context, entity *T) error
    Update(ctx context.Context, entity *T) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter Filter) ([]*T, error)
}

type Service[T any] struct {
    repo Repository[T]
}

type Controller[T any] struct {
    service Service[T]
}

// ĐÚNG: start simple, refactor khi CẦN
type UserStore struct {
    db *sql.DB
}

func (s *UserStore) FindByID(ctx context.Context, id string) (*User, error) {
    // direct SQL, simple, readable
}

// Refactor to interface KHI bạn cần swap implementation (test, different DB, etc.)
```

---

## 7. Go Làm Tốt Hơn Node.js

### 7.1 True Parallelism

```go
// Go: thực sự parallel trên nhiều CPU cores
func main() {
    runtime.GOMAXPROCS(8) // 8 cores

    var wg sync.WaitGroup
    for i := 0; i < 8; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            // CPU-intensive: chạy TRULY PARALLEL
            sum := 0
            for j := 0; j < 1_000_000_000; j++ {
                sum += j
            }
            fmt.Printf("Worker %d: %d\n", id, sum)
        }(i)
    }
    wg.Wait()
}
// 8 goroutines chạy đồng thời trên 8 cores
// Total time ≈ 1x single core time
```

```js
// Node: Worker Threads (limited parallelism)
const { Worker, isMainThread, workerData } = require('worker_threads');

if (isMainThread) {
    const workers = [];
    for (let i = 0; i < 8; i++) {
        workers.push(new Promise((resolve) => {
            const w = new Worker(__filename, { workerData: i });
            w.on('message', resolve);
        }));
    }
    Promise.all(workers).then(() => console.log('Done'));
} else {
    let sum = 0;
    for (let j = 0; j < 1_000_000_000; j++) {
        sum += j;
    }
    parentPort.postMessage(`Worker ${workerData}: ${sum}`);
}
// Worker Threads có overhead lớn hơn goroutines nhiều
// Và API phức tạp hơn
```

### 7.2 CPU-Intensive Tasks

| Task | Go | Node |
|------|-----|------|
| JSON parsing (large) | Native, fast | Fast (V8 optimized) |
| Image processing | Excellent (standard library + native bindings) | Need sharp/jimp (native bindings anyway) |
| Crypto | Standard library, no external deps | Node crypto (good but limited algorithms) |
| Compression | Standard gzip/zlib | zlib module (OK) |
| Data processing | Fast, parallelizable | Single-threaded bottleneck |

```go
// Go: standard library cho hầu hết tasks
import (
    "compress/gzip"
    "crypto/sha256"
    "encoding/json"
    "image"
    "image/png"
)

// Không cần npm install cho basic operations
```

### 7.3 Memory Efficiency

```go
// Go struct: compact memory layout
type User struct {
    ID    int64   // 8 bytes
    Age   int32   // 4 bytes
    Score float64 // 8 bytes
    Name  string  // 16 bytes (pointer + length)
}
// Total: 36 bytes per User (trước alignment padding)

// Slice of 1M users: ~36MB contiguous memory
users := make([]User, 1_000_000)
```

```js
// Node: object overhead lớn
const users = [];
for (let i = 0; i < 1_000_000; i++) {
    users.push({
        id: i,
        age: 25,
        score: 0.0,
        name: "User " + i
    });
}
// Each object: ~100-200 bytes overhead (hidden classes, properties, etc.)
// Total: ~100-200MB cho 1M users
```

### 7.4 Single Binary Deployment

```bash
# Go: deploy anywhere với single binary
GOOS=linux GOARCH=amd64 go build -o server .
scp server user@prod:/usr/local/bin/
ssh user@prod 'systemctl restart myapp'

# Docker image: 5-15MB (FROM scratch)

# Serverless: upload binary, cold start ~1ms
```

```bash
# Node: runtime + dependencies
scp -r . user@prod:/app/
ssh user@prod 'cd /app && npm ci && pm2 restart myapp'

# Docker image: 100-300MB (node:alpine + node_modules)

# Serverless: zip bundle, cold start 100-5000ms
```

### 7.5 Standard Library Completeness

```go
// Go standard library bao gồm HẦU HẾT thứ bạn cần
import (
    // HTTP
    "net/http"
    "net/http/httputil"

    // Encoding
    "encoding/json"
    "encoding/xml"
    "encoding/csv"

    // Crypto
    "crypto/sha256"
    "crypto/aes"
    "crypto/rsa"

    // Compression
    "compress/gzip"
    "compress/zlib"

    // Template
    "html/template"
    "text/template"

    // Testing
    "testing"

    // File system
    "os"
    "io"
    "path/filepath"

    // Concurrency
    "sync"
    "context"

    // Networking
    "net"
    "net/url"

    // Database
    "database/sql"
)
// Không cần Express, body-parser, compression, helmet, etc.
```

```js
// Node: thường cần thêm packages
const express = require('express');           // framework
const bodyParser = require('body-parser');    // middleware
const compression = require('compression');   // middleware
const helmet = require('helmet');             // middleware
const cors = require('cors');                 // middleware
const winston = require('winston');           // logging
const joi = require('joi');                   // validation
// ... thêm nhiều nữa
```

### 7.6 Built-in Tooling

```bash
# Go: tất cả built-in, không cần cài thêm
go fmt ./...           # Code formatter (standard, không debate style)
go vet ./...           # Static analysis
go test ./...          # Testing framework
go test -cover ./...   # Coverage
go test -bench ./...   # Benchmarking
go generate ./...      # Code generation
go doc fmt.Println     # Documentation
go tool pprof          # Profiling
go tool trace          # Tracing
go mod tidy            # Dependency management
go build               # Compilation
go run main.go         # Quick run

# Node: cần cài thêm nhiều tools
# eslint hoặc prettier (style debates)
# jest hoặc mocha (testing)
# nyc hoặc c8 (coverage)
# clinic.js hoặc 0x (profiling)
# nodemon (dev server)
# TypeScript compiler (nếu dùng TS)
# bundler (webpack/vite/esbuild)
```

**Mental model:** Go ecosystem "batteries included". Node ecosystem "choose your own adventure".

---

## 8. Migration Patterns: Node → Go

### 8.1 HTTP Server

```js
// Node: Express
const express = require('express');
const app = express();

app.get('/users/:id', async (req, res) => {
    try {
        const user = await db.findUser(req.params.id);
        res.json(user);
    } catch (err) {
        res.status(500).json({ error: err.message });
    }
});

app.listen(3000);
```

```go
// Go: net/http (standard library, không cần framework)
func main() {
    http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
        id := strings.TrimPrefix(r.URL.Path, "/users/")

        user, err := db.FindUser(r.Context(), id)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(user)
    })

    log.Fatal(http.ListenAndServe(":3000", nil))
}
```

### 8.2 Database Query

```js
// Node: pg/mysql2
const { rows } = await db.query(
    'SELECT * FROM users WHERE id = $1',
    [id]
);
const user = rows[0];
```

```go
// Go: database/sql
var user User
err := db.QueryRowContext(ctx,
    "SELECT id, name, email FROM users WHERE id = $1", id,
).Scan(&user.ID, &user.Name, &user.Email)
if err != nil {
    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    return nil, fmt.Errorf("query user: %w", err)
}
```

### 8.3 Config Loading

```js
// Node: dotenv
require('dotenv').config();
const dbUrl = process.env.DATABASE_URL;
const port = parseInt(process.env.PORT || '3000');
```

```go
// Go: standard library hoặc explicit config
type Config struct {
    DatabaseURL string
    Port        int
}

func LoadConfig() (*Config, error) {
    port, err := strconv.Atoi(getEnvOrDefault("PORT", "3000"))
    if err != nil {
        return nil, fmt.Errorf("invalid PORT: %w", err)
    }
    return &Config{
        DatabaseURL: os.Getenv("DATABASE_URL"),
        Port:        port,
    }, nil
}
```

---

## 9. Checklist Tư Duy Khi Chuyển Từ Node Sang Go

- [ ] **Explicit > Implicit**: Mọi thứ phải rõ ràng, không magic.
- [ ] **Value semantics first**: Struct là value type. Hiểu khi nào copy, khi nào pointer.
- [ ] **Error là value**: Handle mọi error. Không ignore, không swallow.
- [ ] **Goroutine lifecycle**: Luôn track goroutine (WaitGroup, context, channel).
- [ ] **Interface ở consumer side**: Define interface khi bạn CẦN abstraction.
- [ ] **Composition > Inheritance**: Dùng embedding, không phải extends.
- [ ] **Standard library trước**: Hầu hết thứ bạn cần đã có sẵn.
- [ ] **Zero value là useful**: Không cần constructor cho mọi thứ.
- [ ] **Channel cho coordination**: Dùng channels để sync goroutines.
- [ ] **Context cho lifecycle**: Mọi long-running operation cần context.

---

## 10. Rule Thực Dụng

1. **Nếu code Go nhìn "quá thông minh"** → khả năng cao chưa đúng tinh thần Go.
2. **Nếu bạn cần `interface{}` nhiều** → có thể design sai hoặc cần generics.
3. **Nếu goroutine leak** → thiếu context cancellation hoặc channel close.
4. **Nếu error handling chiếm >50% code** → bình thường, đó là Go.
5. **Nếu bạn thấy cần framework lớn** → thử standard library trước, nó đủ mạnh cho hầu hết cases.
6. **Nếu binary quá lớn** → check build flags (`-ldflags="-s -w"`), strip debug symbols.
7. **Nếu test chậm** → có thể đang dùng real database, mock bằng interface.

---

> **TL;DR cho senior Node dev:** Go là ngôn ngữ compiled, explicit, value-oriented với concurrency built-in. Bạn trade flexibility và development speed cho performance, type safety, và deployment simplicity. Đừng cố viết JS trong Go — embrace the explicitness.
