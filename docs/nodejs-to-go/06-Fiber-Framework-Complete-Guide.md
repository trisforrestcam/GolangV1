# 🖼️ Fiber v2 Framework - Hiểu Sâu từ Gốc đến Ngọn

> **Mục tiêu:** Không chỉ biết "làm thế nào" mà hiểu "TẠI SAO làm vậy"
> **Dành cho:** NodeJS dev muốn hiểu sâu Fiber framework

---

## 🧠 Phần 1: Hiểu bản chất Fiber là gì

### 1.1 Tại sao Fiber tồn tại?

Vấn đề của Go standard library (`net/http`):

```go
// Go standard library - Rất dài dòng
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    // 1. Phải check method thủ công
    if r.Method != "GET" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // 2. Phải parse query thủ công
    name := r.URL.Query().Get("name")
    if name == "" {
        http.Error(w, "Name required", http.StatusBadRequest)
        return
    }
    
    // 3. Phải set header thủ công
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    // 4. Phải encode JSON thủ công
    json.NewEncoder(w).Encode(map[string]string{"hello": name})
})
```

```go
// Fiber - Ngắn gọn, rõ ràng
app.Get("/", func(c *fiber.Ctx) error {
    name := c.Query("name")
    if name == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Name required"})
    }
    return c.JSON(fiber.Map{"hello": name})
})
```

**Tại sao Fiber tốt hơn `net/http`?**

1. **Express-like API** - Nếu bạn biết Express, bạn đã biết 80% Fiber
2. **Zero allocation** - Fiber reuse memory objects giữa các requests → nhanh hơn
3. **Built-in helpers** - `c.JSON()`, `c.Query()`, `c.Params()` thay vì viết thủ công
4. **Middleware chain** - Dễ dàng thêm auth, logging, CORS

### 1.2 Fiber vs Express - Mental Model

```
┌─────────────────────────────────────────────────────────────────┐
│                         Express.js                              │
│                                                                 │
│   Request → Middleware 1 → Middleware 2 → Handler → Response    │
│                 ↓              ↓                                │
│              req, res       req, res                            │
│              next()         next()                              │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                          Fiber                                  │
│                                                                 │
│   Request → Middleware 1 → Middleware 2 → Handler → Response    │
│                 ↓              ↓                                │
│              Ctx            Ctx                                 │
│              c.Next()       c.Next()                            │
└─────────────────────────────────────────────────────────────────┘
```

**Khác biệt chính:**
- Express: `req` và `res` là **2 objects riêng biệt**
- Fiber: `Ctx` là **1 object chứa tất cả** (request + response + helpers)

**Tại sao Fiber gộp lại?**
- Trong Go, passing nhiều parameters tốn memory
- 1 object `Ctx` → ít allocations hơn → nhanh hơn
- `Ctx` được **reuse** giữa các requests (zero allocation)

---

## 🏗️ Phần 2: App - Hiểu sâu cách khởi tạo

### 2.1 `fiber.New()` làm gì bên trong?

```go
app := fiber.New()
```

Đoạn code này làm gì?

```go
// Bên trong fiber.New() (đơn giản hóa)
func New(config ...Config) *App {
    app := &App{
        // 1. Tạo router để match routes
        router: &Router{},
        
        // 2. Tạo server pool để reuse contexts
        pool: sync.Pool{
            New: func() interface{} {
                return &Ctx{app: app}
            },
        },
        
        // 3. Set default config
        config: defaultConfig(),
    }
    
    // 4. Apply custom config nếu có
    if len(config) > 0 {
        app.config = config[0]
    }
    
    return app
}
```

**Tại sao dùng `sync.Pool`?**

```go
// sync.Pool là "object pool" pattern
// Thay vì tạo mới Ctx cho mỗi request → reuse Ctx cũ

// Không dùng Pool (chậm):
func handleRequest() {
    ctx := &Ctx{}  // Tạo mới mỗi request → tốn memory
    // ...
    // ctx bị garbage collected
}

// Dùng Pool (nhanh):
func handleRequest() {
    ctx := pool.Get().(*Ctx)  // Lấy Ctx từ pool
    defer pool.Put(ctx)       // Trả lại pool khi xong
    // ...
}
```

**NodeJS Comparison:**
```javascript
// JavaScript không cần quan tâm vì có Garbage Collector
// Mỗi request tạo mới req, res objects
app.get('/', (req, res) => {
    // req, res được tạo mới mỗi request
    // GC sẽ dọn dẹp sau
});
```

