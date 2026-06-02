# 📁 Phân Tích Project Structure - vtvcab-identity-api

> **Dành cho:** NodeJS dev muốn hiểu Go project architecture
> **Mục tiêu:** Hiểu cách một Go web API được tổ chức và hoạt động

---

## 🗺️ Directory Tree Overview

```
vtvcab-identity-api/
├── main.go                          # 🚀 Entry point - App bootstrap
├── go.mod                           # 📦 Dependencies (giống package.json)
├── go.sum                           # 📦 Lock file (giống package-lock.json)
├── test.go                          # ⚠️ Scratch code - KHÔNG phải app chính
├── Dockerfile                       # 🐳 Docker build
├── .env.example                     # 📝 Env vars template
│
├── app/                             # 📱 Application Layer
│   ├── controllers/                 # 🎮 HTTP Handlers (giống Express routes)
│   │   ├── login.go                 #    Login, register, OTP, token
│   │   └── user.go                  #    User CRUD, device management
│   │
│   ├── repository/                  # 💾 Data Access Layer
│   │   ├── postgres_repo.go         #    PostgreSQL queries (GORM)
│   │   ├── redis_repo.go            #    Redis cache operations
│   │   ├── sms_repo.go              #    SMS OTP sending
│   │   └── social_repo.go           #    OAuth social login
│   │
│   ├── service/                     # ⚙️ Business Logic Layer
│   │   ├── create_user_core.go      #    User creation logic
│   │   ├── deactive_user.go         #    User deactivation
│   │   └── device_session.go        #    Device session management
│   │
│   ├── models/                      # 📊 Data Models (giống TypeScript interfaces)
│   │   ├── user_entity.go           #    UserAccount, UserSocialAccount structs
│   │   ├── request.go               #    Request body structs
│   │   ├── response.go              #    Response structs (Google, Facebook, Apple)
│   │   ├── session_entity.go        #    Session device model
│   │   └── user_status.go           #    User block status
│   │
│   ├── helpers/                     # 🔧 OAuth Helpers
│   │   ├── apple.go                 #    Apple Sign-In
│   │   ├── facebook.go              #    Facebook OAuth
│   │   └── google.go                #    Google OAuth
│   │
│   ├── producer/                    # 📤 RabbitMQ Publisher
│   │   └── producer.go              #    AMQP connection & publish
│   │
│   ├── worker/                      # 📥 RabbitMQ Consumer
│   │   ├── event_log_worker.go      #    Event log consumer
│   │   └── model.go                 #    Event log model
│   │
│   └── testing/                     # 🧪 Integration Tests
│       ├── common_test.go           #    Test utilities
│       └── token_test.go            #    Token tests
│
├── platform/                        # 🏗️ Infrastructure Layer
│   ├── database/                    #    Database connections
│   │   ├── postgres.go              #    PostgreSQL (GORM + APM)
│   │   ├── mongodb.go               #    MongoDB connection
│   │   └── mssql.go                 #    MSSQL connection
│   │
│   └── cache/                       #    Cache connections
│       └── redis.go                 #    Redis connection (3 instances!)
│
└── settings/                        # ⚙️ Configuration & Middleware
    ├── config/                      #    App configuration
    │   └── config.go                #    Viper config loader
    │
    ├── middleware/                   #    HTTP Middleware
    │   ├── jwt.go                   #    JWT auth, token generation
    │   ├── token.go                 #    Token manager (Redis)
    │   ├── session.go               #    Device session manager (MongoDB)
    │   ├── device_session.go        #    Device session middleware
    │   └── sentry.go                #    Sentry error tracking
    │
    ├── router/                      #    Route definitions
    │   ├── versioned/
    │   │   ├── v1/api.go            #    V1 routes (~30+ endpoints)
    │   │   └── v2/api.go            #    V2 routes (subset + event logging)
    │   └── exeption.go              #    404 handler
    │
    ├── pagination/                  #    Pagination helper
    │   └── pagination.go
    │
    ├── utils/                       #    Utility functions
    │   ├── response.go              #    HTTP response helpers
    │   ├── validator.go             #    Request validation
    │   ├── common.go                #    Common utilities
    │   ├── extension.go             #    Crypto extensions
    │   ├── connection_url_builder.go #   DB URL builder
    │   ├── enum.go                  #    Constants & enums
    │   ├── convert_time.go          #    Time utilities
    │   ├── covert_thumb.go          #    Thumbnail utilities
    │   └── start_server.go          #    Graceful shutdown
    │
    ├── settings.dev.toml            #    Dev config
    ├── settings.staging.toml        #    Staging config
    ├── settings.prod.toml           #    Prod config
    ├── privateKey.pem               #    JWT signing key
    ├── appleid_auth.pem             #    Apple auth key
    └── interactiveKey.pem           #    Interactive key
```

