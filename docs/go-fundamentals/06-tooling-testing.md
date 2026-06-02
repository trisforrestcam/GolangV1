# Tooling, Testing và Workflow Trong Go

> Mục tiêu: Thành thạo Go toolchain, viết test chuyên nghiệp, debug và profile hiệu quả.
> So sánh trực tiếp với npm/Jest/ESLint để bạn map kiến thức nhanh.

---

## 1. Go Toolchain

### 1.1 go run, go build, go install

```bash
# Chạy file trực tiếp (compile + run, không tạo binary)
go run main.go
go run . # Chạy package trong thư mục hiện tại

# Compile thành binary
go build                    # Tạo binary tên giống thư mục
go build -o myapp           # Đặt tên binary
go build -o ./bin/myapp     # Output vào thư mục khác

# Chạy binary đã compile
./myapp

# Install - compile và copy binary vào $GOPATH/bin
go install                  # Binary tên giống module
go install ./cmd/api        # Install specific package
```

**So sánh với Node.js:**

| Go | Node.js |
|----|---------|
| `go run main.go` | `node main.js` |
| `go build -o myapp` | `npm run build` (tùy framework) |
| `go install` | `npm install -g .` |
| Binary standalone | Cần Node.js runtime |

**Điểm mạnh Go:** Binary standalone, không cần runtime. Deploy chỉ cần copy 1 file.

### 1.2 go mod - Dependency Management

```bash
# Khởi tạo module mới (tương đương npm init)
go mod init github.com/username/myapp

# Thêm dependency (tương đương npm install)
go get github.com/go-chi/chi/v5
go get github.com/go-chi/chi/v5@v5.0.10    # Specific version
go get github.com/go-chi/chi/v5@latest      # Latest version

# Cleanup unused dependencies (tương đương npm prune)
go mod tidy

# Download dependencies vào vendor/
go mod vendor

# Xem dependency graph
go mod graph

# Verify checksums
go mod verify

# Xem giải thích tại sao cần dependency
go mod why github.com/some/package

# Edit module path
go mod edit -module new/path
```

**So sánh với npm:**

| Go | npm |
|----|-----|
| `go.mod` + `go.sum` | `package.json` + `package-lock.json` |
| `go mod init` | `npm init` |
| `go get pkg` | `npm install pkg` |
| `go mod tidy` | `npm prune` + `npm install` |
| `go mod vendor` | `npm ci` (với lockfile) |
| `go mod graph` | `npm ls` |
| `go.sum` (checksums) | `package-lock.json` (integrity) |

**go.mod file:**

```go
module github.com/username/myapp

go 1.22

require (
    github.com/go-chi/chi/v5 v5.0.10
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/lib/pq v1.10.9
)

require (
    // Indirect dependencies (của dependencies)
    golang.org/x/crypto v0.17.0 // indirect
)
```

**go.sum file:**
- Chứa cryptographic hashes của dependencies
- Tương tự `package-lock.json` nhưng chỉ chứa checksums
- Commit vào git để đảm bảo reproducible builds

### 1.3 go fmt - Non-negotiable trong Go Culture

```bash
# Format tất cả files trong package
go fmt ./...

# Format specific file
go fmt main.go

# Check formatting mà không sửa (dùng trong CI)
gofmt -l .                  # Liệt kê files cần format
gofmt -d .                  # Hiển thị diff

# go fmt = gofmt -l -w (tự động sửa)
```

**Quan trọng:**
- Go community CỰC KỲ strict về formatting
- Không có debates về tabs vs spaces, braces placement, etc.
- `go fmt` là tool duy nhất - mọi người dùng cùng style
- Code không format = code không review được

**So sánh với JavaScript:**

```bash
# JavaScript - nhiều tools khác nhau
eslint --fix .
prettier --write .
standard --fix

# Go - chỉ 1 tool
go fmt ./...
```

**Setup editor để auto-format:**

```json
// VS Code settings.json
{
    "go.formatTool": "goimports",
    "editor.formatOnSave": true,
    "[go]": {
        "editor.defaultFormatter": "golang.go"
    }
}
```

### 1.4 go vet - Static Analysis

```go
// go vet tìm các lỗi phổ biến mà compiler không bắt

// Ví dụ 1: Printf format mismatch
fmt.Printf("Hello %s", 123)  // ❌ vet warning: %s với int

// Ví dụ 2: Unreachable code
func example() {
    return
    fmt.Println("never")  // ❌ vet warning: unreachable
}

// Ví dụ 3: Struct tag errors
type User struct {
    Name string `json:"name`
    // ❌ vet warning: missing closing quote
}

// Ví dụ 4: Unused result của certain functions
fmt.Println("hello")  // Kết quả bị bỏ qua (ok)
strconv.Atoi("123")   // ❌ vet warning: result unused
```

```bash
# Chạy vet
go vet ./...

# Vet là 1 phần của go test
go test ./...  # Tự động chạy vet trước khi test
```

### 1.5 go doc - Documentation

```bash
# Xem doc cho package
go doc fmt
go doc net/http

# Xem doc cho function/type
go doc fmt.Printf
go doc http.Handler
go doc http.ResponseWriter

# Xem doc với examples
go doc -all fmt