**Suy nghĩ:** Go tối ưu performance bằng cách **tự quản lý memory**. NodeJS dựa vào **Garbage Collector** (chậm hơn nhưng đơn giản hơn).

### 2.2 Config - Tại sao cần cấu hình?

```go
app := fiber.New(fiber.Config{
    CaseSensitive: true,      // /Foo != /foo
    StrictRouting: true,      // /foo != /foo/
    BodyLimit:     4 * 1024 * 1024,  // 4MB
})
```

**Tại sao cần `CaseSensitive`?**

```go
// CaseSensitive: false (default)
app.Get("/User", handler)  // Match cả /User, /user, /USER

// CaseSensitive: true
app.Get("/User", handler)  // Chỉ match /User
```

**Khi nào dùng?**
- API public: `false` (user-friendly)
- API internal: `true` (strict, ít lỗi)

**Tại sao cần `StrictRouting`?**

```go
// StrictRouting: false (default)
app.Get("/user", handler)  // Match cả /user và /user/

// StrictRouting: true
app.Get("/user", handler)  // Chỉ match /user
app.Get("/user/", handler) // Match /user/
```

**Khi nào dùng?**
- SEO: `true` (tránh duplicate content)
- API: `false` (linh hoạt hơn)

**Tại sao cần `BodyLimit`?**

```go
// BodyLimit: 4MB (default)
// Nếu client gửi body > 4MB → trả về 413 Request Entity Too Large

// Bảo vệ khỏi:
// 1. Attack gửi body lớn để crash server
// 2. Upload file quá lớn
```

### 2.3 Immutable - Khái niệm quan trọng nhất

```go
app := fiber.New(fiber.Config{
    Immutable: false,  // Default
})
```

**Vấn đề:**
```go
app.Get("/:name", func(c *fiber.Ctx) error {
    name := c.Params("name")  // "john"
    
    // ⚠️ NGUY HIỂM: name sẽ bị thay đổi khi request tiếp theo đến!
    go func() {
        time.Sleep(1 * time.Second)
        fmt.Println(name)  // Có thể là "jane" (request tiếp theo)
    }()
    
    return c.SendString("Hello " + name)
})
```

**Tại sao?**
```
Request 1: /john → name = "john" → goroutine bắt đầu
Request 2: /jane → name = "jane" → goroutine vẫn đang chạy
                                    → name bị ghi đè thành "jane"
```

**Giải pháp 1: Copy thủ công**
```go
app.Get("/:name", func(c *fiber.Ctx) error {
    name := c.Params("name")  // "john"
    
    // Copy để an toàn
    nameCopy := make([]byte, len(name))
    copy(nameCopy, name)
    safeName := string(nameCopy)
    
    go func() {
        time.Sleep(1 * time.Second)
        fmt.Println(safeName)  // Luôn là "john"
    }()
    
    return c.SendString("Hello " + name)
})
```

**Giải pháp 2: Bật Immutable**
```go
app := fiber.New(fiber.Config{
    Immutable: true,  // Tự động copy mọi values
})

app.Get("/:name", func(c *fiber.Ctx) error {
    name := c.Params("name")  // Đã được copy, an toàn
    
    go func() {
        fmt.Println(name)  // Luôn đúng
    }()
    
    return c.SendString("Hello " + name)
})
```

**Trade-off:**
- `Immutable: false` → Nhanh hơn (reuse memory) nhưng phải cẩn thận
- `Immutable: true` → An toàn hơn nhưng chậm hơn (copy mỗi lần)

**NodeJS Comparison:**
```javascript
// JavaScript không có vấn đề này vì:
// 1. Strings là immutable (không thể thay đổi)
// 2. GC tự quản lý memory
app.get('/:name', (req, res) => {
    const name = req.params.name;  // Luôn an toàn
    setTimeout(() => console.log(name), 1000);  // OK
});
```

---

## 🛤️ Phần 3: Routing - Hiểu sâu cách match routes

### 3.1 Route matching hoạt động thế nào?

```go
app.Get("/user/:id", handler)
```

Khi request `GET /user/123` đến, Fiber làm gì?

```
1. Tách URL thành segments: ["user", "123"]
2. So sánh với route pattern: ["user", ":id"]
3. Match! → Extract param: id = "123"
4. Gọi handler với ctx có param "id"
```

