# 🎖️ Senior NodeJS Dev → Senior Go Dev: Production-Ready Patterns

> **Mục tiêu:** Giúp bạn code Go như một Senior Dev, không phải Junior mới học
> **Dành cho:** NodeJS dev 4+ năm kinh nghiệm muốn code Go production-ready ngay

---

## 🧠 Phần 1: Senior Mindset - Cách suy nghĩ khác nhau

### 1.1 Junior vs Senior trong NodeJS

```javascript
// Junior: Code chạy được là được
app.post('/login', async (req, res) => {
    const user = await db.findUser(req.body.username);
    const token = jwt.sign({ id: user.id }, SECRET);
    res.json({ token });
});

// Senior: Code chạy được, AN TOÀN, MAINTAINABLE, OBSERVABLE
app.post('/login', validateBody(loginSchema), rateLimiter, async (req, res, next) => {
    try {
        const { username, password } = req.body;
        
        // Structured logging
        logger.info({ username, ip: req.ip }, 'Login attempt');
        
        // Error handling cụ thể
        const user = await userService.findByUsername(username);
        if (!user || !await bcrypt.compare(password, user.password)) {
            logger.warn({ username, ip: req.ip }, 'Login failed');
            throw new AppError('Invalid credentials', 401);
        }
        
        // Rate limiting check
        await rateLimiter.check(user.id);
        
        // Generate token với proper claims
        const token = tokenService.generate({
            sub: user.id,
            roles: user.roles,
            jti: uuid(),
        });
        
        // Audit log
        await auditLog.record({ userId: user.id, action: 'login', ip: req.ip });
        
        logger.info({ userId: user.id }, 'Login successful');
        res.json({ token, user: sanitizeUser(user) });
    } catch (err) {
        next(err);  // Centralized error handling
    }
});
```

### 1.2 Junior vs Senior trong Go

```go
// Junior: Code chạy được là được
app.Post("/login", func(c *fiber.Ctx) error {
    body := new(LoginBody)
    c.BodyParser(body)
    
    user, _ := db.FindByUsername(body.Username)
    token, _ := generateToken(user)
    
    return c.JSON(fiber.Map{"token": token})
})

// Senior: Code chạy được, AN TOÀN, MAINTAINABLE, OBSERVABLE
app.Post("/login", 
    middleware.ValidateBody[LoginBody](),
    middleware.RateLimiter(10, time.Minute),
    func(c *fiber.Ctx) error {
        // Structured logging với context
        logger := log.WithFields(log.Fields{
            "ip":       c.IP(),
            "endpoint": "/login",
        })
        
        // Parse và validate
        body := middleware.GetBody[LoginBody](c)
        if err := validator.Struct(body); err != nil {
            logger.WithError(err).Warn("Validation failed")
            return fiber.NewError(400, "Invalid request body")
        }
        
        logger.Info("Login attempt")
        
        // Business logic với proper error handling
        user, err := userService.FindByUsername(c.Context(), body.Username)
        if err != nil {
            if errors.Is(err, ErrUserNotFound) {
                logger.Warn("User not found")
                return fiber.NewError(401, "Invalid credentials")
            }
            logger.WithError(err).Error("Database error")
            return fiber.NewError(500, "Internal server error")
        }
        
        // Password verification
        if err := userService.VerifyPassword(user, body.Password); err != nil {
            logger.WithField("userId", user.ID).Warn("Invalid password")
            return fiber.NewError(401, "Invalid credentials")
        }
        
        // Generate token
        token, err := tokenService.Generate(c.Context(), &TokenClaims{
            UserID: user.ID,
            Roles:  user.Roles,
            JTI:    uuid.New().String(),
        })
        if err != nil {
            logger.WithError(err).Error("Token generation failed")
            return fiber.NewError(500, "Failed to generate token")
        }
        
        // Audit log (async, không block response)
        go auditService.Record(c.Context(), &AuditEvent{
            UserID: user.ID,
            Action: "login",
            IP:     c.IP(),
        })
        
        logger.WithField("userId", user.ID).Info("Login successful")
        
        return c.JSON(LoginResponse{
            Token: token,
            User:  sanitizeUser(user),
        })
    },
)
```

