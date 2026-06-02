# HTTP API Và Project Layout Trong Go

> Mục tiêu: Hiểu sâu `net/http`, xây dựng API RESTful hoàn chỉnh, và tổ chức project đúng cách.
> Nếu bạn quen Express.js hoặc NestJS, bài này sẽ map trực tiếp kiến thức đó sang Go.

---

## 1. net/http Deep Dive

### 1.1 ServeMux - Router mặc định

`http.ServeMux` là router built-in của Go. Nó dispatch request dựa trên pattern matching.

```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// Đăng ký handler cho từng route
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User list")
	})

	mux.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id") // Go 1.22+: lấy path param
		fmt.Fprintf(w, "User ID: %s\n", id)
	})

	// Method-based routing (Go 1.22+)
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Create user")
	})

	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		fmt.Fprintf(w, "Get user %s\n", id)
	})

	http.ListenAndServe(":8080", mux)
}
```

**So sánh với Express.js:**

```js
// Express.js
const express = require('express');
const app = express();

app.get('/users', (req, res) => res.send('User list'));
app.get('/users/:id', (req, res) => res.send(`User ID: ${req.params.id}`));
app.post('/users', (req, res) => res.send('Create user'));

app.listen(8080);
```

| Concept | Express.js | Go net/http |
|---------|-----------|-------------|
| Router | `express.Router()` | `http.NewServeMux()` |
| Route params | `req.params.id` | `r.PathValue("id")` |
| Query params | `req.query.page` | `r.URL.Query().Get("page")` |
| Listen | `app.listen(8080)` | `http.ListenAndServe(":8080", mux)` |

### 1.2 Handler Interface và HandlerFunc

Đây là core abstraction của Go HTTP. Mọi thứ đều implement `Handler` interface:

```go
// Handler interface - chỉ có 1 method
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}

// HandlerFunc - adapter biến function thành Handler
type HandlerFunc func(ResponseWriter, *Request)

// HandlerFunc implement Handler interface
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	f(w, r)
}
```

**Ví dụ đầy đủ:**

```go
// Cách 1: Dùng function trực tiếp (với HandleFunc)
mux.HandleFunc("/api/users", listUsersHandler)

// Cách 2: Implement Handler interface
type UserHandler struct {
	service *UserService
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listUsers(w, r)
	case http.MethodPost:
		h.createUser(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	// ...
}

func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	// ...
}

// Đăng ký
mux.Handle("/api/users", &UserHandler{service: userService})
```

**So sánh với Express.js:**

```js
// Express: middleware chain
app.get('/users', authMiddleware, validateMiddleware, listUsers);

// Go: handler là một object hoặc function
// Middleware = function wrap handler (xem phần middleware bên dưới)
```

### 1.3 Request Object

