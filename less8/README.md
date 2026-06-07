# Lesson 8: HTTP Server với Fiber

## 🎯 Mục tiêu
Xây dựng **production-ready HTTP API** với Fiber - framework nhanh nhất cho Go, tương tự Express.js nhưng performance cao hơn.

## 🧠 Mental Model - Góc nhìn Expert

### Fiber = Express.js cho Go
| Express.js | Fiber |
|---|---|
| `app.use(middleware)` | `app.Use(middleware)` |
| `app.get('/', handler)` | `app.Get('/', handler)` |
| `req.params.id` | `c.Params("id")` |
| `req.body` | `c.BodyParser(&struct)` |
| `res.json({...})` | `c.JSON(fiber.Map{...})` |

### Go HTTP khác NodeJS
- NodeJS: Single-threaded event loop, non-blocking I/O
- Go: Multi-threaded, mỗi request chạy trong goroutine riêng
- **Không cần async/await** - mỗi handler tự động chạy concurrently

### Middleware Chain
```go
app.Use(recover.New())    // Catch panics
app.Use(logger.New())     // Log requests
app.Use(cors.New(...))    // CORS headers
// ... your handlers
```
- Middleware chạy theo thứ tự đăng ký
- Mỗi middleware có thể modify request/response hoặc abort

## 📚 Kiến thức cốt lõi

### App Configuration
```go
app := fiber.New(fiber.Config{
    AppName:      "My API",
    BodyLimit:    4 * 1024 * 1024,  // 4MB
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
    ErrorHandler: customErrorHandler,
})
```

### Route Handlers
```go
// Basic handler
app.Get("/", func(c *fiber.Ctx) error {
    return c.SendString("Hello, World!")
})

// Handler với params
app.Get("/users/:id", func(c *fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid ID",
        })
    }
    return c.JSON(user)
})

// Handler với body parsing
app.Post("/users", func(c *fiber.Ctx) error {
    user := new(User)
    if err := c.BodyParser(user); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid body",
        })
    }
    // ... save user
    return c.Status(201).JSON(user)
})
```

### Middleware
```go
// Global middleware
app.Use(func(c *fiber.Ctx) error {
    fmt.Println("Request:", c.Method(), c.Path())
    return c.Next()  // Continue to next handler
})

// Route-specific middleware
app.Get("/admin", authMiddleware, adminHandler)
```

### Route Grouping
```go
api := app.Group("/api/v1")

api.Get("/users", getUsers)
api.Post("/users", createUser)
api.Get("/users/:id", getUser)
```

### Error Handling
```go
// Custom error handler
app := fiber.New(fiber.Config{
    ErrorHandler: func(c *fiber.Ctx, err error) error {
        code := fiber.StatusInternalServerError
        if e, ok := err.(*fiber.Error); ok {
            code = e.Code
        }
        return c.Status(code).JSON(fiber.Map{
            "error": err.Error(),
        })
    },
})
```

## ⚠️ Sai lầm NodeJS dev hay mắc

1. **Tưởng handler = async function**
```go
// ❌ Không cần async/await trong handler signature
app.Get("/", async func(c *fiber.Ctx) error {  // ❌ Không có async!

// ✅ Đúng
app.Get("/", func(c *fiber.Ctx) error {
    result, err := fetchData()  // Gọi sync function
    if err != nil {
        return err
    }
    return c.JSON(result)
})
```

2. **Quên return sau error**
```go
// ❌ Bug: code tiếp tục chạy sau error
if err != nil {
    c.Status(500).JSON(fiber.Map{"error": err.Error()})
}
// Code ở đây vẫn chạy!

// ✅ Đúng
if err != nil {
    return c.Status(500).JSON(fiber.Map{"error": err.Error()})
}
```

3. **Không parse body**
```go
// ❌ Body không tự động parse
app.Post("/users", func(c *fiber.Ctx) error {
    name := c.FormValue("name")  // Chỉ lấy form data, không phải JSON!
    
// ✅ Đúng
app.Post("/users", func(c *fiber.Ctx) error {
    user := new(User)
    if err := c.BodyParser(user); err != nil {
        return err
    }
})
```

## 🏆 Best Practices (Expert Level)

### 1. Validation với struct tags
```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=0,max=150"`
}

func createUser(c *fiber.Ctx) error {
    req := new(CreateUserRequest)
    if err := c.BodyParser(req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }
    
    if err := validate.Struct(req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }
    // ...
}
```

### 2. Consistent response format
```go
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func success(c *fiber.Ctx, data interface{}) error {
    return c.JSON(Response{Success: true, Data: data})
}

func fail(c *fiber.Ctx, status int, err string) error {
    return c.Status(status).JSON(Response{Success: false, Error: err})
}
```

### 3. Graceful shutdown
```go
func main() {
    app := fiber.New()
    
    // ... routes
    
    go func() {
        if err := app.Listen(":3000"); err != nil {
            log.Printf("Server error: %v", err)
        }
    }()
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    
    <-quit
    
    if err := app.Shutdown(); err != nil {
        log.Fatal(err)
    }
}
```

### 4. Dependency injection với closures
```go
func NewUserHandler(db *sql.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // db available via closure
        users, err := getUsers(db)
        if err != nil {
            return err
        }
        return c.JSON(users)
    }
}

app.Get("/users", NewUserHandler(db))
```

### 5. Request timeout
```go
app.Use(func(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
    defer cancel()
    
    c.SetUserContext(ctx)
    return c.Next()
})
```

## 🔗 Liên kết
- **Previous:** [Lesson 7: Real-world Concurrency Patterns](../less7/README.md)
- **Next:** [Lesson 9: JSON Handling](../less9/README.md)
- **Docs:** [go-fundamentals/05-http-api-project-layout.md](../../docs/go-fundamentals/05-http-api-project-layout.md)

---
*"Fiber cho phép bạn xây dựng API production-ready trong vài phút. Nhưng đừng quên: performance tốt không thay thế cho architecture tốt. Luôn nghĩ đến graceful shutdown, validation, và consistent error handling."*
