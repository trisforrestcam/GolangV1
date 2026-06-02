# Go Quick Reference Cheat Sheet — vtvcab-identity-api

> **Module:** `vtvcab-identity-api` | **Go version:** 1.21 | **Framework:** Fiber v2

---

## 1. Essential Go Commands

```bash
# Build & Run
go build .                        # Compile binary
go run .                          # Run locally (needs .env + infra)
STAGE_STATUS=dev go run .         # Run with specific environment (dev/staging/prod)

# Dependencies
go mod tidy                       # Remove unused deps, add missing ones
go mod download                   # Download all deps to cache
go mod verify                     # Check deps match go.sum

# Static Analysis & Testing
go vet ./...                      # Static analysis (only check configured in this repo)
go test ./app/testing/ -v         # Run integration tests (requires live infra)

# Useful one-offs
go list -m all                    # List all dependencies
go doc fiber.Ctx                  # View docs for a symbol
```

---

## 2. Project-Specific Commands

### Run Locally

You need all infrastructure running + `.env` configured. Required env vars (see `.env.example`):

```bash
# Database
POSTGRES_URL=postgres://user:pass@host:5432/dbname
MONGODB_URL=mongodb://user:pass@host/dbname
MSSQL_URL=...                      # (also connected but lightly used)

# Redis (THREE separate instances!)
REDIS_URL=redis://...              # Cache store
REDIS_URL_TOKEN=redis://...        # JWT token store
REDIS_URL_BAN_CHAT=redis://...     # Ban/chat store

# Message Queue
RABBIT_URL=amqp://...
RABBIT_QUEUE_NAME=task_queue

# Environment
STAGE_STATUS=dev                   # dev | staging | prod (selects settings.{env}.toml)
```

```bash
# Run
STAGE_STATUS=dev go run .

# Or with explicit .env
cp .env.example .env
# Edit .env with local values
STAGE_STATUS=dev go run .
```

### Run Tests

```bash
# Tests require live Redis, MongoDB, Postgres — no mocks
STAGE_STATUS=dev go test ./app/testing/ -v
```

### Build Docker Image

```bash
docker build -t vtvcab-identity-api .
docker run -p 5000:5000 --env-file .env vtvcab-identity-api
```

> **Note:** Dockerfile uses `golang:1.19-buster` but `go.mod` says `go 1.21`. Config TOML files and PEM keys are copied into the image.

---

## 3. Common Go Syntax Quick Reference

### Variable Declaration

```go
// Short declaration (inside functions)
name := "hello"
count := 42

// Var declaration
var name string
var count int = 42

// Multiple variables
var (
    host string = "localhost"
    port int    = 5000
)

// Zero values: string="", int=0, bool=false, pointer/slice/map=nil
```

### Function Signatures

```go
// Basic function
func DoSomething(name string) error {
    return nil
}

// Multiple return values
func GetUser(id uint) (*UserAccount, error) {
    // ...
}

// Variadic function
func Sum(nums ...int) int { /* ... */ }

// Function as parameter
func Handler(fn func(*fiber.Ctx) error) { /* ... */ }

// Method on struct
func (u *UserAccount) HashPassword(password string) error {
    // ...
}
```

### Struct Definition

```go
type UserAccount struct {
    Id           uint      `gorm:"primaryKey;autoIncrement:true;unique" json:"id"`
    Name         *string   `gorm:"index:idx_user_name,unique" json:"name"`
    Phone        *string   `gorm:"index:idx_user_phone,unique" json:"phone"`
    FullName     string    `gorm:"column:fullname" json:"fullname"`
    PasswordHash string    `gorm:"column:password_hash" json:"password_hash"`
    CreatedAt    time.Time `sql:"DEFAULT:'current_timestamp'" json:"-"`
}

// Custom table name
func (UserAccount) TableName() string {
    return "users"
}
```

### Interface Definition

```go
type Repository interface {
    GetUser(ctx context.Context, id uint) (*UserAccount, error)
    CreateUser(ctx context.Context, user *UserAccount) error
}
```

### Error Handling Pattern (this project's convention)

```go
// Standard pattern in this codebase
result, err := someOperation()
if err != nil {
    log.WithError(err).Error("Something went wrong")
    return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)
}
```

### Goroutine Syntax

```go
go func() {
    // runs concurrently
    worker.DoWork()
}()

// With channel communication
go func(ch chan EventLog) {
    for msg := range ch {
        process(msg)
    }
}(chEVLog)
```

### Channel Syntax

```go
// Create channel
ch := make(chan EventLog)        // unbuffered
ch := make(chan EventLog, 100)   // buffered

// Send
ch <- eventLog

// Receive
msg := <-ch

// Range over channel
for msg := range ch {
    process(msg)
}
```

### Slice & Map Operations

```go
// Slice
roles := []string{"admin", "user"}
roles = append(roles, "moderator")
first := roles[0]
sub := roles[1:3]          // [1, 3)

// Map
data := map[string]interface{}{
    "key": "value",
}
val, ok := data["key"]     // ok = false if not found
delete(data, "key")
```

