# Practical Patterns from vtvcab-identity-api

> A real-world Go project pattern reference based on the VTVcab Identity API — a Fiber v2 authentication service using Postgres, Redis, MongoDB, and RabbitMQ.

---

## 1. HTTP Handler Pattern (Controllers)

### Structure

Controllers live in `app/controllers/` and use **closure-based handler factories** — functions that accept dependencies and return `fiber.Handler`.

```go
// app/controllers/login.go
func Login(sql *gorm.DB, redisToken *redis.Client, mongo *mongo.Database) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // handler body
    }
}
```

**Why closures?** Dependencies (DB connections, Redis clients) are injected at route registration time, not per-request. The returned `fiber.Handler` captures them via closure.

### Request Parsing

Two patterns are used:

**1. Body parsing with `c.BodyParser`:**
```go
body := &models.LoginBody{}
err := c.BodyParser(&body)
if err != nil {
    return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)
}
```

**2. Query parameter parsing:**
```go
phone := c.Query("phone")           // single param
var searchParam models.SearchParams
err = c.QueryParser(&searchParam)   // struct binding
```

**3. URL path params:**
```go
param := c.Params("user_id")
userId, err := strconv.ParseUint(param, 10, 32)
```

### Validation

Uses `go-playground/validator` via a generic helper:

```go
// settings/utils/validator.go
func ValidateStruct[T any](item T) error {
    validate := validator.New()
    err := validate.Struct(item)
    if err != nil {
        errorItems := err.(validator.ValidationErrors)
        for _, item := range errorItems {
            if item.Tag() == "required" {
                err = fmt.Errorf("Field %s is required", item.Field())
                break
            }
        }
    }
    return err
}
```

Usage in V2 controllers:
```go
errValidate := utils.ValidateStruct[models.LoginBody](*body)
if errValidate != nil {
    return utils.HTTPFailResponse(c, fiber.StatusBadRequest, errValidate)
}
```

Models use `validate` tags:
```go
type LoginBody struct {
    MetaData *MetaData `json:"metadata" validate:"required"`
    UserName string    `json:"name"`
    Password string    `json:"password"`
}
```

### Response Formatting

Standard response envelope (`settings/utils/response.go`):

```go
type Response struct {
    Success bool        `json:"success"`
    Error   int         `json:"error_code"`
    Message []string    `json:"error_message"`
    Data    interface{} `json:"data"`
}
```

**Success response:**
```go
return utils.HTTPSuccessResponse(c, fiber.StatusOK, map[string]interface{}{
    "access_token":  token["access_token"].(string),
    "refresh_token": token["refresh_token"].(string),
    "user_id":       userEntity.Id,
})
```

**Error response:**
```go
return utils.HTTPFailResponse(c, fiber.StatusBadRequest, errors.New(utils.USER_NOT_FOUND))
```

**List response with pagination:**
```go
return utils.HTTPListResponse(c, fiber.StatusOK, listUser, total)
// Returns: { "success": true, "data": [...], "total": 100 }
```

### Error Handling Pattern

Every error path follows this pattern:
1. Log with `logrus.WithFields`
2. Optionally send to APM via `utils.SendErrorApm(c, err)`
3. Return `utils.HTTPFailResponse` with appropriate HTTP status

```go
if err != nil {
    log.WithError(err).Error("Something went wrong")
    utils.SendErrorApm(c, err)
    return utils.HTTPFailResponse(c, fiber.StatusBadRequest, errors.New(utils.COMMON_ERR))
}
```

### V1 vs V2 Controller Pattern

V2 controllers add:
- Input validation via `ValidateStruct`
- Event logging via `defer` + channel:
```go
defer func() { writeLogIfLogin(err, userEntity, chEVLog, body, c) }()
```

---

## 2. Repository Pattern

### Structure

Repositories in `app/repository/` wrap database clients in structs with methods:

```go
// app/repository/postgres_repo.go
type PostgresQueries struct {
    Conn   *gorm.DB
    Client *redis.Client
}

func NewPostgresQueries(conn *gorm.DB) *PostgresQueries {
    return &PostgresQueries{Conn: conn}
}
```