---

## 🚀 Entry Point Analysis - main.go

### NodeJS vs Go Comparison

```javascript
// NodeJS (Express)
const app = express();
app.use(cors());
app.use(authMiddleware);
app.post('/api/login', loginController);
app.listen(3000);
```

```go
// Go (Fiber)
type App struct {
    *fiber.App                    // Embed Fiber app
    postgres   *gorm.DB           // Database connections
    redis      *redis.Client      // Cache
    mongoDb    *mongo.Database    // MongoDB
    // ... more dependencies
}

func main() {
    // 1. Connect databases
    pos, _ := database.PostgresConnection()
    redisStore, _ := cache.RedisConnection(os.Getenv("REDIS_URL"), 250)
    
    // 2. Create app with dependencies
    app := App{
        App:      fiber.New(),
        postgres: pos,
        redis:    redisStore,
    }
    
    // 3. Register routes with dependencies injected
    api.RegisterAPI(groupAPI, app.postgres, app.redis, ...)
    
    // 4. Start server
    utils.StartServerWithGracefulShutdown(app.App)
}
```

### Key Differences từ NodeJS:
1. **Dependency Injection rõ ràng** - Dependencies được inject qua function params, không phải global imports
2. **Multiple Database Connections** - Go quản lý nhiều DB connections cùng lúc
3. **Graceful Shutdown** - Go có built-in graceful shutdown pattern

---

## 🏗️ Key Architectural Patterns

### 1. Repository Pattern (Data Access Layer)

```
Controller → Service → Repository → Database
```

```go
// Repository struct chứa database connection
type PostgresQueries struct {
    Conn   *gorm.DB        // PostgreSQL connection
    Client *redis.Client   // Redis client (optional)
}

// Constructor function (giống class constructor trong JS)
func NewPostgresQueries(conn *gorm.DB) *PostgresQueries {
    return &PostgresQueries{Conn: conn}
}

// Method trên struct (giống class method)
func (query *PostgresQueries) LoginAccountDb(ctx context.Context, req *models.LoginBody) (*models.UserAccount, error) {
    user := &models.UserAccount{}
    // GORM query
    result := query.Conn.Where("name = ?", req.UserName).First(&user)
    return user, result.Error
}
```

**NodeJS Equivalent:**
```javascript
class PostgresRepository {
    constructor(db) {
        this.db = db;
    }
    
    async loginAccount(req) {
        const user = await this.db('users').where({ name: req.userName }).first();
        return user;
    }
}
```

### 2. Controller Pattern (HTTP Handlers)

```go
// Controller là function trả về fiber.Handler
func Login(sql *gorm.DB, redisToken *redis.Client, mongo *mongo.Database) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 1. Parse request body
        body := &models.LoginBody{}
        err := c.BodyParser(&body)
        
        // 2. Call repository
        connect := service.PostgresQueries{Conn: sql}
        user, err := connect.LoginAccountDb(c.Context(), body)
        
        // 3. Return response
        return utils.HTTPSuccessResponse(c, fiber.StatusOK, map[string]interface{}{
            "access_token": token,
            "user_id":      user.Id,
        })
    }
}
```

**NodeJS Equivalent:**
```javascript
const login = async (req, res) => {
    const body = req.body;
    const user = await postgresRepo.loginAccount(body);
    res.json({ success: true, data: { access_token: token } });
};
```

### 3. Middleware Pattern

```go
// Middleware là function kiểm tra request trước khi đến controller
func VerifyToken(redisToken *redis.Client, mongo *mongo.Database) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 1. Extract token từ header
        tokenAuth, err := ValidateExtractToken(c)
        
        // 2. Validate token
        userClaims, err := ParseTokenClaims(tokenAuth)
        
        // 3. Kiểm tra token trong Redis
        tokenManager, _ := NewTokenManager(c.Context(), redisToken)
        err = validateToken(c.Context(), userClaims.Jti, ...)
        
        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(...)
        }
        
        // 4. Tiếp tục đến controller
        return c.Next()
    }
}

// Sử dụng trong route
api.Post("login", controller.Login(...))
api.Get("my", middleware.VerifyToken(...), controller.GetMy(...))  // Protected route
```

---

## 🔄 Data Flow - Request Lifecycle