**Khác biệt chính:**
1. **Error handling** - Junior bỏ qua, Senior handle từng loại lỗi
2. **Logging** - Junior không log, Senior log structured
3. **Validation** - Junior tin tưởng input, Senior validate mọi thứ
4. **Security** - Junior không nghĩ đến, Senior nghĩ đến đầu tiên
5. **Observability** - Junior không monitor, Senior instrument mọi thứ

---

## 🏗️ Phần 2: Production-Ready Architecture

### 2.1 Dependency Injection Pattern

```go
// ❌ Junior: Global variables
var db *gorm.DB
var redis *redis.Client
var logger *logrus.Logger

func init() {
    db = connectDB()
    redis = connectRedis()
    logger = logrus.New()
}

func handler(c *fiber.Ctx) error {
    // Dùng global variables - khó test, khó maintain
}

// ✅ Senior: Dependency injection
type App struct {
    db     *gorm.DB
    redis  *redis.Client
    logger *logrus.Logger
    
    // Services
    userService    *UserService
    tokenService   *TokenService
    auditService   *AuditService
}

func NewApp(cfg *Config) (*App, error) {
    // Initialize dependencies
    db, err := database.Connect(cfg.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("connect database: %w", err)
    }
    
    redis, err := cache.Connect(cfg.RedisURL)
    if err != nil {
        return nil, fmt.Errorf("connect redis: %w", err)
    }
    
    logger := logrus.New()
    logger.SetFormatter(&logrus.JSONFormatter{})
    
    // Initialize services
    userService := NewUserService(db, logger)
    tokenService := NewTokenService(redis, cfg.JWTSecret, logger)
    auditService := NewAuditService(db, logger)
    
    return &App{
        db:           db,
        redis:        redis,
        logger:       logger,
        userService:  userService,
        tokenService: tokenService,
        auditService: auditService,
    }, nil
}

// Handler với injected dependencies
func (app *App) LoginHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Dùng app.userService, app.tokenService, etc.
        // Dễ test: mock dependencies
        // Dễ maintain: biết rõ dependencies
    }
}
```

### 2.2 Service Layer Pattern

```go
// ❌ Junior: Business logic trong controller
func (app *App) LoginHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        body := new(LoginBody)
        c.BodyParser(body)
        
        // Business logic nằm trong controller
        user := &User{}
        app.db.Where("username = ?", body.Username).First(user)
        
        if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
            return fiber.NewError(401, "Invalid credentials")
        }
        
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
            "sub": user.ID,
            "exp": time.Now().Add(time.Hour).Unix(),
        })
        tokenString, _ := token.SignedString([]byte("secret"))
        
        return c.JSON(fiber.Map{"token": tokenString})
    }
}

// ✅ Senior: Business logic trong service
type UserService struct {
    db     *gorm.DB
    logger *logrus.Logger
}

func NewUserService(db *gorm.DB, logger *logrus.Logger) *UserService {
    return &UserService{db: db, logger: logger}
}

func (s *UserService) Authenticate(ctx context.Context, username, password string) (*User, error) {
    // Tìm user
    user, err := s.FindByUsername(ctx, username)
    if err != nil {
        return nil, fmt.Errorf("find user: %w", err)
    }
    
    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
        s.logger.WithFields(logrus.Fields{
            "userId":   user.ID,
            "username": username,
        }).Warn("Invalid password attempt")
        return nil, ErrInvalidCredentials
    }
    
    // Update last login
    if err := s.UpdateLastLogin(ctx, user.ID); err != nil {
        s.logger.WithError(err).Error("Failed to update last login")
        // Không return error vì login vẫn thành công
    }
    
    return user, nil
}

// Controller chỉ gọi service
func (app *App) LoginHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        body := new(LoginBody)
        if err := c.BodyParser(body); err != nil {
            return fiber.NewError(400, "Invalid request body")
        }
        
        user, err := app.userService.Authenticate(c.Context(), body.Username, body.Password)
        if err != nil {
            if errors.Is(err, ErrInvalidCredentials) {
                return fiber.NewError(401, "Invalid credentials")
            }
            return fiber.NewError(500, "Internal server error")
        }
        
        token, err := app.tokenService.Generate(c.Context(), user)
        if err != nil {
            return fiber.NewError(500, "Failed to generate token")
        }
        
        return c.JSON(LoginResponse{Token: token, User: sanitizeUser(user)})
    }
}
```