`*http.Request` chứa toàn bộ thông tin về request:

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Method và URL
	fmt.Println("Method:", r.Method)         // GET, POST, PUT, DELETE
	fmt.Println("Path:", r.URL.Path)         // /api/users
	fmt.Println("Raw Query:", r.URL.RawQuery) // page=1&limit=10

	// Query parameters
	page := r.URL.Query().Get("page")        // "1"
	limit := r.URL.Query().Get("limit")      // "10"
	tags := r.URL.Query()["tags"]            // []string{"go", "web"}

	// Headers
	contentType := r.Header.Get("Content-Type")
	auth := r.Header.Get("Authorization")
	allHeaders := r.Header                   // map[string][]string

	// Path parameters (Go 1.22+)
	id := r.PathValue("id")

	// Context (quan trọng!)
	ctx := r.Context()
	userID := ctx.Value("userID") // giá trị do middleware set

	// Body (cho POST/PUT/PATCH)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Cookies
	cookie, err := r.Cookie("session_id")
	if err == nil {
		fmt.Println("Session:", cookie.Value)
	}

	// Form data
	r.ParseForm()
	email := r.FormValue("email")

	// Multipart form (file upload)
	r.ParseMultipartForm(10 << 20) // 10 MB
	file, header, err := r.FormFile("avatar")
	if err == nil {
		defer file.Close()
		fmt.Printf("File: %s, Size: %d\n", header.Filename, header.Size)
	}

	_ = page
	_ = limit
	_ = tags
	_ = contentType
	_ = auth
	_ = userID
}
```

**So sánh với Express.js Request:**

```js
// Express.js                          // Go
req.method                              // r.Method
req.path                                // r.URL.Path
req.query.page                          // r.URL.Query().Get("page")
req.params.id                           // r.PathValue("id")
req.headers['content-type']             // r.Header.Get("Content-Type")
req.body                                // json.NewDecoder(r.Body).Decode(&v)
req.cookies.session_id                  // r.Cookie("session_id")
req.file                                // r.FormFile("avatar")
```

### 1.4 ResponseWriter

`http.ResponseWriter` là interface để gửi response:

```go
func writeResponses(w http.ResponseWriter, r *http.Request) {
	// Set headers TRƯỚC khi write body
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", "abc123")
	w.Header().Set("Cache-Control", "no-cache")

	// Write status code (mặc định là 200)
	w.WriteHeader(http.StatusCreated) // 201

	// Write body
	w.Write([]byte(`{"message": "created"}`))

	// Hoặc dùng fmt.Fprintf
	fmt.Fprintf(w, `{"id": %d, "name": "%s"}`, 1, "John")
}
```

**Lưu ý quan trọng:**
- `Header().Set()` phải gọi TRƯỚC `WriteHeader()` hoặc `Write()`
- `WriteHeader()` chỉ gọi được 1 lần
- Nếu không gọi `WriteHeader()`, lần đầu gọi `Write()` sẽ tự động gửi 200
- Không thể thay đổi status code sau khi đã write body

```go
// ❌ SAI - Header set sau khi đã write
w.Write([]byte("hello"))
w.Header().Set("Content-Type", "application/json") // Quá muộn!

// ✅ ĐÚNG - Set header trước
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
w.Write([]byte(`{"status":"ok"}`))
```

### 1.5 JSON Encoding/Decoding Patterns

**Decode JSON từ request body:**

```go
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"gte=0,lte=150"`
	Password string `json:"password" validate:"required,min=8"`
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest

	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate (dùng thư viện hoặc tự viết)
	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	// Process...
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}
```

**Encode JSON ra response:**

```go
type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := UserResponse{
		ID:    1,
		Name:  "John",
		Email: "john@example.com",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Cách 1: Encode trực tiếp vào ResponseWriter
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, `{"error":"encoding failed"}`, http.StatusInternalServerError)
	}

	// Cách 2: Marshal trước rồi write (nếu cần check lỗi trước)
	// data, err := json.Marshal(user)
	// if err != nil { ... }
	// w.Write(data)
}
```

**JSON Response helper tái sử dụng:**

```go
// response.go - Helper functions
package response

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
	}
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}

func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}
```

```go
// Sử dụng helper trong handler
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := UserResponse{ID: 1, Name: "John"}
	response.OK(w, user) // Clean!
}
```

### 1.6 Form Data Handling

```go
func handleForm(w http.ResponseWriter, r *http.Request) {
	// Phải parse trước khi lấy giá trị
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	// Lấy giá trị
	name := r.FormValue("name")      // "John"
	email := r.FormValue("email")    // "john@example.com"
	tags := r.Form["tags"]           // []string{"go", "web"}

	// POST form (application/x-www-form-urlencoded)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		// fallback to ParseForm for non-multipart
		r.ParseForm()
	}

	_ = name
	_ = email
	_ = tags
}
```

### 1.7 Multipart File Upload

```go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form, max 10 MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}

	// Lấy file từ form field "avatar"
	file, header, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fmt.Printf("Uploaded: %s, Size: %d bytes\n", header.Filename, header.Size)

	// Tạo file trên server
	dst, err := os.Create("./uploads/" + header.Filename)
	if err != nil {
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy nội dung file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}

	// Lấy thêm text field cùng lúc
	description := r.FormValue("description")

	response.OK(w, map[string]string{
		"filename":    header.Filename,
		"description": description,
	})
}
```

**Test upload với curl:**

```bash
curl -X POST http://localhost:8080/upload \
  -F "avatar=@/path/to/photo.jpg" \
  -F "description=My avatar"
```