# Start local doc server (deprecated nhưng vẫn hữu ích)
godoc -http=:6060
# Mở http://localhost:6060
```

**Viết doc comments:**

```go
// UserService xử lý business logic liên quan đến User.
// Nó validate input, gọi repository, và trả về kết quả.
type UserService struct {
    repo UserRepository
}

// GetUser tìm user theo ID.
// Trả về error nếu user không tồn tại hoặc id không hợp lệ.
//
// Example:
//
//	user, err := svc.GetUser(1)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(user.Name)
func (s *UserService) GetUser(id int) (*User, error) {
    // ...
}
```

### 1.6 Cross-Compilation

Go compile native code cho mọi platform. Không cần cross-compile tools:

```bash
# Xem danh sách platforms hỗ trợ
go tool dist list

# Compile cho Linux từ macOS/Windows
GOOS=linux GOARCH=amd64 go build -o myapp-linux

# Compile cho Windows
GOOS=windows GOARCH=amd64 go build -o myapp.exe

# Compile cho macOS ARM (M1/M2)
GOOS=darwin GOARCH=arm64 go build -o myapp-mac-arm

# Compile cho macOS Intel
GOOS=darwin GOARCH=amd64 go build -o myapp-mac-intel

# Compile cho ARM (Raspberry Pi)
GOOS=linux GOARCH=arm GOARM=7 go build -o myapp-arm

# Compile cho Alpine Linux (dùng musl)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o myapp-alpine
```

**Dockerfile cho Go (multi-stage build):**

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp ./cmd/api

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/myapp /usr/local/bin/
EXPOSE 8080
CMD ["myapp"]
```

**So sánh với Node.js:**

```dockerfile
# Node.js Dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
EXPOSE 8080
CMD ["node", "server.js"]
# Image size: ~150-300MB

# Go Dockerfile (multi-stage)
# Image size: ~10-20MB (Alpine + binary)
```

### 1.7 Build Tags

Build tags cho phép compile conditionally:

```go
// file: db_postgres.go
//go:build postgres

package db

import "github.com/lib/pq"

func Connect() (*sql.DB, error) {
    return sql.Open("postgres", os.Getenv("DATABASE_URL"))
}
```

```go
// file: db_sqlite.go
//go:build sqlite

package db

import "github.com/mattn/go-sqlite3"

func Connect() (*sql.DB, error) {
    return sql.Open("sqlite3", "./data.db")
}
```

```bash
# Build với postgres
go build -tags postgres -o myapp-pg .

# Build với sqlite
go build -tags sqlite -o myapp-sqlite .

# Build với nhiều tags
go build -tags "postgres,jsoniter" -o myapp .
```

**Common use cases:**
- Database drivers khác nhau
- Platform-specific code (linux, darwin, windows)
- Feature flags
- Test builds

```go
// file: platform_linux.go
//go:build linux

package platform

import "syscall"

func GetOpenFileLimit() int {
    var rlimit syscall.Rlimit
    syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit)
    return int(rlimit.Cur)
}
```

### 1.8 ldflags - Version Injection

Inject version info vào binary tại build time:

```go
// main.go
package main

import "fmt"

// Biến sẽ được set qua ldflags
var (
    version   = "dev"
    commit    = "unknown"
    buildTime = "unknown"
)

func main() {
    fmt.Printf("Version: %s\n", version)
    fmt.Printf("Commit: %s\n", commit)
    fmt.Printf("Built: %s\n", buildTime)
}
```

```bash
# Build với version info
go build -ldflags "\
  -X main.version=1.0.0 \
  -X main.commit=$(git rev-parse --short HEAD) \
  -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
" -o myapp .

# Chạy
./myapp
# Output:
# Version: 1.0.0
# Commit: abc1234
# Built: 2024-01-15T10:30:00Z
```

**Makefile helper:**

```makefile
VERSION := $(shell git describe --tags --always)
COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -ldflags "\
  -X main.version=$(VERSION) \
  -X main.commit=$(COMMIT) \
  -X main.buildTime=$(BUILD_TIME)"

build:
	go build $(LDFLAGS) -o bin/myapp ./cmd/api

build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/myapp-linux ./cmd/api
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/myapp-mac ./cmd/api
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/myapp.exe ./cmd/api
```

---

## 2. Testing - Exhaustive

### 2.1 Testing Package Deep Dive

**Cấu trúc file test:**
- File test: `user_test.go` (cùng package với `user.go`)
- Function test: `func TestXxx(t *testing.T)` (bắt đầu bằng `Test`)
- Package: `_test` suffix (black-box) hoặc cùng package (white-box)

```go
// user.go
package user

type User struct {
    Name  string
    Email string
}

func ValidateEmail(email string) bool {
    return strings.Contains(email, "@")
}
```

```go
// user_test.go - White-box test (cùng package)
package user

import "testing"

func TestValidateEmail(t *testing.T) {
    if !ValidateEmail("test@example.com") {
        t.Error("expected valid email")
    }
    if ValidateEmail("invalid") {
        t.Error("expected invalid email")
    }
}
```

```go
// user_test.go - Black-box test (khác package)
package user_test

import (
    "testing"
    "myapp/user"
)

func TestValidateEmail(t *testing.T) {
    if !user.ValidateEmail("test@example.com") {
        t.Error("expected valid email")
    }
}
```