### 2.3 Repository Pattern

```go
// ❌ Junior: Query trực tiếp trong service
func (s *UserService) FindByUsername(ctx context.Context, username string) (*User, error) {
    user := &User{}
    result := s.db.WithContext(ctx).Where("username = ?", username).First(user)
    if result.Error != nil {
        return nil, result.Error
    }
    return user, nil
}

// ✅ Senior: Repository pattern
type UserRepository interface {
    FindByUsername(ctx context.Context, username string) (*User, error)
    FindByID(ctx context.Context, id uint) (*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
}

type userRepository struct {
    db     *gorm.DB
    logger *logrus.Logger
}

func NewUserRepository(db *gorm.DB, logger *logrus.Logger) UserRepository {
    return &userRepository{db: db, logger: logger}
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
    user := &User{}
    
    // Context timeout
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    // Prepared statement cho performance
    session := r.db.Session(&gorm.Session{PrepareStmt: true})
    
    result := session.WithContext(ctx).
        Where("username = ?", username).
        First(user)
    
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        r.logger.WithError(result.Error).Error("Failed to find user by username")
        return nil, fmt.Errorf("find user: %w", result.Error)
    }
    
    return user, nil
}

// Service dùng repository (dễ mock để test)
type UserService struct {
    repo   UserRepository
    logger *logrus.Logger
}

func (s *UserService) FindByUsername(ctx context.Context, username string) (*User, error) {
    return s.repo.FindByUsername(ctx, username)
}
```

### 2.4 Error Handling Patterns

```go
// ❌ Junior: Error handling không nhất quán
func handler(c *fiber.Ctx) error {
    user, err := db.FindById(id)
    if err != nil {
        return err  // Stack trace lộ ra ngoài
    }
}

// ✅ Senior: Centralized error handling
// 1. Define error types
type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Err     error  `json:"-"`
}

func (e *AppError) Error() string {
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Pre-defined errors
var (
    ErrUserNotFound      = &AppError{Code: 404, Message: "User not found"}
    ErrInvalidCredentials = &AppError{Code: 401, Message: "Invalid credentials"}
    ErrRateLimited       = &AppError{Code: 429, Message: "Too many requests"}
)

// 2. Centralized error handler
func ErrorHandler(c *fiber.Ctx, err error) error {
    // Log error
    logger.WithError(err).Error("Request error")
    
    // Check error type
    var appErr *AppError
    if errors.As(err, &appErr) {
        return c.Status(appErr.Code).JSON(ErrorResponse{
            Code:    appErr.Code,
            Message: appErr.Message,
        })
    }
    
    var fiberErr *fiber.Error
    if errors.As(err, &fiberErr) {
        return c.Status(fiberErr.Code).JSON(ErrorResponse{
            Code:    fiberErr.Code,
            Message: fiberErr.Message,
        })
    }
    
    // Unknown error - hide details
    return c.Status(500).JSON(ErrorResponse{
        Code:    500,
        Message: "Internal server error",
    })
}

// 3. Error wrapping để preserve context
func (s *UserService) FindByUsername(ctx context.Context, username string) (*User, error) {
    user, err := s.repo.FindByUsername(ctx, username)
    if err != nil {
        return nil, fmt.Errorf("UserService.FindByUsername: %w", err)
    }
    return user, nil
}
```

### 2.5 Middleware Patterns