---

## 4. HTTP Status Codes in Fiber

```go
// Success
fiber.StatusOK                  // 200
fiber.StatusCreated             // 201
fiber.StatusNoContent           // 204

// Client Error
fiber.StatusBadRequest          // 400
fiber.StatusUnauthorized        // 401
fiber.StatusForbidden           // 403
fiber.StatusNotFound            // 404
fiber.StatusConflict            // 409
fiber.StatusUnprocessableEntity // 422
fiber.StatusTooManyRequests     // 429

// Server Error
fiber.StatusInternalServerError  // 500
fiber.StatusServiceUnavailable  // 503
```

---

## 5. Common Pattern Snippets

### Parse JSON Request Body

```go
func Login(c *fiber.Ctx) error {
    body := &models.LoginBody{}
    if err := c.BodyParser(&body); err != nil {
        log.WithError(err).Error("Something went wrong")
        return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)
    }
    // use body...
}
```

### Validate Request Body

```go
if err := utils.ValidateStruct(body); err != nil {
    return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)
}
```

### Return JSON Response

```go
// Success response (standard pattern)
return utils.HTTPSuccessResponse(c, fiber.StatusOK, data)

// Error response
return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)

// Custom response
return c.Status(fiber.StatusOK).JSON(&fiber.Map{
    "success": true,
    "data":    result,
})
```

### Get URL Params

```go
// Route: /api/user/:id
id := c.Params("id")

// Route: /api/user/:id/action/:action
id := c.Params("id")
action := c.Params("action")
```

### Get Query Params

```go
// URL: /api/users?search=john&phone=090
search := c.Query("search")
phone := c.Query("phone")
page := c.Query("page", "1")    // with default value

// Bind query params to struct
params := &models.SearchParams{}
if err := c.QueryParser(params); err != nil {
    return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)
}
```

### Set Headers

```go
// Read header
auth := c.Get("Authorization")
contentType := c.Get("Content-Type")

// Set response header
c.Set("X-Custom-Header", "value")
```

### Read from Redis

```go
// Get
val, err := redisClient.Get(c.Context(), "key").Result()
if err == redis.Nil {
    // key does not exist
} else if err != nil {
    // error
}

// Set with TTL
err := redisClient.Set(c.Context(), "key", "value", 10*time.Minute).Err()
```

### Write to Postgres (GORM)

```go
// Inject DB: sql *gorm.DB (passed through handler closure)
connect := service.PostgresQueries{Conn: sql.WithContext(c.Context())}

// Query one record
user := &models.UserAccount{}
result := sql.WithContext(c.Context()).Where("phone = ?", phone).First(user)
if result.Error != nil {
    // handle error
}

// Create
sql.WithContext(c.Context()).Create(&user)

// Update
sql.WithContext(c.Context()).Model(&user).Update("name", "new name")

// Transaction
tx := sql.WithContext(c.Context()).Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()
// ... operations ...
tx.Commit()
```

### MongoDB Operations

```go
// Inject: mongo *mongo.Database (passed through handler)

// Find one
collection := mongo.Collection("session_device")
var session models.UserSession
err := collection.FindOne(ctx, bson.M{"jti": jti}).Decode(&session)

// Find many
cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
if err != nil { /* handle */ }
defer cursor.Close(ctx)

var sessions []models.UserSession
if err = cursor.All(ctx, &sessions); err != nil { /* handle */ }

// Insert
_, err := collection.InsertOne(ctx, session)

// Update
_, err := collection.UpdateOne(ctx,
    bson.M{"jti": jti},
    bson.M{"$set": bson.M{"revoked": true}},
)

// Delete
_, err := collection.DeleteOne(ctx, bson.M{"jti": jti})
```

### Extract JWT Claims

```go
userClaims, err := middleware.ExtractTokenMetadata(c)
if err != nil {
    return utils.HTTPFailResponse(c, fiber.StatusUnauthorized, err)
}
userID := userClaims.Id
userName := userClaims.Name
roles := userClaims.Roles
```

### Make HTTP Request (external API)

```go
client := &http.Client{}
HTTPClient := apmhttp.WrapClient(client)  // with APM tracing
req, err := http.NewRequest("GET", url, nil)
if err != nil { /* handle */ }
req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
res, err := HTTPClient.Do(req.WithContext(c.Context()))
if err != nil { /* handle */ }
defer res.Body.Close()
body, _ := ioutil.ReadAll(res.Body)
var result map[string]interface{}
json.Unmarshal(body, &result)
```

---

## 6. Debugging Tips

### Logging with Logrus (this project's logger)

```go
import log "github.com/sirupsen/logrus"

// Structured fields
log.WithFields(log.Fields{
    "user_id": id,
    "action":  "login",
}).Info("User logged in")

// With error
log.WithError(err).Error("Something went wrong")

// Debug
log.WithFields(log.Fields{"data": result}).Debug("Processing complete")

// Warning
log.Warn("Deprecated endpoint called")

// Global logger (JSON format in production — set in main.go)
log.SetFormatter(&log.JSONFormatter{})
```