**Testing.T methods:**

```go
func TestExample(t *testing.T) {
    // Log (chỉ hiện khi test fail hoặc -v)
    t.Log("this is a log message")

    // Fail - đánh dấu fail nhưng tiếp tục chạy
    t.Fail()

    // FailNow - đánh dấu fail và dừng test ngay
    t.FailNow()

    // Error - Log + Fail
    t.Error("something went wrong")

    // Fatal - Log + FailNow
    t.Fatal("cannot continue")

    // Skip - bỏ qua test
    if runtime.GOOS == "windows" {
        t.Skip("skipping on windows")
    }

    // Helper - đánh dấu là helper function
    t.Helper()

    // Cleanup - chạy sau khi test xong
    t.Cleanup(func() {
        os.Remove("temp-file")
    })

    // Parallel - chạy song song với tests khác
    t.Parallel()
}
```

### 2.2 Table-Driven Tests

Đây là pattern idiomatic nhất trong Go:

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
   }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"zero", 0, 0, 0},
        {"mixed signs", -2, 3, 1},
        {"large numbers", 1000000, 2000000, 3000000},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.expected)
            }
        })
    }
}
```

**Table-driven test cho HTTP handler:**

```go
func TestCreateUser(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        wantStatus int
        wantErr    string
    }{
        {
            name:       "valid input",
            body:       `{"name":"John","email":"john@example.com"}`,
            wantStatus: http.StatusCreated,
        },
        {
            name:       "missing name",
            body:       `{"email":"john@example.com"}`,
            wantStatus: http.StatusBadRequest,
            wantErr:    "name is required",
        },
        {
            name:       "invalid email",
            body:       `{"name":"John","email":"invalid"}`,
            wantStatus: http.StatusBadRequest,
            wantErr:    "invalid email",
        },
        {
            name:       "invalid JSON",
            body:       `{invalid`,
            wantStatus: http.StatusBadRequest,
            wantErr:    "invalid JSON",
        },
        {
            name:       "empty body",
            body:       "",
            wantStatus: http.StatusBadRequest,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            repo := repository.NewInMemoryUserRepository()
            svc := service.NewUserService(repo)
            handler := handler.NewUserHandler(svc)

            // Create request
            req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")
            rr := httptest.NewRecorder()

            // Execute
            handler.CreateUser(rr, req)

            // Assert
            if rr.Code != tt.wantStatus {
                t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
            }

            if tt.wantErr != "" {
                var resp map[string]string
                json.NewDecoder(rr.Body).Decode(&resp)
                if resp["error"] != tt.wantErr {
                    t.Errorf("error = %q, want %q", resp["error"], tt.wantErr)
                }
            }
        })
    }
}
```

### 2.3 Subtests

```go
func TestUserService(t *testing.T) {
    repo := repository.NewInMemoryUserRepository()
    svc := service.NewUserService(repo)

    // Tạo user trước
    user, err := svc.CreateUser(model.CreateUserRequest{
        Name:  "John",
        Email: "john@example.com",
    })
    if err != nil {
        t.Fatalf("failed to create user: %v", err)
    }

    // Subtests share state
    t.Run("GetUser", func(t *testing.T) {
        got, err := svc.GetUser(user.ID)
        if err != nil {
            t.Fatalf("GetUser() error = %v", err)
        }
        if got.Name != "John" {
            t.Errorf("Name = %q, want %q", got.Name, "John")
        }
    })

    t.Run("UpdateUser", func(t *testing.T) {
        newName := "Jane"
        got, err := svc.UpdateUser(user.ID, model.UpdateUserRequest{
            Name: &newName,
        })
        if err != nil {
            t.Fatalf("UpdateUser() error = %v", err)
        }
        if got.Name != "Jane" {
            t.Errorf("Name = %q, want %q", got.Name, "Jane")
        }
    })

    t.Run("DeleteUser", func(t *testing.T) {
        err := svc.DeleteUser(user.ID)
        if err != nil {
            t.Fatalf("DeleteUser() error = %v", err)
        }

        // Verify deleted
        _, err = svc.GetUser(user.ID)
        if err == nil {
            t.Error("expected error after delete")
        }
    })
}
```

### 2.4 Test Fixtures

```go
// testdata/ directory - Go convention
// testdata/ được tự động ignore bởi go tools

func TestParseConfig(t *testing.T) {
    // Đọc fixture file
    data, err := os.ReadFile("testdata/config.json")
    if err != nil {
        t.Fatalf("failed to read fixture: %v", err)
    }

    cfg, err := ParseConfig(data)
    if err != nil {
        t.Fatalf("ParseConfig() error = %v", err)
    }

    if cfg.Port != 8080 {
        t.Errorf("Port = %d, want 8080", cfg.Port)
    }
}

// Tạo fixture helper
func writeFixture(t *testing.T, name string, data []byte) {
    t.Helper()
    path := filepath.Join("testdata", name)
    if err := os.WriteFile(path, data, 0644); err != nil {
        t.Fatalf("failed to write fixture: %v", err)
    }
}
```

**Test fixtures với golden files:**

```go
func TestRenderTemplate(t *testing.T) {
    got := RenderTemplate(data)

    golden := filepath.Join("testdata", t.Name()+".golden")

    if *update {
        // Cập nhật golden file: go test -update
        os.WriteFile(golden, []byte(got), 0644)
        return
    }

    want, err := os.ReadFile(golden)
    if err != nil {
        t.Fatalf("failed to read golden file: %v", err)
    }

    if got != string(want) {
        t.Errorf("mismatch:\ngot:\n%s\nwant:\n%s", got, string(want))
    }
}
```

### 2.5 Test Helpers

```go
// testutil/helpers.go
package testutil