**Bên trong Fiber:**
```go
// Route tree structure (đơn giản hóa)
type Route struct {
    Method  string
    Path    string
    Params  []string  // [":id"]
    Handler func(*Ctx) error
}

// Match function
func (r *Route) Match(method, path string) (bool, map[string]string) {
    if r.Method != method {
        return false, nil
    }
    
    segments := strings.Split(path, "/")
    if len(segments) != len(r.Params) {
        return false, nil
    }
    
    params := make(map[string]string)
    for i, param := range r.Params {
        if param[0] == ':' {
            params[param[1:]] = segments[i]  // :id → id = "123"
        } else if param != segments[i] {
            return false, nil
        }
    }
    
    return true, params
}
```

### 3.2 Named Parameters vs Wildcards

```go
// Named parameter (:id)
app.Get("/user/:id", handler)
// Match: /user/123, /user/abc
// Không match: /user/123/456

// Wildcard (*)
app.Get("/user/*", handler)
// Match: /user/123, /user/123/456, /user/abc/def
// * sẽ match TẤT CẢ sau /user/
```

**Khi nào dùng Wildcard?**
```go
// Serve static files
app.Get("/static/*", func(c *fiber.Ctx) error {
    path := c.Params("*")  // "css/style.css"
    return c.SendFile("./public/" + path)
})

// API versioning
app.Get("/api/v1/*", func(c *fiber.Ctx) error {
    path := c.Params("*")  // "users/123"
    // Handle v1 API
})
```

### 3.3 Optional Parameters

```go
// Optional parameter (:name?)
app.Get("/user/:name?", handler)
// Match: /user, /user/john
// Không match: /user/john/123
```

**Tại sao cần optional?**
```go
// Search API với optional filters
app.Get("/search/:category?", func(c *fiber.Ctx) error {
    category := c.Params("category")
    
    if category == "" {
        // Search all categories
        return searchAll()
    }
    
    // Search specific category
    return searchByCategory(category)
})
```

### 3.4 Route Priority

```go
// Route nào được match trước?
app.Get("/user/:id", handler1)   // Pattern 1
app.Get("/user/admin", handler2) // Pattern 2

// Request: GET /user/admin
// Match Pattern 1? id = "admin" → Có
// Match Pattern 2? Exact match → Có
// Ưu tiên Pattern 2 (exact match trước)
```

**Quy tắc ưu tiên:**
1. Exact match (`/user/admin`)
2. Named param (`/user/:id`)
3. Wildcard (`/user/*`)

---

## 📦 Phần 4: Context (Ctx) - Hiểu sâu request lifecycle

### 4.1 Ctx lifecycle

```go
app.Get("/", func(c *fiber.Ctx) error {
    // 1. Request đến → Fiber lấy Ctx từ pool
    // 2. Handler xử lý request
    // 3. Handler return → Ctx trả lại pool
    return c.SendString("Hello")
})
```

**Chi tiết:**
```
Request 1 đến:
  → pool.Get() → lấy Ctx
  → handler(ctx)
  → pool.Put(ctx) → trả lại pool

Request 2 đến:
  → pool.Get() → lấy Ctx CŨ (đã được reset)
  → handler(ctx)
  → pool.Put(ctx) → trả lại pool
```

**Tại sao reuse?**
- Tạo mới object mỗi lần → tốn memory → GC chạy nhiều → chậm
- Reuse object → ít memory → GC chạy ít → nhanh

### 4.2 BodyParser - Tại sao cần parse body?

```go
app.Post("/user", func(c *fiber.Ctx) error {
    // Raw body: {"name": "john", "age": 30}
    
    type User struct {
        Name string `json:"name"`
        Age  int    `json:"age"`
    }
    
    user := new(User)
    if err := c.BodyParser(user); err != nil {
        return err
    }
    
    // user.Name = "john"
    // user.Age = 30
    return c.JSON(user)
})
```

**BodyParser làm gì bên trong?**

```go
// Đơn giản hóa BodyParser
func (c *Ctx) BodyParser(out interface{}) error {
    // 1. Lấy raw body
    body := c.Body()  // []byte(`{"name": "john", "age": 30}`)
    
    // 2. Check Content-Type
    contentType := c.Get("Content-Type")
    
    switch {
    case strings.Contains(contentType, "application/json"):
        // 3a. Parse JSON
        return json.Unmarshal(body, out)
        
    case strings.Contains(contentType, "application/x-www-form-urlencoded"):
        // 3b. Parse form data
        return parseForm(body, out)
        
    case strings.Contains(contentType, "multipart/form-data"):
        // 3c. Parse multipart form
        return parseMultipart(c, out)
    }
    
    return errors.New("unsupported content type")
}
```