### 1.8 Middleware Chaining Pattern

Middleware trong Go là function nhận `http.Handler` và trả về `http.Handler`:

```go
package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logging middleware
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter để capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		log.Printf("%s %s %d %s",
			r.Method,
			r.URL.Path,
			rw.statusCode,
			time.Since(start),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Recovery middleware (panic recovery)
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORS middleware
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Auth middleware
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Verify token...
		userID := "user123" // từ token

		// Set user vào context
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

**Chain middleware:**

```go
// Cách 1: Wrap thủ công (đơn giản, dễ hiểu)
mux := http.NewServeMux()
mux.HandleFunc("/api/users", listUsersHandler)

// Chain: Recovery -> CORS -> Logging -> Auth -> handler
handler := Recovery(CORS(Logging(Auth(mux))))

http.ListenAndServe(":8080", handler)

// Cách 2: Helper function
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

handler := Chain(mux, Recovery, CORS, Logging, Auth)
```

**So sánh với Express.js:**

```js
// Express.js
app.use(cors());
app.use(express.json());
app.use(loggingMiddleware);
app.use(authMiddleware);

// Go - thứ tự wrap ngược với thứ tự mong muốn
// Recovery(CORS(Logging(Auth(mux))))
// Recovery chạy đầu tiên (outermost), Auth chạy cuối cùng (gần handler nhất)
```

### 1.9 Router Group Patterns

**Dùng `net/http` thuần (Go 1.22+):**

```go
func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /auth/login", loginHandler)
	mux.HandleFunc("POST /auth/register", registerHandler)

	// Protected routes - wrap group với auth middleware
	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/users", listUsersHandler)
	protected.HandleFunc("POST /api/users", createUserHandler)
	protected.HandleFunc("GET /api/users/{id}", getUserHandler)
	protected.HandleFunc("PUT /api/users/{id}", updateUserHandler)
	protected.HandleFunc("DELETE /api/users/{id}", deleteUserHandler)

	// Wrap protected group với auth
	mux.Handle("/api/", Auth(protected))

	return mux
}
```

**Dùng chi router (giống Express nhất):**

```go
import "github.com/go-chi/chi/v5"

func setupRouter() *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CORS)

	// Public routes
	r.Get("/health", healthHandler)
	r.Post("/auth/login", loginHandler)

	// Route groups (giống Express.Router)
	r.Route("/api", func(r chi.Router) {
		r.Use(AuthMiddleware) // Auth chỉ cho group này

		r.Route("/users", func(r chi.Router) {
			r.Get("/", listUsersHandler)
			r.Post("/", createUserHandler)

			// Subrouter cho /:id
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", getUserHandler)
				r.Put("/", updateUserHandler)
				r.Delete("/", deleteUserHandler)
			})
		})

		r.Route("/posts", func(r chi.Router) {
			r.Get("/", listPostsHandler)
			r.Post("/", createPostHandler)
		})
	})

	return r
}
```

**Dùng gorilla/mux:**

```go
import "github.com/gorilla/mux"

func setupRouter() *mux.Router {
	r := mux.NewRouter()

	// Middleware
	r.Use(loggingMiddleware)

	// Routes
	r.HandleFunc("/api/users", listUsersHandler).Methods("GET")
	r.HandleFunc("/api/users", createUserHandler).Methods("POST")
	r.HandleFunc("/api/users/{id:[0-9]+}", getUserHandler).Methods("GET")
	r.HandleFunc("/api/users/{id:[0-9]+}", updateUserHandler).Methods("PUT")

	// Subrouter cho versioning
	v1 := r.PathPrefix("/api/v1").Subrouter()
	v1.Use(authMiddleware)
	v1.HandleFunc("/users", listUsersV1Handler).Methods("GET")

	return r
}
```

---

## 2. HTTP Client

### 2.1 http.Client Configuration

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// Tạo client với cấu hình
	client := &http.Client{
		Timeout: 30 * time.Second, // Timeout tổng cho toàn bộ request
		Transport: &http.Transport{
			MaxIdleConns:        100,              // Max connections trong pool
			MaxIdleConnsPerHost: 10,               // Max idle connections per host
			IdleConnTimeout:     90 * time.Second, // Thời gian giữ idle connection
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	// Tạo request
	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/users", nil)
	if err != nil {
		panic(err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer token123")

	// Gửi request
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Đọc response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Body: %s\n", string(body))
}
```