```go
// ❌ Junior: Middleware không có dependencies
func AuthMiddleware(c *fiber.Ctx) error {
    token := c.Get("Authorization")
    // Validate token... nhưng lấy secret từ đâu?
}

// ✅ Senior: Middleware với dependencies (closure pattern)
func AuthMiddleware(tokenService *TokenService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Extract token
        token := c.Get("Authorization")
        if token == "" {
            return fiber.NewError(401, "Missing authorization header")
        }
        
        // Parse "Bearer <token>"
        parts := strings.SplitN(token, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            return fiber.NewError(401, "Invalid authorization header format")
        }
        
        // Validate token
        claims, err := tokenService.Validate(c.Context(), parts[1])
        if err != nil {
            if errors.Is(err, ErrTokenExpired) {
                return fiber.NewError(401, "Token expired")
            }
            return fiber.NewError(401, "Invalid token")
        }
        
        // Set claims vào context
        c.Locals("claims", claims)
        c.Locals("userID", claims.UserID)
        c.Locals("roles", claims.Roles)
        
        return c.Next()
    }
}

// Generic validation middleware
func ValidateBody[T any]() fiber.Handler {
    return func(c *fiber.Ctx) error {
        body := new(T)
        if err := c.BodyParser(body); err != nil {
            return fiber.NewError(400, "Invalid request body")
        }
        
        if err := validator.Struct(body); err != nil {
            return fiber.NewError(400, fmt.Sprintf("Validation failed: %v", err))
        }
        
        c.Locals("body", body)
        return c.Next()
    }
}

// Helper để lấy body đã validate
func GetBody[T any](c *fiber.Ctx) *T {
    return c.Locals("body").(*T)
}
```

---

## ⚡ Phần 3: Performance Optimization

### 3.1 Connection Pooling

```go
// ❌ Junior: Không config connection pool
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

// ✅ Senior: Config connection pool
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{
    SkipDefaultTransaction: true,  // Tăng performance
    PrepareStmt:            true,  // Cache prepared statements
})

sqlDB, _ := db.DB()

// Connection pool config
sqlDB.SetMaxIdleConns(10)           // Idle connections
sqlDB.SetMaxOpenConns(100)          // Max open connections
sqlDB.SetConnMaxLifetime(time.Hour) // Max lifetime
sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Max idle time
```

**Tại sao cần config?**
- `MaxIdleConns`: Giữ connections sẵn sàng → giảm latency
- `MaxOpenConns`: Giới hạn connections → tránh overload database
- `ConnMaxLifetime`: Recycle connections → tránh stale connections

### 3.2 Redis Pipeline

```go
// ❌ Junior: Nhiều requests riêng lẻ
redis.Set(ctx, "key1", "value1", time.Hour)
redis.Set(ctx, "key2", "value2", time.Hour)
redis.Set(ctx, "key3", "value3", time.Hour)
// 3 round trips → chậm

// ✅ Senior: Pipeline
pipe := redis.Pipeline()
pipe.Set(ctx, "key1", "value1", time.Hour)
pipe.Set(ctx, "key2", "value2", time.Hour)
pipe.Set(ctx, "key3", "value3", time.Hour)
_, err := pipe.Exec(ctx)
// 1 round trip → nhanh
```

### 3.3 Goroutines cho Background Tasks

```go
// ❌ Junior: Block response
func (app *App) LoginHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        user, err := app.userService.Authenticate(...)
        if err != nil {
            return err
        }
        
        // Audit log (block response - chậm)
        app.auditService.Record(c.Context(), &AuditEvent{
            UserID: user.ID,
            Action: "login",
        })
        
        return c.JSON(...)
    }
}

// ✅ Senior: Async với goroutines
func (app *App) LoginHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        user, err := app.userService.Authenticate(...)
        if err != nil {
            return err
        }
        
        // Audit log (async - không block response)
        go func() {
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            
            if err := app.auditService.Record(ctx, &AuditEvent{
                UserID: user.ID,
                Action: "login",
            }); err != nil {
                app.logger.WithError(err).Error("Failed to record audit log")
            }
        }()
        
        return c.JSON(...)
    }
}
```

### 3.4 Caching Strategy