```
Client Request
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│  Fiber Router                                                │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Middleware Chain                                     │    │
│  │  1. APM Middleware (tracing)                          │    │
│  │  2. CORS Middleware                                   │    │
│  │  3. Logger Middleware                                 │    │
│  │  4. JWT Verify Token (nếu là protected route)        │    │
│  └─────────────────────────────────────────────────────┘    │
│                           │                                   │
│                           ▼                                   │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Controller (HTTP Handler)                           │    │
│  │  1. Parse request body/query                         │    │
│  │  2. Validate input                                   │    │
│  │  3. Call Service/Repository                          │    │
│  │  4. Return response                                  │    │
│  └─────────────────────────────────────────────────────┘    │
│                           │                                   │
│                           ▼                                   │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Repository (Data Access)                            │    │
│  │  1. Build query                                      │    │
│  │  2. Execute database operation                       │    │
│  │  3. Return result/error                              │    │
│  └─────────────────────────────────────────────────────┘    │
│                           │                                   │
│                           ▼                                   │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Database Layer                                      │    │
│  │  • PostgreSQL (users, accounts)                      │    │
│  │  • Redis (cache, tokens, ban chat)                   │    │
│  │  • MongoDB (device sessions)                         │    │
│  │  • MSSQL (legacy data)                               │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### Ví dụ cụ thể: Login Flow

```go
// 1. Client gửi POST /api/login
// Route: api.Post("login/postgres", controller.Login(pos, redisToken, mongoDb))

// 2. Controller nhận request
func Login(sql *gorm.DB, redisToken *redis.Client, mongo *mongo.Database) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 3. Parse body
        body := &models.LoginBody{}
        c.BodyParser(&body)
        
        // 4. Query PostgreSQL
        connect := service.PostgresQueries{Conn: sql}
        user, _ := connect.LoginAccountDb(c.Context(), body)
        
        // 5. Generate JWT token
        userClaims := middleware.ConvertMapClaims(user)
        token, _ := userClaims.GenerateToken(mongo, c, redisToken)
        
        // 6. Return response
        return utils.HTTPSuccessResponse(c, fiber.StatusOK, token)
    }
}
```

---

## 📦 External Dependencies (go.mod)

### Core Framework
| Package | Purpose | NodeJS Equivalent |
|---------|---------|-------------------|
| `github.com/gofiber/fiber/v2` | HTTP framework | Express.js |
| `gorm.io/gorm` | ORM | Sequelize, TypeORM |
| `github.com/go-redis/redis/v8` | Redis client | ioredis |
| `go.mongodb.org/mongo-driver` | MongoDB driver | mongoose |
| `github.com/golang-jwt/jwt/v4` | JWT library | jsonwebtoken |

### Configuration & Logging
| Package | Purpose | NodeJS Equivalent |
|---------|---------|-------------------|
| `github.com/spf13/viper` | Config management | dotenv + config |
| `github.com/sirupsen/logrus` | Structured logging | winston, pino |
| `github.com/joho/godotenv` | .env loader | dotenv |

### Monitoring & Security
| Package | Purpose | NodeJS Equivalent |
|---------|---------|-------------------|
| `go.elastic.co/apm/v2` | APM tracing | elastic-apm-node |
| `github.com/getsentry/sentry-go` | Error tracking | @sentry/node |
| `golang.org/x/crypto` | Crypto utilities | bcrypt, crypto |

### OAuth & External Services
| Package | Purpose | NodeJS Equivalent |
|---------|---------|-------------------|
| `golang.org/x/oauth2` | OAuth2 client | passport |
| `google.golang.org/api` | Google API | googleapis |
| `github.com/streadway/amqp` | RabbitMQ | amqplib |

---

## ⚙️ Config System - Viper + TOML + .env

### How It Works

```
.env (environment variables)
    ↓
settings.{dev,staging,prod}.toml (app config)
    ↓
Viper (config loader)
    ↓
Config struct (accessible throughout app)
```

### Config Loading Flow

```go
// 1. .env auto-loaded via godotenv
_ "github.com/joho/godotenv/autoload"  // In main.go import

// 2. STAGE_STATUS env var determines which TOML to load
func LoadSetting(env string) *Config {
    switch env {
    case "prod":
        config.SetConfigName("settings.prod")
    case "staging":
        config.SetConfigName("settings.staging")
    default:
        config.SetConfigName("settings.dev")
    }
    config.SetConfigType("toml")
    config.ReadInConfig()
}