import (
    "os"
    "testing"
)

// Helper function - t.Helper() giúp error message trỏ đúng đến caller
func AssertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func AssertEqual[T comparable](t *testing.T, got, want T) {
    t.Helper()
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}

func AssertNotEqual[T comparable](t *testing.T, got, want T) {
    t.Helper()
    if got == want {
        t.Errorf("got %v, want not %v", got, want)
    }
}

func AssertContains(t *testing.T, s, substr string) {
    t.Helper()
    if !strings.Contains(s, substr) {
        t.Errorf("expected %q to contain %q", s, substr)
    }
}

// Setup temp directory
func TempDir(t *testing.T) string {
    t.Helper()
    dir, err := os.MkdirTemp("", "test-*")
    if err != nil {
        t.Fatalf("failed to create temp dir: %v", err)
    }
    t.Cleanup(func() { os.RemoveAll(dir) })
    return dir
}

// Setup test database
func SetupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("failed to open test db: %v", err)
    }
    t.Cleanup(func() { db.Close() })

    // Run migrations
    _, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)`)
    if err != nil {
        t.Fatalf("failed to create table: %v", err)
    }

    return db
}
```

### 2.6 TestMain - Setup/Teardown

```go
// main_test.go
package myapp_test

import (
    "os"
    "testing"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
    // Setup - chạy trước tất cả tests
    var err error
    testDB, err = setupTestDatabase()
    if err != nil {
        log.Fatalf("failed to setup test db: %v", err)
    }

    // Chạy tất cả tests
    code := m.Run()

    // Teardown - chạy sau tất cả tests
    testDB.Close()
    cleanupTestFiles()

    os.Exit(code)
}

func setupTestDatabase() (*sql.DB, error) {
    // Setup test database
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        return nil, err
    }

    // Create tables
    _, err = db.Exec(`
        CREATE TABLE users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            email TEXT UNIQUE NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        return nil, err
    }

    return db, nil
}

func cleanupTestFiles() {
    os.RemoveAll("./test-uploads")
}

// Test functions có thể dùng testDB
func TestGetUser(t *testing.T) {
    // testDB đã sẵn sàng
    repo := repository.NewSQLUserRepository(testDB)
    // ...
}
```

### 2.7 Parallel Tests

```go
func TestParallel(t *testing.T) {
    tests := []struct {
        name string
        // ...
    }{
        {"case1", /* ... */},
        {"case2", /* ... */},
        {"case3", /* ... */},
    }

    for _, tt := range tests {
        tt := tt // Capture range variable (Go < 1.22)
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // Chạy song song

            // Test logic...
            // Không được share mutable state
        })
    }
}
```

**Lưu ý khi dùng Parallel:**
- Mỗi subtest phải độc lập
- Không share mutable state
- Dùng `t.Parallel()` ngay đầu subtest
- Go 1.22+ không cần capture range variable

### 2.8 Benchmarks

```go
// Benchmark function bắt đầu bằng Benchmark
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}

// Benchmark với setup
func BenchmarkSortLargeSlice(b *testing.B) {
    // Setup (không tính vào benchmark)
    data := make([]int, 10000)
    for i := range data {
        data[i] = rand.Intn(10000)
    }

    b.ResetTimer() // Reset timer sau setup

    for i := 0; i < b.N; i++ {
        // Copy data vì sort modifies slice
        d := make([]int, len(data))
        copy(d, data)
        sort.Ints(d)
    }
}

// Benchmark với allocations tracking
func BenchmarkJSONMarshal(b *testing.B) {
    user := User{Name: "John", Email: "john@example.com"}

    b.ReportAllocs() // Track allocations

    for i := 0; i < b.N; i++ {
        json.Marshal(user)
    }
}

// Sub-benchmarks
func BenchmarkParseInt(b *testing.B) {
    for _, size := range []int{10, 100, 1000, 10000} {
        b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
            s := strings.Repeat("9", size)
            for i := 0; i < b.N; i++ {
                strconv.Atoi(s)
            }
        })
    }
}
```

```bash
# Chạy benchmarks
go test -bench=.                    # Tất cả benchmarks
go test -bench=BenchmarkAdd         # Specific benchmark
go test -bench=. -benchmem          # Hiển thị memory allocations
go test -bench=. -count=5           # Chạy 5 lần
go test -bench=. -benchtime=2s      # Chạy mỗi benchmark 2 giây

