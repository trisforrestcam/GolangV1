# Fast Track Roadmap - Học Go Trong 7 Ngày (Cho Senior Node.js Dev)

> **Mục tiêu:** Từ Node.js developer → Go developer trong 1 tuần. Mỗi ngày ~4-6 giờ.
>
> **Yêu cầu:** Đã đọc [[01-go-mental-model-for-nodejs-dev]] để hiểu tổng quan.

---

## Ngày 1: Foundation (Đọc 01 + 02)

### Buổi sáng: Setup + Mental Model

```bash
# 1. Cài Go
# macOS: brew install go
# Linux: wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
# Windows: https://go.dev/dl/

# 2. Verify installation
go version
go env GOPATH
go env GOROOT
```

```bash
# 3. Tạo project đầu tiên
mkdir hello-go && cd hello-go
go mod init hello-go
```

```go
// main.go
package main

import "fmt"

func main() {
    fmt.Println("Hello from Go!")
    
    // Thử các biến
    name := "Tris"
    age := 28
    fmt.Printf("Name: %s, Age: %d\n", name, age)
}
```

```bash
go run .
```

### Buổi chiều: Syntax Drills

Viết `playground.go` và thử MỌI thứ:

```go
package main

import "fmt"

func main() {
    // 1. Variables
    var explicit int = 10
    inferred := "hello"
    const pi = 3.14
    _, _ = explicit, inferred
    _ = pi

    // 2. Collections
    nums := []int{1, 2, 3, 4, 5}
    nums = append(nums, 6)
    
    scores := map[string]int{
        "alice": 95,
        "bob":   87,
    }
    
    // 3. Control flow
    for i, v := range nums {
        if v > 3 {
            fmt.Printf("nums[%d] = %d\n", i, v)
        }
    }
    
    // 4. Struct
    type User struct {
        ID   int
        Name string
    }
    
    u := User{ID: 1, Name: "Alice"}
    fmt.Println(u)
    
    // 5. Function với multiple returns
    result, err := divide(10, 3)
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Result:", result)
    }
    
    // 6. Pointer
    x := 42
    p := &x
    *p = 100
    fmt.Println("x =", x) // 100
}

func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}
```

### Checklist Ngày 1

- [ ] Cài Go, chạy được `go run .`
- [ ] Hiểu `var`, `:=`, `const`
- [ ] Thử `slice`, `map`, `for`, `if`, `switch`
- [ ] Viết `struct` với methods
- [ ] Hiểu `pointer` basics (`&`, `*`)
- [ ] Chạy `go fmt ./...`

---

## Ngày 2: Struct + Interface + Error (Đọc 03)

### Buổi sáng: Struct Patterns

```go
// model/user.go
package main

import (
    "fmt"
    "time"
)

// Base fields
type AuditFields struct {
    CreatedAt time.Time
    UpdatedAt time.Time
}

// User với composition
type User struct {
    ID    int
    Name  string
    Email string
    AuditFields
}

// Constructor
func NewUser(name, email string) *User {
    now := time.Now()
    return &User{
        Name:  name,
        Email: email,
        AuditFields: AuditFields{
            CreatedAt: now,
            UpdatedAt: now,
        },
    }
}

// Methods
func (u *User) UpdateEmail(email string) {
    u.Email = email
    u.UpdatedAt = time.Now()
}

func (u User) String() string {
    return fmt.Sprintf("User{id=%d, name=%s, email=%s}", u.ID, u.Name, u.Email)
}
```

### Buổi chiều: Interface + Error Handling

```go
// interface pattern
type UserRepository interface {
    FindByID(id int) (*User, error)
    Create(user *User) error
}

// In-memory implementation
type InMemoryUserRepo struct {
    users map[int]*User
    nextID int
}

func (r *InMemoryUserRepo) FindByID(id int) (*User, error) {
    user, ok := r.users[id]
    if !ok {
        return nil, fmt.Errorf("user %d: %w", id, ErrNotFound)
    }
    return user, nil
}

// Sentinel error
var ErrNotFound = errors.New("not found")

// Custom error type
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation '%s': %s", e.Field, e.Message)
}
```

### Checklist Ngày 2

- [ ] Viết `User` struct với embedded `AuditFields`
- [ ] Viết `NewUser` constructor
- [ ] Viết `UserRepository` interface
- [ ] Implement `InMemoryUserRepo`
- [ ] Dùng `errors.Is()` và `errors.As()`
- [ ] Tạo custom error type

---

## Ngày 3: HTTP API (Đọc 05)

### Buổi sáng: Minimal API