**Tại sao cần struct tags?**

```go
type User struct {
    Name string `json:"name"`      // JSON key là "name"
    Age  int    `json:"age"`       // JSON key là "age"
    Email string `json:"email"`    // JSON key là "email"
    Password string `json:"-"`     // Bỏ qua trong JSON
}

// JSON input: {"name": "john", "age": 30, "email": "john@example.com"}
// → User{Name: "john", Age: 30, Email: "john@example.com"}
```

**NodeJS Comparison:**
```javascript
// Express.js - req.body đã được parse sẵn
app.post('/user', (req, res) => {
    const { name, age } = req.body;  // Đã là object
    res.json({ name, age });
});

// Nhưng cần middleware:
app.use(express.json());  // Parse JSON body
app.use(express.urlencoded({ extended: true }));  // Parse form data
```

**Suy nghĩ:**
- Express: Middleware parse body → `req.body` là object
- Fiber: `c.BodyParser(&struct)` → parse và fill vào struct

### 4.3 Query vs Params vs Body - Khi nào dùng gì?

```go
// Params: Route parameters
app.Get("/user/:id", func(c *fiber.Ctx) error {
    id := c.Params("id")  // /user/123 → id = "123"
})

// Query: URL query string
app.Get("/search", func(c *fiber.Ctx) error {
    q := c.Query("q")     // /search?q=fiber → q = "fiber"
    page := c.Query("page", "1")  // /search?page=2 → page = "2"
})

// Body: Request body (POST, PUT, PATCH)
app.Post("/user", func(c *fiber.Ctx) error {
    user := new(User)
    c.BodyParser(user)    // {"name": "john"} → user.Name = "john"
})
```

**Quy tắc:**
- **Params**: Resource ID (`/user/:id`)
- **Query**: Filters, sorting, pagination (`?page=2&sort=name`)
- **Body**: Data để tạo/cập nhật (`{"name": "john"}`)

### 4.4 Locals - Pass data giữa middleware

```go
// Middleware set data
app.Use(func(c *fiber.Ctx) error {
    // Validate token
    token := c.Get("Authorization")
    user, err := validateToken(token)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
    }
    
    // Lưu user vào Locals
    c.Locals("user", user)
    c.Locals("userID", user.ID)
    
    return c.Next()
})

// Handler lấy data
app.Get("/profile", func(c *fiber.Ctx) error {
    user := c.Locals("user").(*User)  // Type assertion
    userID := c.Locals("userID").(uint)
    
    return c.JSON(fiber.Map{
        "user": user,
        "id":   userID,
    })
})
```

**Tại sao dùng Locals thay vì global variables?**

```go
// ❌ SAI: Global variable
var currentUser *User

app.Use(func(c *fiber.Ctx) error {
    currentUser = validateToken(c.Get("Authorization"))  // ⚠️ Race condition!
    return c.Next()
})

// ✅ ĐÚNG: Locals (scoped to request)
app.Use(func(c *fiber.Ctx) error {
    c.Locals("user", validateToken(c.Get("Authorization")))  // An toàn
    return c.Next()
})
```

**NodeJS Comparison:**
```javascript
// Express.js - Set trực tiếp trên req
app.use((req, res, next) => {
    req.user = validateToken(req.headers.authorization);
    next();
});

app.get('/profile', (req, res) => {
    res.json({ user: req.user });
});
```

**Suy nghĩ:**
- Express: `req.user = data` (set trực tiếp)
- Fiber: `c.Locals("user", data)` (dùng method)

Tại sao Fiber dùng method?
- Tránh namespace pollution (không biết `req.user` có bị ghi đè không)
- Type-safe hơn (dùng generics)

---

## 🔗 Phần 5: Middleware - Hiểu sâu chuỗi xử lý

### 5.1 Middleware chain hoạt động thế nào?

```go
app.Use(middleware1)
app.Use(middleware2)
app.Get("/", handler)
```

**Execution flow:**
```
Request → middleware1 → middleware2 → handler → Response
              ↓              ↓
           c.Next()       c.Next()
              ↓              ↓
         (chạy tiếp)    (chạy tiếp)
```