### Print Debugging

```go
fmt.Printf("DEBUG: %+v\n", myStruct)    // print with field names
fmt.Printf("DEBUG: %#v\n", myStruct)    // Go syntax representation
fmt.Println("DEBUG:", someVar)           // simple print
```

### Common Error Messages & Solutions

| Error | Cause | Fix |
|-------|-------|-----|
| `dial tcp: connect: connection refused` | Infrastructure not running | Start Postgres/Redis/MongoDB |
| `redis: nil` | Key not found in Redis | Check key exists, verify correct Redis instance (cache vs token vs ban_chat) |
| `record not found` | GORM query returned no rows | Add `.Error` check, handle `gorm.ErrRecordNotFound` |
| `invalid token` | JWT validation failed | Check token isn't expired, verify Redis token store is accessible |
| `EOF` from MongoDB | MongoDB connection lost | Check MongoDB is running, verify MONGODB_URL |

### Adding Sentry Error Tracking

```go
// Errors sent to Sentry automatically via utils.SendErrorApm()
// Manual capture:
sentry.CaptureException(err)
```

---

## 7. Frequently Used Imports

```go
// Standard library
import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "time"
)

// Project internal (module: vtvcab-identity-api)
import (
    "vtvcab-identity-api/app/models"
    "vtvcab-identity-api/app/repository"
    "vtvcab-identity-api/app/service"
    "vtvcab-identity-api/app/worker"
    "vtvcab-identity-api/platform/cache"
    "vtvcab-identity-api/platform/database"
    configs "vtvcab-identity-api/settings/config"
    "vtvcab-identity-api/settings/middleware"
    "vtvcab-identity-api/settings/utils"
)

// Third-party (common in this project)
import (
    "github.com/go-redis/redis/v8"           // Redis client
    "github.com/gofiber/fiber/v2"             // HTTP framework
    "github.com/sirupsen/logrus"              // Structured logging
    "go.mongodb.org/mongo-driver/bson"        // MongoDB BSON (use bson.M)
    "go.mongodb.org/mongo-driver/mongo"       // MongoDB driver
    "gorm.io/gorm"                            // ORM for Postgres/MSSQL
    "github.com/go-playground/validator/v10"  // Struct validation
    "github.com/golang-jwt/jwt/v4"            // JWT tokens
)
```

---

## 8. File & Naming Conventions

| Element | Convention | Example |
|---------|-----------|---------|
| File names | `snake_case.go` | `user_entity.go`, `token_test.go` |
| Package names | lowercase, single word | `models`, `controller`, `utils` |
| Exported names | `CamelCase` | `UserAccount`, `HTTPSuccessResponse` |
| Unexported names | `camelCase` | `signKey`, `defaultErrorHandler` |
| Constants | `CamelCase` or `SCREAMING_SNAKE` | `Male`, `JWT_ACCESS_TOKEN_EXPIRES` |
| Interface names | `CamelCase` (often `-er` suffix) | `Repository` |
| Struct tags | `key:"value"` with spaces | `` `gorm:"primaryKey" json:"id"` `` |

### Project Directory Structure

```
app/
  controllers/          → HTTP handlers (login.go, user.go)
  repository/           → Data access (postgres queries, redis, sms, social)
  service/              → Business logic (create_user, deactive_user, device_session)
  models/               → Request/response structs, DB entities
  helpers/              → OAuth helpers (Apple, Facebook, Google)
  producer/             → RabbitMQ publisher
  worker/               → RabbitMQ consumer (event logging)
  testing/              → Integration tests
platform/
  database/             → DB connections (postgres.go, mongodb.go, mssql.go)
  cache/                → Redis connection
settings/
  config/               → Viper config loader
  router/versioned/     → Route registration (v1/api.go, v2/api.go)
  middleware/           → JWT auth, device sessions, Sentry
  utils/                → Response helpers, validators
```

---

## 9. Key Gotchas Quick Reference

1. **Three Redis instances** — `REDIS_URL` (cache), `REDIS_URL_TOKEN` (JWT), `REDIS_URL_BAN_CHAT`. Don't mix them up!
2. **`.env` auto-loaded** — via `godotenv/autoload` import in `main.go`
3. **Config files baked into Docker** — `settings.{dev,staging,prod}.toml` + PEM keys are `COPY`'d in Dockerfile
4. **JWT access/refresh ratio** — refresh token TTL = 90× access token TTL
5. **V1 vs V2 routes** — V1: `/api/*`, V2: `/api/v2/*` (V2 sends events via RabbitMQ channel)
6. **`test.go` at root** — scratch code, not part of main app
7. **No linter** — `go vet ./...` is the only static check
8. **Elastic APM hooks** — wired into Fiber, Redis, Mongo, Postgres. Don't remove without checking dependencies
9. **Model table names** — `UserAccount` → `"users"`, `UserSocialAccount` → `"user_social_account"`, `SmartTvCode` → `"smart_tv_code"`