# Compare benchmarks (dùng benchstat)
go test -bench=. -count=10 > old.txt
# ... code changes ...
go test -bench=. -count=10 > new.txt
benchstat old.txt new.txt
```

**Output mẫu:**

```
BenchmarkAdd-8          1000000000           0.25 ns/op        0 B/op          0 allocs/op
BenchmarkSortLargeSlice-8     1000       1500000 ns/op        0 B/op          0 allocs/op
BenchmarkJSONMarshal-8     5000000           350 ns/op         128 B/op         1 allocs/op
```

### 2.9 Fuzzing (Go 1.18+)

Fuzzing tự động tạo test inputs để tìm bugs:

```go
// Fuzz function bắt đầu bằng Fuzz
func FuzzReverse(f *testing.F) {
    // Seed corpus - inputs ban đầu
    f.Add("hello")
    f.Add("world")
    f.Add("")
    f.Add("a")

    // Fuzz function
    f.Fuzz(func(t *testing.T, input string) {
        // Reverse twice should return original
        reversed := Reverse(input)
        doubleReversed := Reverse(reversed)

        if input != doubleReversed {
            t.Errorf("Reverse(Reverse(%q)) = %q, want %q", input, doubleReversed, input)
        }
    })
}

func FuzzParseUser(f *testing.F) {
    f.Add([]byte(`{"name":"John","age":30}`))
    f.Add([]byte(`{}`))

    f.Fuzz(func(t *testing.T, data []byte) {
        user, err := ParseUser(data)
        if err != nil {
            return // Invalid input is ok
        }

        // User should be valid after parsing
        if user.Name == "" && user.Age != 0 {
            t.Errorf("inconsistent user: %+v", user)
        }
    })
}
```

```bash
# Chạy fuzzing
go test -fuzz=FuzzReverse
go test -fuzz=FuzzReverse -fuzztime=30s    # Chạy 30 giây
go test -fuzz=FuzzReverse -fuzztime=1000x  # Chạy 1000 iterations

# Fuzzing với corpus trong testdata/fuzz/
go test -fuzz=FuzzReverse -fuzztime=5m
```

### 2.10 Test Coverage

```bash
# Chạy tests với coverage
go test -cover ./...

# Tạo coverage report
go test -coverprofile=coverage.out ./...

# Xem coverage trong terminal
go tool cover -func=coverage.out

# Xem coverage HTML
go tool cover -html=coverage.out -o coverage.html

# Coverage cho specific package
go test -coverprofile=coverage.out ./internal/service/

# Merge coverage từ nhiều packages
go test -coverpkg=./... -coverprofile=coverage.out ./...
```

**Coverage output mẫu:**

```
myapp/internal/service    coverage: 85.2% of statements
myapp/internal/handler    coverage: 72.1% of statements
myapp/internal/repository coverage: 90.5% of statements
```

**So sánh với Jest:**

```bash
# Jest
npm test -- --coverage
# Tạo coverage/lcov-report/

# Go
go test -cover ./...
go tool cover -html=coverage.out
```

### 2.11 So sánh với Jest/Mocha

| Feature | Jest/Mocha | Go testing |
|---------|-----------|------------|
| File naming | `*.test.js`, `*.spec.js` | `*_test.go` |
| Test function | `describe`, `it`, `test` | `func TestXxx(t *testing.T)` |
| Assertions | `expect(x).toBe(y)` | `if x != y { t.Error() }` |
| Setup/Teardown | `beforeEach`, `afterEach` | `t.Cleanup()`, `TestMain` |
| Mocking | `jest.mock()` | Interfaces + mock implementations |
| Coverage | `--coverage` | `-cover` |
| Snapshot | `toMatchSnapshot()` | Golden files |
| Parallel | `--runInBand` (disable) | `t.Parallel()` |
| Watch mode | `--watch` | `air`, `CompileDaemon` |

---

## 3. Mocking và Test Doubles

### 3.1 Interface-based Mocking

Go dùng interfaces để mock dependencies:

```go
// internal/repository/user.go
package repository

type UserRepository interface {
    FindByID(id int) (*model.User, error)
    Create(user *model.User) error
}

// Mock implementation trong test
type MockUserRepository struct {
    FindByIDFunc func(id int) (*model.User, error)
    CreateFunc   func(user *model.User) error
}

func (m *MockUserRepository) FindByID(id int) (*model.User, error) {
    return m.FindByIDFunc(id)
}

func (m *MockUserRepository) Create(user *model.User) error {
    return m.CreateFunc(user)
}
```

**Sử dụng mock trong test:**

```go
func TestGetUser(t *testing.T) {
    mockRepo := &MockUserRepository{
        FindByIDFunc: func(id int) (*model.User, error) {
            return &model.User{ID: id, Name: "John"}, nil
        },
    }

    svc := service.NewUserService(mockRepo)

    user, err := svc.GetUser(1)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Name != "John" {
        t.Errorf("Name = %q, want %q", user.Name, "John")
    }
}

func TestGetUser_NotFound(t *testing.T) {
    mockRepo := &MockUserRepository{
        FindByIDFunc: func(id int) (*model.User, error) {
            return nil, fmt.Errorf("not found")
        },
    }

    svc := service.NewUserService(mockRepo)

    _, err := svc.GetUser(999)
    if err == nil {
        t.Fatal("expected error")
    }
}
```

### 3.2 Testify Mock

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock với testify
type MockUserRepo struct {
    mock.Mock
}

func (m *MockUserRepo) FindByID(id int) (*model.User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) Create(user *model.User) error {
    args := m.Called(user)
    return args.Error(0)
}

// Test với testify assertions
func TestGetUser(t *testing.T) {
    mockRepo := new(MockUserRepo)

    // Setup expectations
    mockRepo.On("FindByID", 1).Return(&model.User{
        ID:   1,
        Name: "John",
    }, nil)

    svc := service.NewUserService(mockRepo)

    user, err := svc.GetUser(1)

    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, "John", user.Name)
    assert.Equal(t, 1, user.ID)

    // Verify expectations were met
    mockRepo.AssertExpectations(t)
}

func TestGetUser_Error(t *testing.T) {
    mockRepo := new(MockUserRepo)
    mockRepo.On("FindByID", 999).Return(nil, fmt.Errorf("not found"))

    svc := service.NewUserService(mockRepo)

    _, err := svc.GetUser(999)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")

    mockRepo.AssertExpectations(t)
}
```