// 3. Config accessible via Viper
config := configs.LoadSetting(os.Getenv("STAGE_STATUS"))
jwtSecret := config.GetString("JWT_SECRET_KEY")
accessTokenExpiry := config.GetInt("JWT_ACCESS_TOKEN_EXPIRES")
```

### TOML File Structure (settings.dev.toml)

```toml
[database]
POSTGRES_URL = "postgres://user:pass@localhost:5432/db"
MONGODB_URL = "mongodb://localhost:27017"
MSSQL_URL = "sqlserver://..."

[redis]
REDIS_URL = "redis://localhost:6379/0"
REDIS_URL_TOKEN = "redis://localhost:6379/1"
REDIS_URL_BAN_CHAT = "redis://localhost:6379/2"

[jwt]
JWT_SECRET_KEY = "your-secret-key"
JWT_ACCESS_TOKEN_EXPIRES = 3600
JWT_REFRESH_TOKEN_EXPIRES = 300
```

---

## 🗄️ Multi-Database Architecture

### Why 3 Redis Instances?

```
┌─────────────────────────────────────────────────────────────┐
│  Redis Instance 1 (REDIS_URL)                                │
│  └── Cache Store                                             │
│      • User info cache                                       │
│      • Session data                                          │
│      • Temporary data                                        │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  Redis Instance 2 (REDIS_URL_TOKEN)                          │
│  └── Token Store                                             │
│      • JWT access tokens                                     │
│      • JWT refresh tokens                                    │
│      • Token blacklist                                       │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  Redis Instance 3 (REDIS_URL_BAN_CHAT)                       │
│  └── Ban Chat Store                                          │
│      • Banned user list                                      │
│      • Chat restrictions                                     │
└─────────────────────────────────────────────────────────────┘
```

### Why Multiple Databases?

| Database | Purpose | Data |
|----------|---------|------|
| **PostgreSQL** | Primary database | User accounts, social accounts, SmartTV codes |
| **MongoDB** | Document store | Device sessions (flexible schema) |
| **Redis** | Cache + Token store | User cache, JWT tokens, ban list |
| **MSSQL** | Legacy integration | Legacy data (light usage) |

### Connection Setup

```go
// PostgreSQL - Connection pooling with APM
func PostgresConnection() (*gorm.DB, error) {
    db, _ := gorm.Open(postgresApm.Open(url), &gorm.Config{
        SkipDefaultTransaction: true,
        PrepareStmt:            true,
    })
    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(5)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    return db, nil
}

// Redis - 3 separate instances
redisStore, _ := cache.RedisConnection(os.Getenv("REDIS_URL"), 250)
redisToken, _ := cache.RedisConnection(os.Getenv("REDIS_URL_TOKEN"), 250)
redisBanChat, _ := cache.RedisConnection(os.Getenv("REDIS_URL_BAN_CHAT"), 150)
```

---

## 🎯 Key Takeaways cho NodeJS Dev

### 1. **Go dùng Struct thay vì Class**
```go
type UserAccount struct {
    Id       uint   `json:"id"`
    Name     string `json:"name"`
    Phone    string `json:"phone"`
}

// Methods defined outside struct
func (user *UserAccount) HashPassword(password string) error {
    // ...
}
```

### 2. **Go dùng Error Return thay vì Try/Catch**
```go
user, err := connect.LoginAccountDb(ctx, body)
if err != nil {
    return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)
}
```

### 3. **Go dùng Dependency Injection rõ ràng**
```go
// Dependencies được inject qua function parameters
func Login(sql *gorm.DB, redisToken *redis.Client, mongo *mongo.Database) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Use sql, redisToken, mongo here
    }
}
```

### 4. **Go dùng Context cho cancellation**
```go
func (query *PostgresQueries) LoginAccountDb(ctx context.Context, req *models.LoginBody) (*models.UserAccount, error) {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    // Use ctx for database queries
}
```

### 5. **Go dùng Goroutines cho async operations**
```go
// Fire and forget (giống Promise không await)
go tokenManager.SaveToken(accessJti, int(timeExpire), int(claims.Id), ...)

// Channel để communicate giữa goroutines
chEVLog := make(chan worker.EventLog)
go worker.Consume(chEVLog)
```

---

## 📚 Tiếp theo

- [[02-Go-vs-NodeJS-Comparison]] - So sánh chi tiết Go vs NodeJS
- [[03-Go-Fundamentals-for-NodeJS-Devs]] - Go fundamentals cho NodeJS dev
- [[04-Practical-Patterns-from-Project]] - Patterns thực tế trong project
- [[05-Quick-Reference-CheatSheet]] - Cheat sheet tra cứu nhanh