```go
// main.go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "strings"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

var users = map[int]*User{}
var nextID = 1

func main() {
    mux := http.NewServeMux()
    
    mux.HandleFunc("GET /health", healthHandler)
    mux.HandleFunc("GET /api/users", listUsersHandler)
    mux.HandleFunc("POST /api/users", createUserHandler)
    mux.HandleFunc("GET /api/users/{id}", getUserHandler)
    
    log.Println("Server on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
    list := make([]*User, 0, len(users))
    for _, u := range users {
        list = append(list, u)
    }
    json.NewEncoder(w).Encode(list)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    
    user := &User{ID: nextID, Name: req.Name, Email: req.Email}
    users[nextID] = user
    nextID++
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }
    
    user, ok := users[id]
    if !ok {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }
    
    json.NewEncoder(w).Encode(user)
}
```

### Buổi chiều: Add Middleware + Test với curl

```bash
# Test API
curl http://localhost:8080/health
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
curl http://localhost:8080/api/users
curl http://localhost:8080/api/users/1
```

### Checklist Ngày 3

- [ ] Viết API `/health` chạy được
- [ ] Viết CRUD cho `User`
- [ ] Test với curl hoặc Postman
- [ ] Thêm logging middleware
- [ ] Thêm error handling

---

## Ngày 4: Testing (Đọc 06)

### Buổi sáng: Basic Tests

```go
// user_test.go
package main

import "testing"

func TestNewUser(t *testing.T) {
    user := NewUser("Alice", "alice@example.com")
    
    if user.Name != "Alice" {
        t.Errorf("expected name Alice, got %s", user.Name)
    }
    if user.Email != "alice@example.com" {
        t.Errorf("expected email alice@example.com, got %s", user.Email)
    }
    if user.ID != 0 {
        t.Errorf("expected ID 0, got %d", user.ID)
    }
}

func TestUpdateEmail(t *testing.T) {
    user := NewUser("Alice", "old@example.com")
    user.UpdateEmail("new@example.com")
    
    if user.Email != "new@example.com" {
        t.Errorf("expected new@example.com, got %s", user.Email)
    }
}
```

### Buổi chiều: Table-driven Tests

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name    string
        a, b    float64
        want    float64
        wantErr bool
    }{
        {"positive", 10, 3, 3.333, false},
        {"zero divisor", 10, 0, 0, true},
        {"negative", -10, 3, -3.333, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := divide(tt.a, tt.b)
            if (err != nil) != tt.wantErr {
                t.Errorf("divide() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && abs(got-tt.want) > 0.001 {
                t.Errorf("divide() = %v, want %v", got, tt.want)
            }
        })
    }
}

func abs(x float64) float64 {
    if x < 0 {
        return -x
    }
    return x
}
```

```bash
go test -v ./...
go test -cover ./...
go test -race ./...
```

### Checklist Ngày 4

- [ ] Viết basic test cho `NewUser`
- [ ] Viết table-driven tests
- [ ] Chạy `go test -v ./...`
- [ ] Chạy `go test -cover ./...`
- [ ] Viết test cho handler (dùng `httptest`)

---

## Ngày 5: Concurrency (Đọc 04)

### Buổi sáng: Goroutine + WaitGroup

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    var wg sync.WaitGroup
    
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            fmt.Printf("Worker %d starting\n", id)
            time.Sleep(time.Second)
            fmt.Printf("Worker %d done\n", id)
        }(i)
    }
    
    wg.Wait()
    fmt.Println("All workers done!")
}
```

### Buổi chiều: Channel + Select

```go
// Worker pool
func workerPool(jobs <-chan int, results chan<- int, workers int) {
    var wg sync.WaitGroup
    
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for job := range jobs {
                fmt.Printf("Worker %d processing job %d\n", id, job)
                time.Sleep(100 * time.Millisecond)
                results <- job * 2
            }
        }(i)
    }
    
    wg.Wait()
    close(results)
}

// Context timeout
func fetchWithTimeout(ctx context.Context, url string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    return io.ReadAll(resp.Body)
}
```

### Checklist Ngày 5

- [ ] Viết 5 goroutine xử lý song song
- [ ] Dùng `WaitGroup` để đợi
- [ ] Dùng `channel` để truyền kết quả
- [ ] Dùng `select` với timeout
- [ ] Dùng `context.WithTimeout`
- [ ] Chạy `go run -race .`

---

## Ngày 6: Project Layout + Refactor

### Buổi sáng: Tách theo Layer

```
myapp/
├── cmd/api/main.go
├── internal/
│   ├── handler/user.go
│   ├── service/user.go
│   ├── repository/user.go
│   └── model/user.go
├── go.mod
└── go.sum
```

### Buổi chiều: Wire Dependencies trong main.go