### 3.3 gomock

```bash
# Install gomock
go install go.uber.org/mock/mockgen@latest

# Generate mock từ interface
mockgen -source=internal/repository/user.go -destination=mocks/mock_user_repo.go -package=mocks
```

```go
// Generated mock usage
func TestGetUser(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockUserRepository(ctrl)

    // Setup expectations
    mockRepo.EXPECT().
        FindByID(1).
        Return(&model.User{ID: 1, Name: "John"}, nil).
        Times(1)

    svc := service.NewUserService(mockRepo)

    user, err := svc.GetUser(1)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Name != "John" {
        t.Errorf("Name = %q, want %q", user.Name, "John")
    }
}
```

### 3.4 httptest Package

Test HTTP handlers mà không cần start server:

```go
import "net/http/httptest"

func TestHealthHandler(t *testing.T) {
    // Tạo request
    req := httptest.NewRequest(http.MethodGet, "/health", nil)

    // Tạo ResponseRecorder
    rr := httptest.NewRecorder()

    // Gọi handler
    healthHandler(rr, req)

    // Assertions
    if rr.Code != http.StatusOK {
        t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
    }

    expected := `{"status":"ok"}`
    if rr.Body.String() != expected {
        t.Errorf("body = %q, want %q", rr.Body.String(), expected)
    }
}

func TestCreateUserHandler(t *testing.T) {
    body := `{"name":"John","email":"john@example.com"}`
    req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()

    // Setup handler với mock service
    mockSvc := &MockUserService{
        CreateFunc: func(req model.CreateUserRequest) (*model.User, error) {
            return &model.User{ID: 1, Name: req.Name, Email: req.Email}, nil
        },
    }

    handler := handler.NewUserHandler(mockSvc)
    handler.CreateUser(rr, req)

    if rr.Code != http.StatusCreated {
        t.Errorf("status = %d, want %d", rr.Code, http.StatusCreated)
    }

    var user model.User
    json.NewDecoder(rr.Body).Decode(&user)
    if user.Name != "John" {
        t.Errorf("Name = %q, want %q", user.Name, "John")
    }
}

// Test full HTTP server (integration test)
func TestAPIIntegration(t *testing.T) {
    // Setup
    repo := repository.NewInMemoryUserRepository()
    svc := service.NewUserService(repo)
    h := handler.NewUserHandler(svc)

    mux := http.NewServeMux()
    mux.HandleFunc("POST /api/users", h.CreateUser)
    mux.HandleFunc("GET /api/users/{id}", h.GetUser)

    // Tạo test server
    server := httptest.NewServer(mux)
    defer server.Close()

    // Gọi API
    resp, err := http.Post(
        server.URL+"/api/users",
        "application/json",
        strings.NewReader(`{"name":"John","email":"john@example.com"}`),
    )
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
    }
}
```

### 3.5 SQL Mock Patterns

```go
import "github.com/DATA-DOG/go-sqlmock"

func TestUserRepository_FindByID(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("failed to create sqlmock: %v", err)
    }
    defer db.Close()

    repo := repository.NewSQLUserRepository(db)

    // Setup expectations
    rows := sqlmock.NewRows([]string{"id", "name", "email"}).
        AddRow(1, "John", "john@example.com")

    mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
        WithArgs(1).
        WillReturnRows(rows)

    // Execute
    user, err := repo.FindByID(1)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Assert
    if user.Name != "John" {
        t.Errorf("Name = %q, want %q", user.Name, "John")
    }

    // Verify all expectations were met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("unmet expectations: %v", err)
    }
}

func TestUserRepository_Create(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("failed to create sqlmock: %v", err)
    }
    defer db.Close()

    repo := repository.NewSQLUserRepository(db)

    // Setup expectation
    mock.ExpectExec("INSERT INTO users").
        WithArgs("John", "john@example.com").
        WillReturnResult(sqlmock.NewResult(1, 1))

    // Execute
    user := &model.User{Name: "John", Email: "john@example.com"}
    err = repo.Create(user)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if user.ID != 1 {
        t.Errorf("ID = %d, want 1", user.ID)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("unmet expectations: %v", err)
    }
}
```

---

## 4. Debugging và Profiling

### 4.1 Delve Debugger

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug binary
dlv debug ./cmd/api

# Debug test
dlv test ./internal/service/ -- -test.run TestGetUser

# Debug running process
dlv attach <PID>