### 2.2 Request với Context

Context giúp cancel request khi cần (timeout, user disconnect, etc.):

```go
import (
	"context"
	"net/http"
	"time"
)

// Timeout context - tự cancel sau 5 giây
func fetchWithTimeout(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err // timeout hoặc cancel
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Cancel context khi cần (ví dụ: user disconnect)
func fetchWithCancel(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

// Trong HTTP handler - request bị cancel khi client disconnect
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// r.Context() tự động cancel khi client disconnect
	data, err := fetchWithCancel(r.Context(), "https://api.example.com/data")
	if err != nil {
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}

	w.Write(data)
}
```

### 2.3 So sánh với axios/fetch

```js
// Node.js với axios
const axios = require('axios');

const response = await axios.get('https://api.example.com/users', {
  headers: { 'Authorization': 'Bearer token' },
  timeout: 5000,
});
console.log(response.data);

// Node.js với fetch
const response = await fetch('https://api.example.com/users', {
  headers: { 'Authorization': 'Bearer token' },
  signal: AbortSignal.timeout(5000),
});
const data = await response.json();
```

```go
// Go equivalent
client := &http.Client{Timeout: 5 * time.Second}

req, _ := http.NewRequest("GET", "https://api.example.com/users", nil)
req.Header.Set("Authorization", "Bearer token")

resp, err := client.Do(req)
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

var data map[string]interface{}
json.NewDecoder(resp.Body).Decode(&data)
```

### 2.4 Connection Pooling

Go tự động reuse connections qua `http.Transport`. Không cần config thêm:

```go
// Default transport đã có connection pooling
client := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

// Reuse client cho nhiều requests
for i := 0; i < 100; i++ {
    resp, err := client.Get("https://api.example.com/data")
    // Connection được reuse tự động
}
```

### 2.5 Retry Patterns

Go không có built-in retry. Tự implement:

```go
package httpclient

import (
	"net/http"
	"time"
)

func DoWithRetry(client *http.Client, req *http.Request, maxRetries int) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i <= maxRetries; i++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		if resp != nil {
			resp.Body.Close()
		}

		if i < maxRetries {
			// Exponential backoff
			backoff := time.Duration(1<<uint(i)) * time.Second
			time.Sleep(backoff)
		}
	}

	return nil, err
}

// Sử dụng
func fetchWithRetry(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return DoWithRetry(client, req, 3) // Retry tối đa 3 lần
}
```

---

## 3. Project Layout Patterns

### 3.1 Standard Layout

```
myapp/
├── cmd/
│   └── api/
│       └── main.go          # Entrypoint - chỉ bootstrap
├── internal/
│   ├── handler/             # HTTP handlers (controllers)
│   │   ├── user.go
│   │   └── health.go
│   ├── service/             # Business logic
│   │   └── user.go
│   ├── repository/          # Data access layer
│   │   └── user.go
│   ├── model/               # Domain models
│   │   └── user.go
│   └── middleware/           # HTTP middleware
│       ├── auth.go
│       └── logging.go
├── pkg/                     # Code có thể import từ ngoài
│   └── response/
│       └── json.go
├── config/                  # Config files
│   └── config.go
├── go.mod
├── go.sum
└── README.md
```

**Ý nghĩa từng thư mục:**

| Thư mục | Vai trò | Tương đương NestJS |
|----------|---------|---------------------|
| `cmd/api/` | Entrypoint, bootstrap | `main.ts` |
| `internal/handler/` | HTTP layer | `controllers/` |
| `internal/service/` | Business logic | `services/` |
| `internal/repository/` | Data access | `repositories/` |
| `internal/model/` | Domain types | `entities/` |
| `internal/middleware/` | Middleware | `middleware/` |
| `pkg/` | Public utilities | `shared/` |

### 3.2 Khi nào dùng Flat vs Layered Layout

**Flat layout - App nhỏ (< 10 files):**