```go
// ✅ Senior: Cache-aside pattern
func (s *UserService) FindByID(ctx context.Context, id uint) (*User, error) {
    // 1. Try cache
    cacheKey := fmt.Sprintf("user:%d", id)
    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        user := &User{}
        json.Unmarshal([]byte(cached), user)
        return user, nil  // Cache hit
    }
    
    // 2. Cache miss → query database
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 3. Update cache (async)
    go func() {
        data, _ := json.Marshal(user)
        s.redis.Set(ctx, cacheKey, data, 10*time.Minute)
    }()
    
    return user, nil
}

// Cache invalidation
func (s *UserService) Update(ctx context.Context, user *User) error {
    // 1. Update database
    if err := s.repo.Update(ctx, user); err != nil {
        return err
    }
    
    // 2. Invalidate cache
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    s.redis.Del(ctx, cacheKey)
    
    return nil
}
```

---

## 🔒 Phần 4: Security Best Practices

### 4.1 Input Validation

```go
// ❌ Junior: Không validate input
func (app *App) CreateUserHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        body := new(CreateUserBody)
        c.BodyParser(body)
        
        user, err := app.userService.Create(c.Context(), body)
        // ...
    }
}

// ✅ Senior: Validate mọi input
type CreateUserBody struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Phone    string `json:"phone" validate:"omitempty,e164"`
}

func (app *App) CreateUserHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        body := new(CreateUserBody)
        if err := c.BodyParser(body); err != nil {
            return fiber.NewError(400, "Invalid request body")
        }
        
        if err := validator.Struct(body); err != nil {
            // Return specific validation errors
            errors := err.(validator.ValidationErrors)
            return fiber.NewError(400, formatValidationErrors(errors))
        }
        
        // Sanitize input
        body.Username = sanitizeUsername(body.Username)
        body.Email = strings.ToLower(strings.TrimSpace(body.Email))
        
        user, err := app.userService.Create(c.Context(), body)
        // ...
    }
}
```

### 4.2 SQL Injection Prevention

```go
// ❌ Junior: String concatenation (SQL injection!)
func (s *UserService) FindByUsername(ctx context.Context, username string) (*User, error) {
    query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", username)
    // ⚠️ SQL injection nếu username = "'; DROP TABLE users; --"
}

// ✅ Senior: Parameterized queries
func (s *UserService) FindByUsername(ctx context.Context, username string) (*User, error) {
    user := &User{}
    result := s.db.WithContext(ctx).
        Where("username = ?", username).  // Parameterized
        First(user)
    // ✅ An toàn
}
```

### 4.3 Rate Limiting

```go
// ✅ Senior: Rate limiting per endpoint
func RateLimiter(max int, window time.Duration) fiber.Handler {
    return func(c *fiber.Ctx) error {
        key := fmt.Sprintf("rate_limit:%s:%s", c.Path(), c.IP())
        
        current, err := redis.Incr(c.Context(), key).Result()
        if err != nil {
            return fiber.NewError(500, "Rate limiter error")
        }
        
        if current == 1 {
            redis.Expire(c.Context(), key, window)
        }
        
        if current > int64(max) {
            c.Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
            return fiber.NewError(429, "Too many requests")
        }
        
        c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", max))
        c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max-int(current)))
        
        return c.Next()
    }
}
```

### 4.4 Security Headers

```go
// ✅ Senior: Security headers middleware
func SecurityHeaders() fiber.Handler {
    return func(c *fiber.Ctx) error {
        c.Set("X-Content-Type-Options", "nosniff")
        c.Set("X-Frame-Options", "DENY")
        c.Set("X-XSS-Protection", "1; mode=block")
        c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Set("Content-Security-Policy", "default-src 'self'")
        c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
        
        return c.Next()
    }
}
```

---

## 📊 Phần 5: Observability

### 5.1 Structured Logging