```go
// app/repository/redis_repo.go
type RedisQueries struct {
    Conn *redis.Client
}

func NewRedisConnetion(redis *redis.Client) *RedisQueries {
    return &RedisQueries{Conn: redis}
}
```

### Postgres Repository Methods

All methods follow this pattern:

```go
func (query *PostgresQueries) GetUserByPhone(ctx context.Context, phone string) (*models.UserAccount, error) {
    user := &models.UserAccount{}
    ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
    defer cancel()
    session := query.Conn.Session(&gorm.Session{PrepareStmt: true})
    err := session.Where(models.UserAccount{Phone: &phone}).WithContext(ctx).First(&user)
    if err.Error != nil {
        if errors.Is(err.Error, gorm.ErrRecordNotFound) {
            return nil, nil  // not found = nil, nil
        }
        return nil, errors.New(utils.COMMON_ERR)  // wrap errors
    }
    return user, nil
}
```

Key conventions:
- **Context timeout** on every query (2-15 seconds)
- **Prepared statements** via `gorm.Session{PrepareStmt: true}`
- **nil, nil** for "not found" (not an error)
- **Error wrapping** with user-friendly messages from `utils` constants

### Redis Repository Methods

```go
func (query *RedisQueries) LoginAccountRedisDB(ctx context.Context, req *models.LoginBody) (*models.UserAccount, error) {
    user := &models.UserAccount{}
    key := fmt.Sprintf("user.:%s", req.UserName)
    values, err := query.Conn.Get(ctx, key).Bytes()
    if err != nil && err != redis.Nil {
        return nil, err
    }
    if values == nil {
        return nil, errors.New(utils.USER_NOT_FOUND)
    }
    err = json.Unmarshal([]byte(values), &user)
    // ...
}
```

Redis keys follow a naming convention:
- `user.:<phone_or_name>` — user cache
- `token.<jti>` — JWT token store
- `token.users_token.<user_id>` — set of user's tokens
- `temp_sig_key:<user_id>` — temporary sign keys
- `otp_<phone>` / `otp_<ip>` — OTP rate limiting

### SMS Repository

```go
// app/repository/sms_repo.go
type RedisOTPQueries struct {
    Conn *redis.Client
    Ctx  *fiber.Ctx
}
```

Combines Redis operations with HTTP calls to external SMS service. Uses `init()` to load config once at startup.

### How Repositories Are Injected

Repositories are **not** injected via interfaces — they're instantiated inline in controllers:

```go
// In controller:
connectPos := service.PostgresQueries{Conn: sql.WithContext(c.Context()), Client: redis}
userEntity, err = connectPos.GetUserByPhone(c.Context(), body.UserName)

// Or via constructor:
connect := service.NewPostgresQueries(sql.WithContext(c.Context()))
redisConnect := service.NewRedisConnetion(redis)
```

**Note:** The package is named `identity` but imported as `service` — confusing naming.

---

## 3. Service Layer Pattern

### Structure

Services in `app/service/` contain business logic that orchestrates multiple repositories:

```go
// app/service/deactive_user.go
func HandleNewDeactivateUser(c *fiber.Ctx, postConnect *service.PostgresQueries, 
    redisConnect *service.RedisQueries, userId uint) (*models.UserBlockStatus, error) {
    
    userExist, err := postConnect.GetDeactivateUser(c.Context(), userId)
    if err != nil {
        return nil, err
    }
    if userExist == nil {
        userExist, err = postConnect.AddDeactivateUser(userId)
        if err != nil {
            return nil, err
        }
    } else {
        if userExist.Deactivation != utils.BLOCKED_ACC {
            userExist.Deactivation = utils.TEMP_LOCK_ACC
        }
        statusCheck := postConnect.UpdateDeactivateUser(c.Context(), userExist.UserID, userExist.Deactivation)
        if !statusCheck {
            return nil, errors.New("Failed to update deactivation status")
        }
    }
    return userExist, nil
}
```

### MSSQL Service (External System Integration)

```go
// app/service/create_user_core.go
func CreateUserCore(ctx context.Context, db *gorm.DB, mobile, password string) (*CreateUserCoreResult, error) {
    var result CreateUserCoreResult
    sqlDB, err := db.DB()
    _, err = sqlDB.ExecContext(ctx,
        `EXEC @return_value = [dbo].[usp_Accounts_Create_OTT] ...`,
        sql.Named("return_value", sql.Out{Dest: &result.ReturnValue}),
        // ... more named params
    )
    return &result, nil
}
```