```
myapp/
├── main.go
├── user.go
├── user_test.go
├── product.go
├── product_test.go
├── middleware.go
└── go.mod
```

**Layered layout - App lớn:**

```
myapp/
├── cmd/
│   ├── api/main.go
│   └── worker/main.go      # Có thể có nhiều entrypoint
├── internal/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   └── ...
└── ...
```

**Rule of thumb:**
- Bắt đầu flat, refactor khi cần
- Khi file > 10, tách package
- Khi package > 5, tách internal/

### 3.3 Dependency Injection Patterns

**Không dùng framework - Constructor injection:**

```go
// internal/repository/user.go
package repository

type UserRepository interface {
	FindByID(id int) (*User, error)
	FindAll() ([]*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id int) error
}

type inMemoryUserRepo struct {
	users map[int]*User
	mu    sync.RWMutex
}

func NewInMemoryUserRepository() UserRepository {
	return &inMemoryUserRepo{
		users: make(map[int]*User),
	}
}

// internal/service/user.go
package service

type UserService struct {
	repo repository.UserRepository
}

// Constructor injection - inject dependency qua constructor
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// internal/handler/user.go
package handler

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{service: svc}
}
```

**Bootstrap trong main.go:**

```go
// cmd/api/main.go
func main() {
	// Tạo dependencies
	userRepo := repository.NewInMemoryUserRepository()
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// Setup router
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/users", userHandler.ListUsers)
	mux.HandleFunc("POST /api/users", userHandler.CreateUser)
	mux.HandleFunc("GET /api/users/{id}", userHandler.GetUser)

	// Start server
	log.Println("Server starting on :8080")
	http.ListenAndServe(":8080", mux)
}
```

### 3.4 Configuration Management

**Dùng environment variables (đơn giản nhất):**

```go
package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	Debug       bool
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/mydb"),
		JWTSecret:   getEnv("JWT_SECRET", "secret"),
		Debug:       getEnvBool("DEBUG", false),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return defaultValue
}
```

**Dùng config file (YAML/JSON):**

```go
import (
	"encoding/json"
	"os"
)

type Config struct {
	Port     int    `json:"port"`
	DB       DBConfig `json:"db"`
	JWTSecret string `json:"jwt_secret"`
}

type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
```

### 3.5 Graceful Shutdown

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server trong goroutine
	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Chờ signal để shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Timeout 30 giây để hoàn thành request đang xử lý
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
```

### 3.6 Health Check Patterns

```go
package handler

import (
	"net/http"
	"time"
)

type HealthChecker struct {
	startTime time.Time
	checks    map[string]func() error
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		startTime: time.Now(),
		checks:    make(map[string]func() error),
	}
}

func (hc *HealthChecker) RegisterCheck(name string, check func() error) {
	hc.checks[name] = check
}

// Simple health check
func (hc *HealthChecker) LiveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"alive"}`))
}

// Deep health check (kiểm tra dependencies)
func (hc *HealthChecker) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	checks := make(map[string]string)
	allOK := true

	for name, check := range hc.checks {
		if err := check(); err != nil {
			checks[name] = err.Error()
			allOK = false
		} else {
			checks[name] = "ok"
		}
	}

	if !allOK {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "not ready",
			"checks": checks,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ready",
		"uptime":    time.Since(hc.startTime).String(),
		"checks":    checks,
	})
}

// Đăng ký trong main
func setupHealthRoutes(mux *http.ServeMux) {
	hc := handler.NewHealthChecker()

	// Đăng ký các dependency checks
	hc.RegisterCheck("database", func() error {
		return db.Ping()
	})
	hc.RegisterCheck("redis", func() error {
		return redisClient.Ping().Err()
	})

	mux.HandleFunc("/health/live", hc.LiveHandler)
	mux.HandleFunc("/health/ready", hc.ReadyHandler)
}
```

### 3.7 So sánh với NestJS Module System

| NestJS | Go |
|--------|-----|
| `@Module({})` decorator | Manual wiring trong main.go |
| `@Injectable()` | Constructor injection |
| `@Controller()` | Handler struct với methods |
| `@Service()` | Service struct |
| Module imports | Truyền dependencies qua constructor |
| `forRoot()` config | `config.Load()` |