```go
// ❌ Junior: Print debugging
fmt.Println("User logged in:", user.ID)
fmt.Printf("Error: %v\n", err)

// ✅ Senior: Structured logging
type Logger struct {
    *logrus.Logger
}

func (l *Logger) WithRequest(c *fiber.Ctx) *logrus.Entry {
    return l.WithFields(logrus.Fields{
        "requestId": c.Locals("requestId"),
        "method":    c.Method(),
        "path":      c.Path(),
        "ip":        c.IP(),
        "userAgent": c.Get("User-Agent"),
    })
}

func (l *Logger) WithUser(userID uint) *logrus.Entry {
    return l.WithField("userId", userID)
}

// Usage
func (app *App) LoginHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        logger := app.logger.WithRequest(c)
        
        logger.Info("Login attempt")
        
        user, err := app.userService.Authenticate(...)
        if err != nil {
            logger.WithError(err).Warn("Login failed")
            return err
        }
        
        logger.WithUser(user.ID).Info("Login successful")
        return c.JSON(...)
    }
}
```

### 5.2 Metrics Collection

```go
// ✅ Senior: Metrics middleware
func MetricsMiddleware(metrics *Metrics) fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()
        
        // Chạy handler
        err := c.Next()
        
        // Collect metrics
        duration := time.Since(start)
        statusCode := c.Response().StatusCode()
        
        metrics.RequestDuration.WithLabelValues(
            c.Method(),
            c.Path(),
            fmt.Sprintf("%d", statusCode),
        ).Observe(duration.Seconds())
        
        metrics.RequestTotal.WithLabelValues(
            c.Method(),
            c.Path(),
            fmt.Sprintf("%d", statusCode),
        ).Inc()
        
        if statusCode >= 500 {
            metrics.ErrorTotal.WithLabelValues(
                c.Method(),
                c.Path(),
            ).Inc()
        }
        
        return err
    }
}
```

### 5.3 Distributed Tracing

```go
// ✅ Senior: APM integration
func (app *App) LoginHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Start transaction
        tx := apm.DefaultTracer().StartTransaction("POST /login", "request")
        defer tx.End()
        
        // Set context
        ctx := apm.ContextWithTransaction(c.Context(), tx)
        
        // Add labels
        tx.Context.SetLabel("userId", c.Locals("userID"))
        
        // Call service với traced context
        user, err := app.userService.Authenticate(ctx, ...)
        if err != nil {
            // Capture error
            apm.CaptureError(ctx, err).Send()
            return err
        }
        
        return c.JSON(...)
    }
}
```

---

## 🧪 Phần 6: Testing Strategies

### 6.1 Unit Testing

```go
// ✅ Senior: Unit test với mocks
type MockUserRepository struct {
    FindByUsernameFn func(ctx context.Context, username string) (*User, error)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
    return m.FindByUsernameFn(ctx, username)
}

func TestUserService_Authenticate(t *testing.T) {
    tests := []struct {
        name     string
        username string
        password string
        mockFn   func(ctx context.Context, username string) (*User, error)
        wantErr  error
    }{
        {
            name:     "success",
            username: "john",
            password: "password123",
            mockFn: func(ctx context.Context, username string) (*User, error) {
                hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
                return &User{ID: 1, Username: "john", Password: string(hashed)}, nil
            },
            wantErr: nil,
        },
        {
            name:     "user not found",
            username: "unknown",
            password: "password123",
            mockFn: func(ctx context.Context, username string) (*User, error) {
                return nil, ErrUserNotFound
            },
            wantErr: ErrInvalidCredentials,
        },
        {
            name:     "wrong password",
            username: "john",
            password: "wrong",
            mockFn: func(ctx context.Context, username string) (*User, error) {
                hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
                return &User{ID: 1, Username: "john", Password: string(hashed)}, nil
            },
            wantErr: ErrInvalidCredentials,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := &MockUserRepository{FindByUsernameFn: tt.mockFn}
            service := NewUserService(repo, logrus.New())
            
            user, err := service.Authenticate(context.Background(), tt.username, tt.password)
            
            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
                assert.Nil(t, user)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, user)
            }
        })
    }
}
```

### 6.2 Integration Testing