### Device Session Service

```go
// app/service/device_session.go
func KickOutAllDevices(c *fiber.Ctx, mDb *mongo.Database, redisToken *redis.Client, userId *uint) error {
    sessionManager, err := middleware.NewDeviceSessionManager(c.Context(), mDb)
    listDevice, err := sessionManager.GetDeviceSessions(userId, "")
    for _, device := range listDevice {
        tokenManager, err := middleware.NewTokenManager(c.Context(), redisToken)
        if device.Jti != nil {
            err = tokenManager.RevokeToken(*device.Jti, mDb)
            err = service.RevokeTokenPublish(*userId, *device.Jti)
        }
    }
    return nil
}
```

### Key Pattern: Async Cache Updates via Channels

```go
errs := make(chan error)
go redisConnect.CacheUserCient(c.Context(), userEntity, errs)
err = <-errs
if err != nil {
    return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)
}
```

This pattern is used throughout for non-blocking cache writes. The goroutine sends the result (or nil) through the channel.

---

## 4. Middleware Pattern

### JWT Authentication Middleware

```go
// settings/middleware/jwt.go
func VerifyToken(c *fiber.Ctx, redisToken *redis.Client, mongo *mongo.Database) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userClaims, err := ExtractTokenMetadata(c)
        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(utils.ReturnError(err, fiber.StatusUnauthorized))
        }
        tokenManager, err := NewTokenManager(c.Context(), redisToken)
        err = validateToken(c.Context(), userClaims.Jti, userClaims.TokenType, tokenManager, mongo)
        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(utils.ReturnError(err, fiber.StatusUnauthorized))
        }
        return c.Next()  // Continue to next handler
    }
}
```

### Token Extraction

```go
func ExtractTokenMetadata(c *fiber.Ctx) (*UserClaims, error) {
    tokenAuthen, err := ValidateExtractToken(c)  // Extracts "Bearer <token>"
    user, err := ParseTokenClaims(tokenAuthen)   // Parses JWT claims
    return user, nil
}
```

### UserClaims Structure

```go
type UserClaims struct {
    *jwt.StandardClaims
    Id        uint     `json:"id"`
    Uuid      string   `json:"uuid"`
    Name      string   `json:"name"`
    Avatar    string   `json:"avatar"`
    Phone     string   `json:"phone"`
    Sub       uint     `json:"sub"`
    Jti       string   `json:"jti"`
    FullName  string   `json:"fullname"`
    Email     string   `json:"email"`
    TokenType string   `json:"type"`     // "access" or "refresh"
    JtiRf     string   `json:"jti_rf"`   // refresh token JTI
}
```

### Token Manager (Redis-backed)

```go
// settings/middleware/token.go
type TokenManager struct {
    ctx    context.Context
    client *redis.Client
}

type StructToken struct {
    Exp       int    `json:"exp"`
    TokenType int    `json:"type"`
    Parent    string `json:"parent"`
}
```

Token validation walks the parent chain:
```go
func (backend *TokenManager) ValidateToken(jti string) (bool, error) {
    // ... get token from Redis
    if len(strings.TrimSpace(tokenStruct.Parent)) != 0 {
        return backend.ValidateToken(tokenStruct.Parent)  // recursive check
    }
    return true, nil
}
```

### Device Session Manager (MongoDB-backed)

```go
// settings/middleware/device_session.go
type SessionDeviceManager struct {
    ctx               context.Context
    sessionCollection *mongo.Collection
    Conn              *mongo.Database
}
```

Enforces device limits:
```go
func (s *SessionDeviceManager) limitUserActiveSession(userID *uint) error {
    listSession, _ := s.GetSessions(userID)
    if limitDevice <= len(listSession) {
        return app_errors.ERR_OF_LIMIT_DEVICE
    }
    return nil
}
```

### Sentry Error Tracking