```typescript
// NestJS
@Module({
  imports: [DatabaseModule],
  controllers: [UserController],
  providers: [UserService, UserRepository],
})
export class UserModule {}
```

```go
// Go - explicit wiring trong main.go
func main() {
    db := database.Connect(cfg.DatabaseURL)
    userRepo := repository.NewUserRepository(db)
    userService := service.NewUserService(userRepo)
    userHandler := handler.NewUserHandler(userService)
    
    mux := http.NewServeMux()
    mux.HandleFunc("GET /api/users", userHandler.ListUsers)
    // ...
}
```

---

## 4. Real-world API Example - Full CRUD

### 4.1 User Model

```go
// internal/model/user.go
package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
	Age   *int    `json:"age,omitempty"`
}
```

### 4.2 User Repository

```go
// internal/repository/user.go
package repository

import (
	"fmt"
	"sync"
	"time"

	"myapp/internal/model"
)

type UserRepository interface {
	FindByID(id int) (*model.User, error)
	FindAll(page, limit int) ([]*model.User, error)
	Create(user *model.User) error
	Update(user *model.User) error
	Delete(id int) error
	FindByEmail(email string) (*model.User, error)
}

// In-memory implementation
type inMemoryUserRepo struct {
	mu      sync.RWMutex
	users   map[int]*model.User
	nextID  int
}

func NewInMemoryUserRepository() UserRepository {
	return &inMemoryUserRepo{
		users:  make(map[int]*model.User),
		nextID: 1,
	}
}

func (r *inMemoryUserRepo) FindByID(id int) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found: %d", id)
	}
	return user, nil
}

func (r *inMemoryUserRepo) FindAll(page, limit int) ([]*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*model.User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, u)
	}

	// Simple pagination
	start := (page - 1) * limit
	if start >= len(users) {
		return []*model.User{}, nil
	}

	end := start + limit
	if end > len(users) {
		end = len(users)
	}

	return users[start:end], nil
}

func (r *inMemoryUserRepo) Create(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check email uniqueness
	for _, u := range r.users {
		if u.Email == user.Email {
			return fmt.Errorf("email already exists: %s", user.Email)
		}
	}

	user.ID = r.nextID
	r.nextID++
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	r.users[user.ID] = user
	return nil
}

func (r *inMemoryUserRepo) Update(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[user.ID]; !ok {
		return fmt.Errorf("user not found: %d", user.ID)
	}

	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return nil
}

func (r *inMemoryUserRepo) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[id]; !ok {
		return fmt.Errorf("user not found: %d", id)
	}

	delete(r.users, id)
	return nil
}

func (r *inMemoryUserRepo) FindByEmail(email string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found with email: %s", email)
}
```

### 4.3 User Service

```go
// internal/service/user.go
package service

import (
	"fmt"
	"strings"

	"myapp/internal/model"
	"myapp/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(id int) (*model.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user id: %d", id)
	}
	return s.repo.FindByID(id)
}

func (s *UserService) ListUsers(page, limit int) ([]*model.User, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.FindAll(page, limit)
}

func (s *UserService) CreateUser(req model.CreateUserRequest) (*model.User, error) {
	// Validation
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(req.Email) == "" {
		return nil, fmt.Errorf("email is required")
	}
	if !strings.Contains(req.Email, "@") {
		return nil, fmt.Errorf("invalid email format")
	}
	if req.Age < 0 || req.Age > 150 {
		return nil, fmt.Errorf("invalid age: %d", req.Age)
	}

	user := &model.User{
		Name:  strings.TrimSpace(req.Name),
		Email: strings.TrimSpace(req.Email),
		Age:   req.Age,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUser(id int, req model.UpdateUserRequest) (*model.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, fmt.Errorf("name cannot be empty")
		}
		user.Name = strings.TrimSpace(*req.Name)
	}
	if req.Email != nil {
		if !strings.Contains(*req.Email, "@") {
			return nil, fmt.Errorf("invalid email format")
		}
		user.Email = strings.TrimSpace(*req.Email)
	}
	if req.Age != nil {
		if *req.Age < 0 || *req.Age > 150 {
			return nil, fmt.Errorf("invalid age: %d", *req.Age)
		}
		user.Age = *req.Age
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) DeleteUser(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid user id: %d", id)
	}
	return s.repo.Delete(id)
}
```