# Remote debugging
dlv debug --headless --listen=:2345 --api-version=2 ./cmd/api
```

**Delve commands:**

```
(dlv) break main.main           # Set breakpoint
(dlv) break service.go:42       # Break at line
(dlv) break GetUser             # Break at function
(dlv) continue                  # Continue execution
(dlv) next                      # Next line
(dlv) step                      # Step into
(dlv) stepout                   # Step out
(dlv) print user                # Print variable
(dlv) locals                    # Print all locals
(dlv) args                      # Print function args
(dlv) goroutines                # List goroutines
(dlv) goroutine 1 switch        # Switch to goroutine
(dlv) stack                     # Print stack trace
(dlv) restart                   # Restart program
(dlv) quit                      # Exit
```

**VS Code配置:**

```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/api",
            "env": {},
            "args": []
        },
        {
            "name": "Debug Test",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}/internal/service",
            "args": ["-test.run", "TestGetUser"]
        }
    ]
}
```

### 4.2 pprof - Profiling

```go
import (
    "net/http"
    _ "net/http/pprof" // Import để register handlers
)

func main() {
    // Start pprof server (có thể dùng port khác)
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // Your application code...
}
```

```bash
# CPU Profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory Profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine Profile
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Block Profile (blocking operations)
go tool pprof http://localhost:6060/debug/pprof/block

# Mutex Profile (contention)
go tool pprof http://localhost:6060/debug/pprof/mutex
```

**pprof commands:**

```
(pprof) top                    # Top functions by CPU usage
(pprof) top20                  # Top 20
(pprof) list GetUser           # Annotated source code
(pprof) web                    # Open in browser (requires graphviz)
(pprof) png > profile.png      # Generate image
(pprof) save profile.pb.gz     # Save for later analysis
```

**Chương trình CLI profiling:**

```go
import (
    "os"
    "runtime/pprof"
)

func main() {
    // CPU profile
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // Your code here...
    doExpensiveWork()

    // Memory profile
    f2, _ := os.Create("mem.prof")
    pprof.WriteHeapProfile(f2)
    f2.Close()
}
```

```bash
# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

### 4.3 Trace Tool

```bash
# Generate trace
curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=5

# Or trong code
import "runtime/trace"

func main() {
    f, _ := os.Create("trace.out")
    trace.Start(f)
    defer trace.Stop()

    // Your code...
}
```

```bash
# View trace
go tool trace trace.out
# Opens in browser - xem goroutines, GC, network, etc.
```

### 4.4 go tool compile

```bash
# Xem assembly
go tool compile -S main.go

# Compile với optimization info
go build -gcflags="-m" ./...          # Escape analysis
go build -gcflags="-m -m" ./...       # More verbose

# Disable optimizations (for debugging)
go build -gcflags="-N -l" ./...
```

### 4.5 Race Detector

Go có built-in race detector để tìm data races:

```go
// Code có race condition
func main() {
    counter := 0

    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter++ // ❌ Race condition!
        }()
    }
    wg.Wait()
    fmt.Println(counter)
}
```

```bash
# Chạy với race detector
go run -race main.go
go test -race ./...
go build -race -o myapp && ./myapp
```

**Output khi detect race:**

```
WARNING: DATA RACE
Write at 0x00c0000b4010 by goroutine 8:
  main.main.func1()
      /path/to/main.go:12 +0x4e

Previous write at 0x00c0000b4010 by goroutine 7:
  main.main.func1()
      /path/to/main.go:12 +0x4e
```

**Fix race condition:**

```go
func main() {
    var counter int64

    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            atomic.AddInt64(&counter, 1) // ✅ Thread-safe
        }()
    }
    wg.Wait()
    fmt.Println(counter)
}
```

---

## 5. Linting và Code Quality

### 5.1 golangci-lint

```bash
# Install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Chạy
golangci-lint run

# Chạy với config
golangci-lint run --config .golangci.yml
```

**`.golangci.yml` config:**

```yaml
run:
  timeout: 5m
  go: "1.22"

linters:
  enable:
    - errcheck        # Kiểm tra unchecked errors
    - gosimple        # Code simplification
    - govet           # Go vet
    - ineffassign     # Ineffective assignments
    - staticcheck     # Static analysis
    - unused          # Unused code
    - gocritic        # Code critic
    - gofmt           # Formatting
    - goimports       # Import ordering
    - misspell        # Spelling errors
    - gocyclo         # Cyclomatic complexity
    - dupl            # Code duplication
    - gosec           # Security issues
    - bodyclose       # HTTP response body close
    - exhaustive      # Exhaustive switch
    - noctx           # HTTP without context

linters-settings:
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gocyclo
        - dupl
```

### 5.2 Staticcheck

```bash
# Install
go install honnef.co/go/tools/cmd/staticcheck@latest

# Chạy
staticcheck ./...
```

Staticcheck tìm:
- Bugs
- Performance issues
- Style issues
- Deprecated API usage

### 5.3 gosec - Security Linter

```bash
# Install
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Chạy
gosec ./...

# Generate report
gosec -fmt=json -out=results.json ./...
```

**gosec tìm:**
- Hardcoded credentials
- SQL injection
- Command injection
- Insecure TLS configuration
- Weak cryptography
- File path traversal

### 5.4 CI/CD Patterns