```go
// settings/middleware/sentry.go
func SentryMiddlewareError(c *fiber.Ctx) error {
    defer func() {
        if err := recover(); err != nil {
            sentry.CurrentHub().Recover(err)
        }
    }()
    err := c.Next()
    if err != nil {
        sentry.CaptureException(err)
    }
    return err
}
```

### Middleware Chain in Routes

```go
// settings/router/versioned/v1/api.go
api.Post("login", controller.LoginWithRedis(redisLogin, redisToken, pos, mongoDb))
api.Get("my", middleware.VerifyToken(ctx, redisToken, mongoDb), controller.GetMy(redisLogin, pos))
api.Post("/deactive_user", middleware.VerifyToken(ctx, redisToken, mongoDb), 
    controller.DeactivateUser(pos, redisLogin, redisToken, mongoDb))
```

Pattern: `middleware.VerifyToken(...)` is passed as a `fiber.Handler` before the controller handler.

---

## 5. Configuration Pattern

### Viper Config Loading

```go
// settings/config/config.go
func LoadSetting(env string) *Config {
    config := &Config{Viper: viper.New()}
    _, filename, _, _ := runtime.Caller(0)
    folderPatch := path.Join(path.Dir(filename), "../")
    config.AddConfigPath(folderPatch)
    
    switch env {
    case "prod":
        config.SetConfigName("settings.prod")
    case "staging":
        config.SetConfigName("settings.staging")
    default:
        config.SetConfigName("settings.dev")
    }
    config.SetConfigType("toml")
    config.AutomaticEnv()
    config.ReadInConfig()
    return config
}
```

### Environment Selection

```bash
STAGE_STATUS=dev go run .     # loads settings.dev.toml
STAGE_STATUS=prod go run .    # loads settings.prod.toml
```

### Config Struct Embedding

```go
type Config struct {
    *viper.Viper           // Embeds viper for GetString, GetInt, etc.
    fiber        *fiber.Config
    cors         *cors.Config
}
```

### Config Usage Patterns

**In `init()` functions (runs once at startup):**
```go
var signKey []byte

func init() {
    config := config2.LoadSetting(os.Getenv("STAGE_STATUS"))
    signKey = []byte(config.GetString("JWT_SECRET_KEY"))
}
```

**In `main.go`:**
```go
configAppServer := configs.ConfigAppServer()
app.App = fiber.New(*configAppServer.GetFiberConfig())
app.App.Use(cors.New(*configAppServer.GetMidlewareConfig()))
```

### Defaults

Extensive defaults for Fiber and middleware:
```go
config.SetDefault("FIBER_BODYLIMIT", 4194304)
config.SetDefault("FIBER_CONCURRENCY", 262144)
config.SetDefault("MW_FIBER_CORS_ALLOWORIGINS", "*")
```

---

## 6. Database Connection Pattern

### Postgres (GORM + APM)

```go
// platform/database/postgres.go
func PostgresConnection() (*gorm.DB, error) {
    postgresConnURL, _ := utils.ConnectionURLBuilder("postgres")
    url, _ := ParseURL(postgresConnURL)
    
    db, _ := gorm.Open(postgresApm.Open(url), &gorm.Config{
        SkipDefaultTransaction: true,
        PrepareStmt:            true,
        Logger:                 logger.Default.LogMode(logger.Info),
    })
    
    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(5)
    sqlDB.SetMaxOpenConns(maxConnect)     // default 100, configurable via MAXOPENCONN
    sqlDB.SetConnMaxLifetime(time.Hour)
    sqlDB.SetConnMaxIdleTime(time.Minute * 30)
    
    return db, err
}
```

Key: Uses `apmgormv2` driver for Elastic APM instrumentation.

### MongoDB

```go
// platform/database/mongodb.go
func ConnectDatabase() (*MongoInstance, error) {
    clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URL"))
    client, _ := mongo.Connect(context.TODO(), clientOptions, 
        options.Client().SetMonitor(apmmongo.CommandMonitor()))
    
    MI.Client = client
    MI.DB = client.Database("onsports-id")
    MI.Client.Ping(context.TODO(), nil)
    MI.createIndex()  // Creates indexes on session_device collection
    return &MI, nil
}
```