```go
// cmd/api/main.go
func main() {
    // Dependencies
    userRepo := repository.NewInMemoryUserRepository()
    userService := service.NewUserService(userRepo)
    userHandler := handler.NewUserHandler(userService)
    
    // Router
    mux := http.NewServeMux()
    mux.HandleFunc("GET /api/users", userHandler.ListUsers)
    mux.HandleFunc("POST /api/users", userHandler.CreateUser)
    mux.HandleFunc("GET /api/users/{id}", userHandler.GetUser)
    
    // Server
    server := &http.Server{
        Addr:    ":8080",
        Handler: middleware.Recovery(middleware.Logging(mux)),
    }
    
    // Graceful shutdown
    go func() {
        log.Println("Server starting on :8080")
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    server.Shutdown(ctx)
}
```

### Checklist Ngày 6

- [ ] Tách code thành `cmd/`, `internal/`
- [ ] Viết `UserRepository` interface
- [ ] Implement `InMemoryUserRepo`
- [ ] Wire dependencies trong `main.go`
- [ ] Thêm graceful shutdown
- [ ] Pass `context.Context` xuyên suốt

---

## Ngày 7: Mini Project

### Option 1: Todo API

```
Todo API:
- POST /todos         → Create todo
- GET /todos          → List todos
- GET /todos/{id}     → Get todo
- PUT /todos/{id}     → Update todo
- DELETE /todos/{id}  → Delete todo

Học được: CRUD, JSON, routing, testing
```

### Option 2: URL Shortener

```
URL Shortener:
- POST /shorten       → Tạo short URL
- GET /{shortID}      → Redirect to original

Học được: map/storage, HTTP redirect, ID generation
```

### Option 3: Concurrent Web Crawler

```
Web Crawler:
- Input: list of URLs
- Process: fetch concurrently với worker pool
- Output: collected results

Học được: goroutine, channel, worker pool, context cancellation
```

### Option 4: Rate-Limited API Client

```
API Client:
- Fetch nhiều URLs concurrently
- Rate limit: max N requests/second
- Retry with exponential backoff
- Timeout per request

Học được: ticker, semaphore, retry pattern, error handling
```

### Checklist Ngày 7

- [ ] Chọn 1 mini project
- [ ] Viết code + tests
- [ ] Refactor theo project layout
- [ ] Chạy `go test ./...`
- [ ] Chạy `go vet ./...`
- [ ] Chạy `go build ./...`

---

## Kết Quả Mong Đợi Sau Roadmap

Sau 7 ngày, bạn chưa phải Go expert, nhưng đủ mạnh để:

- ✅ Đọc và hiểu code backend Go phổ biến
- ✅ Viết REST API nhỏ đến vừa
- ✅ Sử dụng goroutine và channel cơ bản
- ✅ Viết tests (table-driven, benchmark)
- ✅ Hiểu idiom cốt lõi của Go
- ✅ Tiếp tục học sâu: database, gRPC, profiling, worker system

---

## Thứ Cần Tránh Nếu Muốn Học Nhanh

| ❌ Tránh | ✅ Nên làm |
|----------|-----------|
| Học framework quá sớm | Học standard library trước |
| Cố biến Go thành TypeScript | Embrace Go idioms |
| Architecture quá nặng | Start simple, refactor khi cần |
| Dùng nhiều package | Thử stdlib trước |
| Bỏ qua error handling | Luôn `if err != nil` |
| Interface mọi thứ | Concrete type trước |

---

## Resources Tiếp Tục Học

### Official
- [Go Tour](https://go.dev/tour/) - Interactive tutorial
- [Go by Example](https://gobyexample.com/) - Code examples
- [Effective Go](https://go.dev/doc/effective_go) - Best practices
- [Go Documentation](https://go.dev/doc/) - Official docs

### Books
- "The Go Programming Language" - Donovan & Kernighan
- "Concurrency in Go" - Katherine Cox-Buday
- "Go in Action" - Kennedy, Ketelsen, St. Martin

### Practice
- [Exercism Go Track](https://exercism.org/tracks/go) - Coding exercises
- [LeetCode](https://leetcode.com/) - Algorithms in Go
- [Go Playground](https://go.dev/play/) - Online editor

---

## Workflow Học Mỗi Concept Mới

1. **Đọc doc** - Đọc phần tương ứng trong series này
2. **Viết demo** - Tạo file nhỏ, chạy `go run .`
3. **Viết test** - Viết test cho demo
4. **So sánh** - "Tôi làm thế này trong Node.js?"
5. **Refactor** - Viết lại theo Go idioms
6. **Commit** - Lưu lại để review sau

---

> **TL;DR:** Học Go trong 7 ngày với roadmap này. Mỗi ngày focus vào 1-2 concepts, viết code thực tế, và luôn so sánh với Node.js để hiểu sâu hơn.