### 4.4 User Handler

```go
// internal/handler/user.go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"myapp/internal/model"
	"myapp/internal/service"
	"myapp/pkg/response"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{service: svc}
}

// GET /api/users?page=1&limit=10
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	users, err := h.service.ListUsers(page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(w, map[string]interface{}{
		"data":  users,
		"page":  page,
		"limit": limit,
	})
}

// GET /api/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.service.GetUser(id)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	response.OK(w, user)
}

// POST /api/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	defer r.Body.Close()

	user, err := h.service.CreateUser(req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Created(w, user)
}

// PUT /api/users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	defer r.Body.Close()

	user, err := h.service.UpdateUser(id, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.OK(w, user)
}

// DELETE /api/users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.service.DeleteUser(id); err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

### 4.5 Router Setup và Main

```go
// cmd/api/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myapp/internal/handler"
	"myapp/internal/middleware"
	"myapp/internal/repository"
	"myapp/internal/service"
)

func main() {
	// 1. Create dependencies
	userRepo := repository.NewInMemoryUserRepository()
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// 2. Setup router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// User routes
	mux.HandleFunc("GET /api/users", userHandler.ListUsers)
	mux.HandleFunc("POST /api/users", userHandler.CreateUser)
	mux.HandleFunc("GET /api/users/{id}", userHandler.GetUser)
	mux.HandleFunc("PUT /api/users/{id}", userHandler.UpdateUser)
	mux.HandleFunc("DELETE /api/users/{id}", userHandler.DeleteUser)

	// 3. Apply middleware
	handler := middleware.Recovery(
		middleware.CORS(
			middleware.Logging(mux),
		),
	)

	// 4. Create server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Start server
	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// 6. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
```

### 4.6 Middleware (Logging, Recovery, Auth)

```go
// internal/middleware/middleware.go
package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"
)

// Logging - log mỗi request
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		log.Printf("[%s] %s %s %d %s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			rw.statusCode,
			time.Since(start),
		)
	})
}

// Recovery - bắt panic
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORS - cho phép cross-origin requests
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Auth - xác thực JWT
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Verify JWT (đơn giản hóa - trong thực tế dùng thư viện JWT)
		userID, err := verifyJWT(token)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Set user ID vào context
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func verifyJWT(token string) (string, error) {
	// Placeholder - trong thực tế verify JWT signature
	return "user123", nil
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
```

### 4.7 API Testing với curl

```bash
# Health check
curl http://localhost:8080/health

# Create user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","age":30}'

# List users
curl http://localhost:8080/api/users?page=1&limit=10

# Get user by ID
curl http://localhost:8080/api/users/1

# Update user
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe","age":25}'

# Delete user
curl -X DELETE http://localhost:8080/api/users/1
```

---

## Summary

| Concept | Go | Express.js / NestJS |
|---------|-----|---------------------|
| Router | `http.NewServeMux()` | `express.Router()` |
| Handler | Function hoặc struct implement `Handler` | Controller method |
| Middleware | `func(http.Handler) http.Handler` | `app.use(fn)` |
| Request body | `json.NewDecoder(r.Body).Decode(&v)` | `req.body` |
| Response | `json.NewEncoder(w).Encode(v)` | `res.json(v)` |
| Path params | `r.PathValue("id")` | `req.params.id` |
| Query params | `r.URL.Query().Get("page")` | `req.query.page` |
| DI | Constructor injection | `@Injectable()` |
| Config | `os.Getenv()` | `@nestjs/config` |
| Graceful shutdown | Manual `server.Shutdown()` | Built-in |

---

## Tips

- Handler nên mỏng - chỉ parse request, gọi service, trả response
- Validate input ở service layer, không phải handler
- Luôn pass `context.Context` xuống service/repository
- Return error rõ ràng, không swallow errors
- Dùng `defer r.Body.Close()` khi đọc body
- Set headers trước khi write response
- Dùng `sync.RWMutex` cho concurrent access trong repository