Indexes created:
```go
listIndex := []mongo.IndexModel{
    {Keys: bson.M{"jti": 1}, Options: options.Index().SetUnique(true)},
    {Keys: bson.M{"user_id": 1}},
    {Keys: bson.M{"revoked": 1}},
    {Keys: bson.M{"exp": 1}},
}
```

### Redis (Three Separate Instances!)

```go
// platform/cache/redis.go
func RedisConnection(redisOs string, poolSize int) (*redis.Client, error) {
    opt, _ := redis.ParseURL(redisOs)
    rdb = redis.NewClient(&redis.Options{
        Addr:               opt.Addr,
        Password:           opt.Password,
        DB:                 opt.DB,
        ReadTimeout:        time.Millisecond * 500,
        MinIdleConns:       50,
        PoolSize:           poolSize,
        MaxConnAge:         30 * time.Minute,
        PoolTimeout:        30 * time.Second,
        IdleTimeout:        5 * time.Minute,
        IdleCheckFrequency: 1 * time.Minute,
    })
    rdb.AddHook(apmgoredis.NewHook())  // APM instrumentation
    return rdb, nil
}
```

**Three Redis instances in `main.go`:**
```go
redisStore, _ := cache.RedisConnection(os.Getenv("REDIS_URL"), 250)           // User cache
redisToken, _ := cache.RedisConnection(os.Getenv("REDIS_URL_TOKEN"), 250)     // JWT token store
redisBanChat, _ := cache.RedisConnection(os.Getenv("REDIS_URL_BAN_CHAT"), 150) // Ban list
```

---

## 7. Event/Message Pattern (RabbitMQ)

### Producer

```go
// app/producer/producer.go
type AMQPE struct {
    AMQPExchange string
    AMQPRouting  string
    QUEUEName    string
}

func Publisher(data []byte, info *AMQPE) error {
    conn, ch := GetConnection(info)  // Creates connection per publish!
    defer conn.Close()
    defer ch.Close()
    err = ch.Publish(info.AMQPExchange, info.AMQPRouting, false, false,
        amqp.Publishing{ContentType: "text/plain", Body: data})
    return nil
}

func PushMessage(mes []byte, info *AMQPE) error {
    conn, ch := GetQueueConnection(info)
    defer conn.Close()
    defer ch.Close()
    err := ch.Publish("", info.QUEUEName, false, false,
        amqp.Publishing{ContentType: "text/plain", Body: mes})
    return err
}
```

### Worker/Consumer (Channel-based)

```go
// app/worker/model.go
type EventLog struct {
    CreatedAt  string `json:"created_at"`
    DeviceOS   string `json:"device_os"`
    DeviceName string `json:"device_name"`
    UserID     string `json:"user_id"`
    DeviceID   string `json:"device_id"`
    IP         string `json:"ip"`
    UserName   string `json:"user_name"`
}

// app/worker/event_log_worker.go
func Consume(chEVLog chan EventLog) {
    for {
        select {
        case item, ok := <-chEVLog:
            if ok {
                pushEventLog(&item)
            }
        }
    }
}
```

### Channel Wiring in main.go

```go
chEVLog := make(chan worker.EventLog)
go worker.Consume(chEVLog)  // Background goroutine consuming events
```

### Event Logging in Controllers (V2)

```go
// In controller:
defer func() { writeLogIfLogin(err, userEntity, chEVLog, body, c) }()

// Helper writes to channel:
func writeLogIfLogin(err error, user *models.UserAccount, ch chan worker.EventLog, 
    body *models.LoginBody, c *fiber.Ctx) {
    if err == nil && user != nil {
        ch <- worker.EventLog{
            CreatedAt: time.Now().Format(time.RFC3339),
            UserID:    fmt.Sprintf("%d", user.Id),
            IP:        c.IP(),
            // ...
        }
    }
}
```

---

## 8. OAuth/Social Login Pattern

### Google OAuth

```go
// app/helpers/google.go
func ValidateGoogleToken(token string, tokenType string) (*models.GoogleUser, error) {
    if tokenType == "id_token" {
        return ValidateIdTokenGoogle(token)   // Uses google.golang.org/api/idtoken
    }
    return ValidateAccessTokenGoogle(token)   // Uses OAuth2 service
}
```

Two flows:
- **ID Token:** Validated with `idtoken.Validate()`, checks audience against config
- **Access Token:** Uses `oauth2.Service.Tokeninfo()` then fetches user info