**Code chi tiết:**
```go
func middleware1(c *fiber.Ctx) error {
    fmt.Println("Before 1")
    
    err := c.Next()  // Chạy middleware2 + handler
    
    fmt.Println("After 1")
    return err
}

func middleware2(c *fiber.Ctx) error {
    fmt.Println("Before 2")
    
    err := c.Next()  // Chạy handler
    
    fmt.Println("After 2")
    return err
}

func handler(c *fiber.Ctx) error {
    fmt.Println("Handler")
    return c.SendString("Done")
}

// Output:
// Before 1
// Before 2
// Handler
// After 2
// After 1
```

**Tại sao cần `c.Next()`?**

```go
// Không có Next() → middleware dừng lại
func authMiddleware(c *fiber.Ctx) error {
    token := c.Get("Authorization")
    if token == "" {
        return c.Status(401).JSON(fiber.Map{"error": "No token"})
        // ⚠️ Không gọi Next() → handler không chạy
    }
    
    return c.Next()  // Tiếp tục đến handler
}
```

### 5.2 Pre-processing vs Post-processing

```go
func logger(c *fiber.Ctx) error {
    // Pre-processing: Trước khi handler chạy
    start := time.Now()
    
    // Chạy handler
    err := c.Next()
    
    // Post-processing: Sau khi handler chạy
    duration := time.Since(start)
    fmt.Printf("[%s] %s - %v\n", c.Method(), c.Path(), duration)
    
    return err
}
```

**Tại sao cần post-processing?**
- Logging response time
- Set response headers
- Cleanup resources
- Error handling

### 5.3 Middleware với Dependencies (Closure Pattern)

```go
// ❌ SAI: Middleware không có dependencies
func authMiddleware(c *fiber.Ctx) error {
    // Lấy secret từ đâu?
    // Global variable? ⚠️ Không an toàn
}

// ✅ ĐÚNG: Closure pattern
func authMiddleware(secret string) fiber.Handler {
    // secret được capture trong closure
    return func(c *fiber.Ctx) error {
        token := c.Get("Authorization")
        if !validateToken(token, secret) {
            return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
        }
        return c.Next()
    }
}

// Sử dụng
app.Use(authMiddleware("my-secret-key"))
```

**Tại sao dùng closure?**
- Dependencies được **inject** vào middleware
- Không cần global variables
- Testable (có thể mock dependencies)

### 5.4 Error Handling trong Middleware

```go
func errorHandler(c *fiber.Ctx) error {
    // Chạy handler và bắt lỗi
    err := c.Next()
    
    if err != nil {
        // Log lỗi
        log.Error(err)
        
        // Trả về error response
        return c.Status(500).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    return nil
}

// Sử dụng
app.Use(errorHandler)
```

**Tại sao cần error handler middleware?**
- Tập trung xử lý lỗi ở một chỗ
- Không phải try/catch trong mỗi handler
- Format error response nhất quán

---

## 👥 Phần 6: Grouping - Hiểu sâu cách nhóm routes

### 6.1 Tại sao cần grouping?

```go
// Không dùng grouping (lặp lại prefix)
app.Post("/api/v1/users", handler)
app.Get("/api/v1/users", handler)
app.Get("/api/v1/users/:id", handler)
app.Put("/api/v1/users/:id", handler)
app.Delete("/api/v1/users/:id", handler)

// Dùng grouping (clean hơn)
v1 := app.Group("/api/v1")
v1.Post("/users", handler)
v1.Get("/users", handler)
v1.Get("/users/:id", handler)
v1.Put("/users/:id", handler)
v1.Delete("/users/:id", handler)
```

### 6.2 Group Middleware

```go
// Auth middleware chỉ áp dụng cho group này
admin := app.Group("/admin", authMiddleware("admin-secret"))

admin.Get("/dashboard", handler)      // Cần auth
admin.Get("/settings", handler)       // Cần auth
admin.Post("/users", handler)         // Cần auth

app.Get("/public", handler)           // Không cần auth
```

**Tại sao middleware trong group?**
- Không phải set middleware cho từng route
- Dễ quản lý permissions
- Clean code

### 6.3 Nested Groups

```go
api := app.Group("/api")
v1 := api.Group("/v1")
v2 := api.Group("/v2")

// Routes
v1.Get("/users", handlerV1)  // GET /api/v1/users
v2.Get("/users", handlerV2)  // GET /api/v2/users
```