```go
// ✅ Senior: Integration test với test database
func TestUserRepository_FindByUsername(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(db)
    
    // Seed data
    user := &User{Username: "testuser", Email: "test@example.com"}
    db.Create(user)
    
    // Test
    repo := NewUserRepository(db, logrus.New())
    found, err := repo.FindByUsername(context.Background(), "testuser")
    
    assert.NoError(t, err)
    assert.Equal(t, user.ID, found.ID)
    assert.Equal(t, user.Username, found.Username)
}
```

### 6.3 API Testing

```go
// ✅ Senior: API test với Fiber test utils
func TestLoginEndpoint(t *testing.T) {
    // Setup app
    app := setupTestApp(t)
    
    // Test cases
    tests := []struct {
        name       string
        body       LoginBody
        wantStatus int
        wantErr    string
    }{
        {
            name:       "success",
            body:       LoginBody{Username: "john", Password: "password123"},
            wantStatus: 200,
        },
        {
            name:       "invalid credentials",
            body:       LoginBody{Username: "john", Password: "wrong"},
            wantStatus: 401,
            wantErr:    "Invalid credentials",
        },
        {
            name:       "missing username",
            body:       LoginBody{Password: "password123"},
            wantStatus: 400,
            wantErr:    "Username is required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            body, _ := json.Marshal(tt.body)
            req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
            req.Header.Set("Content-Type", "application/json")
            
            resp, err := app.Test(req)
            assert.NoError(t, err)
            assert.Equal(t, tt.wantStatus, resp.StatusCode)
            
            if tt.wantErr != "" {
                var errResp ErrorResponse
                json.NewDecoder(resp.Body).Decode(&errResp)
                assert.Contains(t, errResp.Message, tt.wantErr)
            }
        })
    }
}
```

---

## 🎯 Phần 7: Real-World Patterns từ vtvcab-identity-api

### 7.1 JWT Token Flow

```go
// Production-ready JWT implementation
type TokenService struct {
    redis        *redis.Client
    secret       []byte
    accessTTL    time.Duration
    refreshTTL   time.Duration
    logger       *logrus.Logger
}

func (s *TokenService) Generate(ctx context.Context, claims *TokenClaims) (*TokenPair, error) {
    // Generate unique JTI
    accessJTI := uuid.New().String()
    refreshJTI := uuid.New().String()
    
    // Create access token
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":  claims.UserID,
        "jti":  accessJTI,
        "type": "access",
        "exp":  time.Now().Add(s.accessTTL).Unix(),
        "iat":  time.Now().Unix(),
    })
    accessTokenString, err := accessToken.SignedString(s.secret)
    if err != nil {
        return nil, fmt.Errorf("sign access token: %w", err)
    }
    
    // Create refresh token
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":  claims.UserID,
        "jti":  refreshJTI,
        "type": "refresh",
        "exp":  time.Now().Add(s.refreshTTL).Unix(),
        "iat":  time.Now().Unix(),
    })
    refreshTokenString, err := refreshToken.SignedString(s.secret)
    if err != nil {
        return nil, fmt.Errorf("sign refresh token: %w", err)
    }
    
    // Store tokens in Redis (async)
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        // Store access token
        s.redis.Set(ctx, fmt.Sprintf("token:%s", accessJTI), claims.UserID, s.accessTTL)
        
        // Store refresh token
        s.redis.Set(ctx, fmt.Sprintf("token:%s", refreshJTI), claims.UserID, s.refreshTTL)
        
        // Link refresh to access
        s.redis.Set(ctx, fmt.Sprintf("token:parent:%s", accessJTI), refreshJTI, s.accessTTL)
    }()
    
    return &TokenPair{
        AccessToken:  accessTokenString,
        RefreshToken: refreshTokenString,
    }, nil
}

func (s *TokenService) Validate(ctx context.Context, tokenString string) (*TokenClaims, error) {
    // Parse token
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return s.secret, nil
    })
    if err != nil {
        return nil, fmt.Errorf("parse token: %w", err)
    }
    
    // Extract claims
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }
    
    // Check token type
    tokenType, _ := claims["type"].(string)
    if tokenType != "access" {
        return nil, ErrInvalidTokenType
    }
    
    // Check JTI in Redis (not revoked)
    jti, _ := claims["jti"].(string)
    exists, err := s.redis.Exists(ctx, fmt.Sprintf("token:%s", jti)).Result()
    if err != nil {
        return nil, fmt.Errorf("check token: %w", err)
    }
    if exists == 0 {
        return nil, ErrTokenRevoked
    }
    
    return &TokenClaims{
        UserID: uint(claims["sub"].(float64)),
        JTI:    jti,
    }, nil
}
```