### Facebook OAuth

```go
// app/helpers/facebook.go
func ValidateIdTokenFacebook(token string) (*models.FacebookUser, error) {
    oauth2 := GetFacebookOAuthConfig()
    user, err := HandleExtendFacebookToken(token, oauth2)
    return user, err
}
```

Flow: Short-lived token → Exchange for long-lived token → Fetch user profile from Graph API.

### Apple OAuth

```go
// app/helpers/apple.go
func ValidateIdTokenIos(redisCache *redis.Client, token string) (*models.AppleUser, error) {
    // 1. Decode JWT header and payload (manual, not using jwt library)
    // 2. Validate issuer (apple.com)
    // 3. Validate audience (APPLE_CLIENT_ID)
    // 4. Check expiration
    // 5. Fetch Apple public keys (cached in Redis for 6 hours)
    // 6. Verify RSA signature manually
    // 7. Extract email, sub from claims
}
```

Apple keys are cached in Redis:
```go
cacheKey := "publicKeyAppleCache_" + kid
cachedData, err := redisCache.Get(context.Background(), cacheKey).Result()
// ... cache for 6 hours
```

### Unified OAuth Entry Point

```go
// app/repository/social_repo.go
func (query *PostgresQueries) LoginWithOauth(ctx context.Context, redisCache *redis.Client, 
    payload *models.LoginBody) (*models.UserAccount, error) {
    
    if payload.ProviderName == "google" {
        userSocailGg, _ := social.ValidateGoogleToken(payload.ProviderAccessToken, payload.ProviderTokenType)
        userEnity, _ = query.LoginAccountGoogle(ctx, userSocailGg, payload)
    }
    if payload.ProviderName == "facebook" {
        userSocailFb, _ := social.ValidateIdTokenFacebook(payload.ProviderAccessToken)
        userEnity, _ = query.LoginAccountFacebook(ctx, userSocailFb)
    }
    if payload.ProviderName == "appleid" {
        userSocailApple, _ := social.ValidateIdTokenIos(redisCache, payload.ProviderAccessToken)
        userEnity, _ = query.LoginAccountApple(ctx, userSocailApple, payload)
    }
    return userEnity, nil
}
```

Each provider follows: **Validate token → Find/Create user → Link social account → Return user**.

---

## Summary: Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                          main.go (App)                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────────────┐  │
│  │ Postgres │ │  Redis   │ │ MongoDB  │ │ RabbitMQ Channel  │  │
│  │  (GORM)  │ │ (3 inst) │ │          │ │                   │  │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────────┬──────────┘  │
│       │            │            │                 │             │
│       ▼            ▼            ▼                 ▼             │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Controllers                          │   │
│  │  (Closure factories: func deps... → fiber.Handler)      │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │   │
│  │  │  login.go    │  │  user.go     │  │  V2 handlers │  │   │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  │   │
│  └─────────────────────────┬───────────────────────────────┘   │
│                            │                                    │
│       ┌────────────────────┼────────────────────┐              │
│       ▼                    ▼                    ▼              │
│  ┌──────────┐      ┌──────────────┐      ┌──────────┐         │
│  │ Repository│      │   Service    │      │ Helpers  │         │
│  │ (Postgres)│      │ (Business    │      │ (OAuth)  │         │
│  │ (Redis)   │      │  Logic)      │      │          │         │
│  │ (SMS)     │      │              │      │          │         │
│  └──────────┘      └──────────────┘      └──────────┘         │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Middleware                             │  │
│  │  JWT Auth │ Token Manager │ Device Session │ Sentry      │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Platform                              │  │
│  │  database/ │ cache/ │ (connections + APM hooks)          │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### Key Files to Start With

| File | Why |
|------|-----|
| `main.go` | App bootstrap, dependency wiring |
| `app/controllers/login.go` | Core auth flow, handler patterns |
| `app/repository/postgres_repo.go` | Data access patterns |
| `settings/middleware/jwt.go` | JWT generation and validation |
| `settings/config/config.go` | Configuration loading |
| `settings/utils/response.go` | Response formatting |
| `settings/router/versioned/v1/api.go` | Route registration |