**Use case:**
- API versioning
- Feature flags
- Permission levels

---

## ❌ Phần 7: Error Handling - Hiểu sâu cách xử lý lỗi

### 7.1 Tại sao Go dùng error return?

```go
// Go: Error là giá trị, không phải exception
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

result, err := divide(10, 0)
if err != nil {
    // Handle error
}
```

**NodeJS Comparison:**
```javascript
// JavaScript: Error là exception
function divide(a, b) {
    if (b === 0) throw new Error("division by zero");
    return a / b;
}

try {
    const result = divide(10, 0);
} catch (err) {
    // Handle error
}
```

**Tại sao Go không dùng try/catch?**
- **Explicit**: Phải check error ngay sau khi gọi hàm
- **Performance**: Không có overhead của exception handling
- **Control flow**: Error là data, không phải control flow

### 7.2 Fiber ErrorHandler

```go
app := fiber.New(fiber.Config{
    ErrorHandler: func(c *fiber.Ctx, err error) error {
        // 1. Check error type
        var fiberErr *fiber.Error
        if errors.As(err, &fiberErr) {
            return c.Status(fiberErr.Code).JSON(fiber.Map{
                "error": fiberErr.Message,
            })
        }
        
        // 2. Default error
        return c.Status(500).JSON(fiber.Map{
            "error": "Internal Server Error",
        })
    },
})
```

**Tại sao cần custom ErrorHandler?**
- Format error response nhất quán
- Log errors cho debugging
- Hide internal errors trong production

### 7.3 Error Handling Patterns

```go
// Pattern 1: Return error directly
app.Get("/", func(c *fiber.Ctx) error {
    return errors.New("something went wrong")
    // ErrorHandler sẽ xử lý
})

// Pattern 2: Return Fiber error
app.Get("/", func(c *fiber.Ctx) error {
    return fiber.NewError(404, "User not found")
})

// Pattern 3: Custom error response
app.Get("/", func(c *fiber.Ctx) error {
    return c.Status(400).JSON(fiber.Map{
        "success": false,
        "error":   "Invalid input",
    })
})
```

### 7.4 Panic Recovery

```go
// Panic là "unrecoverable error" trong Go
app.Get("/", func(c *fiber.Ctx) error {
    panic("something went terribly wrong!")
    // ⚠️ Server sẽ crash nếu không recover
})

// Recover middleware bắt panics
app.Use(recover.New())

app.Get("/", func(c *fiber.Ctx) error {
    panic("something went terribly wrong!")
    // ✅ Recover sẽ bắt lỗi, server không crash
    return nil
})
```

**Tại sao cần recover?**
- Server không crash vì 1 request lỗi
- Log panic để debug
- Trả về 500 thay vì disconnect

---

## 📁 Phần 8: Static Files - Hiểu sâu cách phục vụ file

### 8.1 Tại sao cần static middleware?

```go
// Không dùng middleware (phải handle thủ công)
app.Get("/css/:file", func(c *fiber.Ctx) error {
    file := c.Params("file")
    return c.SendFile("./public/css/" + file)
})

// Dùng static middleware (tự động)
app.Use("/static", static.New("./public"))
```

### 8.2 Static Config

```go
app.Use("/static", static.New("./public", static.Config{
    Browse:     true,    // Cho phép browse directories
    Compress:   true,    // Enable gzip compression
    CacheDuration: 10 * time.Second,
    MaxAge:     3600,    // Cache-Control max-age
}))
```

**Tại sao cần Compress?**
- Giảm bandwidth
- Tăng tốc load trang
- Tiết kiệm chi phí hosting

---

## 🛡️ Phần 9: Built-in Middleware - Hiểu sâu tại sao cần

### 9.1 Logger - Tại sao cần log requests?

```go
app.Use(logger.New(logger.Config{
    Format: "${pid} | ${ip} | ${method} | ${path} | ${status} | ${latency}\n",
}))
```

**Tại sao log?**
- Debug: Request nào lỗi?
- Monitoring: API nào chậm?
- Security: Ai đang attack?

### 9.2 CORS - Tại sao cần?

```go
app.Use(cors.New(cors.Config{
    AllowOrigins: "*",
    AllowMethods: "GET,POST,PUT,DELETE",
}))
```