### 7.2 Device Session Management

```go
// Production-ready session management
type SessionService struct {
    mongo  *mongo.Database
    redis  *redis.Client
    logger *logrus.Logger
}

func (s *SessionService) CreateSession(ctx context.Context, userID uint, deviceInfo *DeviceInfo) (*Session, error) {
    session := &Session{
        ID:        primitive.NewObjectID(),
        UserID:    userID,
        JTI:       uuid.New().String(),
        DeviceID:  deviceInfo.DeviceID,
        DeviceOS:  deviceInfo.DeviceOS,
        UserAgent: deviceInfo.UserAgent,
        IP:        deviceInfo.IP,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
    }
    
    // Check device limit
    count, err := s.mongo.Collection("sessions").CountDocuments(ctx, bson.M{
        "user_id":    userID,
        "expires_at": bson.M{"$gt": time.Now()},
    })
    if err != nil {
        return nil, fmt.Errorf("count sessions: %w", err)
    }
    
    if count >= 5 { // Max 5 devices
        // Remove oldest session
        _, err := s.mongo.Collection("sessions").DeleteOne(ctx, bson.M{
            "user_id":    userID,
            "expires_at": bson.M{"$gt": time.Now()},
        }, options.Delete().SetSort(bson.M{"created_at": 1}))
        if err != nil {
            return nil, fmt.Errorf("remove oldest session: %w", err)
        }
    }
    
    // Insert new session
    _, err = s.mongo.Collection("sessions").InsertOne(ctx, session)
    if err != nil {
        return nil, fmt.Errorf("insert session: %w", err)
    }
    
    return session, nil
}
```

---

## 📚 Tổng kết: Senior Go Dev Checklist

### Code Quality
- [ ] Error handling nhất quán (không bỏ qua errors)
- [ ] Structured logging (không fmt.Println)
- [ ] Input validation (không tin tưởng client)
- [ ] Type safety (không dùng interface{} nếu có thể)

### Architecture
- [ ] Dependency injection (không dùng globals)
- [ ] Service layer (business logic không nằm trong controller)
- [ ] Repository pattern (data access tách riêng)
- [ ] Middleware composition (cross-cutting concerns)

### Performance
- [ ] Connection pooling (DB, Redis)
- [ ] Caching strategy (cache-aside)
- [ ] Async processing (goroutines cho background tasks)
- [ ] Resource limits (body size, rate limiting)

### Security
- [ ] Input validation & sanitization
- [ ] SQL injection prevention (parameterized queries)
- [ ] Rate limiting
- [ ] Security headers

### Observability
- [ ] Structured logging với context
- [ ] Metrics collection
- [ ] Distributed tracing (APM)
- [ ] Error tracking (Sentry)

### Testing
- [ ] Unit tests với mocks
- [ ] Integration tests với test DB
- [ ] API tests
- [ ] Test coverage > 70%

---

## 🔗 Cross-references

- [[01-Project-Structure-Analysis]] - Xem project architecture
- [[02-Go-vs-NodeJS-Comparison]] - So sánh Go vs NodeJS
- [[03-Go-Fundamentals-for-NodeJS-Devs]] - Go basics
- [[04-Practical-Patterns-from-Project]] - Patterns thực tế
- [[05-Quick-Reference-CheatSheet]] - Cheat sheet
- [[06-Fiber-Framework-Complete-Guide]] - Fiber deep dive