**GitHub Actions:**

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Download dependencies
        run: go mod download

      - name: Format check
        run: |
          if [ -n "$(gofmt -l .)" ]; then
            echo "Files not formatted:"
            gofmt -l .
            exit 1
          fi

      - name: Vet
        run: go vet ./...

      - name: Lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

      - name: Test
        run: go test -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: coverage.out

  build:
    needs: test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          go build -ldflags="-s -w" -o myapp-${{ matrix.goos }}-${{ matrix.goarch }} ./cmd/api

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: myapp-${{ matrix.goos }}-${{ matrix.goarch }}
          path: myapp-*
```

---

## 6. Dependency Management

### 6.1 go get vs go install

```bash
# go get - thêm dependency vào go.mod
go get github.com/go-chi/chi/v5
go get github.com/go-chi/chi/v5@v5.0.10
go get github.com/go-chi/chi/v5@latest

# go install - install binary tool
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/gopls@latest

# go get cũng dùng để update
go get -u ./...                    # Update all dependencies
go get -u github.com/go-chi/chi/v5 # Update specific
go get -u=patch ./...              # Update patch versions only
```

### 6.2 Module Versioning

Go modules dùng semantic versioning:

```go
// go.mod
require (
    github.com/foo/bar v1.2.3        // Release version
    github.com/foo/bar v1.2.4-0.20230101120000-abcdef123456  // Pseudo-version (pre-release)
    github.com/foo/bar v0.0.0-20230101120000-abcdef123456     // No release yet
)
```

```bash
# Xem versions available
go list -m -versions github.com/go-chi/chi/v5

# Xem version hiện tại
go list -m github.com/go-chi/chi/v5

# Pin version
go get github.com/go-chi/chi/v5@v5.0.10

# Upgrade minor
go get github.com/go-chi/chi/v5@v5.1.0

# Upgrade major (thay đổi import path)
go get github.com/foo/bar/v2
```

**Major version convention:**

```go
// v0.x.x và v1.x.x - cùng import path
import "github.com/foo/bar"

// v2.x.x+ - phải thêm /v2 vào import path
import "github.com/foo/bar/v2"

// go.mod
module github.com/foo/bar/v2
```

### 6.3 Replace Directives

```go
// go.mod
module myapp

require (
    github.com/some/dependency v1.0.0
)

// Replace với local path (development)
replace github.com/some/dependency => ../dependency

// Replace với fork
replace github.com/some/dependency => github.com/myfork/dependency v1.0.1

// Replace với specific version
replace github.com/some/dependency v1.0.0 => github.com/some/dependency v1.0.1
```

**Khi nào dùng replace:**
- Local development (chưa publish)
- Fork cần fix bug
- Test với local changes

**Lưu ý:** `replace` trong `go.mod` không ảnh hưởng khi package được import bởi project khác.

### 6.4 Private Modules

```bash
# Configure private module access
go env -w GOPRIVATE=github.com/mycompany/*
go env -w GOPRIVATE=github.com/mycompany/*,gitlab.com/myteam/*

# Hoặc dùng GONOSUMCHECK
go env -w GONOSUMCHECK=github.com/mycompany/*

# Git configuration cho private repos
git config --global url."git@github.com:".insteadOf "https://github.com/"
```

### 6.5 So sánh với npm/yarn

| Feature | npm/yarn | Go modules |
|---------|----------|------------|
| Config file | `package.json` | `go.mod` |
| Lock file | `package-lock.json` / `yarn.lock` | `go.sum` |
| Add dependency | `npm install pkg` | `go get pkg` |
| Remove unused | `npm prune` | `go mod tidy` |
| Install all | `npm install` | `go mod download` |
| Global install | `npm install -g` | `go install` |
| Workspaces | `npm workspaces` | Go workspaces (1.18+) |
| Private packages | `.npmrc` config | `GOPRIVATE` env |
| Version ranges | `^1.2.3` | `v1.2.3` (exact) |
| Offline mode | `npm ci` | `go mod vendor` |

**Go workspaces (Go 1.18+):**

```bash
# Tạo workspace
go work init ./module1 ./module2

# Add module vào workspace
go work use ./module3

# Sync dependencies
go work sync
```

```go
// go.work
go 1.22

use (
    ./api
    ./shared
    ./worker
)
```

---

## Tips và Best Practices

### Workflow học nhanh

1. Viết file demo nhỏ
2. Chạy `go run .`
3. Viết test
4. Refactor
5. So sánh với Node.js

### Checklist trước khi commit

```bash
# Format
go fmt ./...

# Vet
go vet ./...

# Lint
golangci-lint run

# Test
go test -race ./...

# Build
go build ./cmd/api
```

### Makefile template

```makefile
.PHONY: build test lint clean

build:
	go build -ldflags="-s -w" -o bin/api ./cmd/api

test:
	go test -race -cover ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...
	goimports -w .

clean:
	rm -rf bin/

run: build
	./bin/api

dev:
	go run ./cmd/api

docker:
	docker build -t myapp .
```

### Thứ nên ưu tiên trong 1-2 tuần đầu

1. `go fmt ./...` - luôn format code
2. `go test ./...` - viết test cho mọi function
3. `go vet ./...` - check lỗi cơ bản
4. Table-driven tests - pattern quan trọng nhất
5. Interface-based mocking - testable code
6. `go test -race` - tìm race conditions
7. Coverage report - biết code nào chưa test