**Vấn đề:**
```
Frontend: http://localhost:3000
Backend: http://localhost:8080

Frontend gọi Backend → Browser chặn!
"Cross-Origin Request Blocked: Same Origin Policy"
```

**Giải pháp: CORS middleware**
- Thêm headers cho phép cross-origin requests
- Browser thấy headers → cho phép request

### 9.3 Rate Limiter - Tại sao cần?

```go
app.Use(limiter.New(limiter.Config{
    Max:        100,           // 100 requests
    Expiration: 1 * time.Minute,  // Trong 1 phút
}))
```

**Tại sao cần?**
- Chống DDoS attacks
- Chống spam
- Bảo vệ server resources

### 9.4 Recover - Tại sao cần?

```go
app.Use(recover.New())
```

**Tại sao cần?**
- Server không crash vì 1 request panic
- Log panics để debug
- Trả về 500 thay vì disconnect

---

## 🎯 Phần 10: Best Practices - Hiểu sâu tại sao

### 10.1 Closure Pattern cho Dependency Injection

```go
// ❌ SAI: Global variables
var db *gorm.DB
var redis *redis.Client

func init() {
    db = connectDB()
    redis = connectRedis()
}

func handler(c *fiber.Ctx) error {
    // Dùng global db, redis
}

// ✅ ĐÚNG: Dependency injection
func handler(db *gorm.DB, redis *redis.Client) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Dùng injected db, redis
    }
}
```

**Tại sao dependency injection tốt hơn?**
- **Testable**: Có thể mock dependencies
- **Explicit**: Biết rõ handler cần gì
- **No globals**: Tránh side effects

### 10.2 Context cho Database Queries

```go
// ❌ SAI: Không dùng context
func handler(c *fiber.Ctx) error {
    user, err := db.FindById(id)  // Query có thể chạy mãi
    // ...
}

// ✅ ĐÚNG: Pass context
func handler(c *fiber.Ctx) error {
    ctx := c.Context()
    user, err := db.WithContext(ctx).FindById(id)
    // Query sẽ cancel khi client disconnect
    // ...
}
```

**Tại sao cần context?**
- Client disconnect → query cancel → tiết kiệm resources
- Timeout protection
- Tracing (APM)

### 10.3 Route Grouping

```go
// ✅ GOOD: Group theo feature
api := app.Group("/api")
auth := api.Group("/auth")
users := api.Group("/users")

auth.Post("/login", loginHandler)
auth.Post("/register", registerHandler)
users.Get("/", listUsers)
users.Get("/:id", getUser)

// ❌ BAD: Flat routes
app.Post("/api/auth/login", loginHandler)
app.Post("/api/auth/register", registerHandler)
app.Get("/api/users", listUsers)
app.Get("/api/users/:id", getUser)
```

**Tại sao grouping tốt hơn?**
- Dễ đọc
- Dễ maintain
- Dễ thêm middleware cho group

---

## 📊 Tổng kết: Mental Model

### Fiber vs Express

| Concept | Express | Fiber |
|---------|---------|-------|
| **App** | `express()` | `fiber.New()` |
| **Request** | `req` (object) | `c` (Ctx) |
| **Response** | `res` (object) | `c` (Ctx) |
| **Next** | `next()` | `c.Next()` |
| **Body** | `req.body` | `c.BodyParser()` |
| **Params** | `req.params` | `c.Params()` |
| **Query** | `req.query` | `c.Query()` |
| **Headers** | `req.headers` | `c.Get()` |
| **Locals** | `req.user = data` | `c.Locals("user", data)` |
| **Router** | `express.Router()` | `app.Group()` |
| **Middleware** | `app.use()` | `app.Use()` |

### Cách suy nghĩ khi dùng Fiber

1. **Handlers trả về error** - Không phải `res.send()`
2. **Ctx chứa tất cả** - Request + Response + Helpers
3. **Closures cho dependencies** - Không dùng globals
4. **Locals cho request data** - Không set trực tiếp trên Ctx
5. **Groups cho route organization** - Không lặp lại prefix
6. **Middleware cho cross-cutting concerns** - Auth, logging, CORS

### Tiếp theo

- [[01-Project-Structure-Analysis]] - Xem project dùng Fiber thế nào
- [[04-Practical-Patterns-from-Project]] - Patterns thực tế với Fiber
- [Fiber Official Docs](https://docs.gofiber.io/) - Documentation đầy đủ
