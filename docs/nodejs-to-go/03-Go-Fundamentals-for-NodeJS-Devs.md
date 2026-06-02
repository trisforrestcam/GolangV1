# 🎯 Go Fundamentals cho NodeJS Developers

> **Dành cho:** NodeJS dev 4 năm kinh nghiệm muốn nắm vững Go basics
> **Mục tiêu:** Hiểu Go types, functions, structs, interfaces qua ví dụ thực tế từ project

---

## 📊 1. Go Type System Crash Course

### Basic Types

```go
// String
name := "John"                    // string (type inferred)
var name string = "John"          // explicit type
var empty string                  // zero value: ""

// Numbers
age := 30                         // int
var count int = 42                // explicit int
price := 99.99                    // float64
var ratio float32 = 0.5           // explicit float32

// Boolean
isActive := true                  // bool
var verified bool                 // zero value: false

// Byte (alias for uint8)
var b byte = 'A'
```

**NodeJS Comparison:**
```javascript
// JavaScript has only number (float64)
const age = 30;          // number
const price = 99.99;     // number (same type!)

// Go distinguishes int, int32, int64, float32, float64
```

---

### Composite Types

#### Arrays (Fixed size - rarely used)

```go
// Fixed size array
var arr [3]int = [3]int{1, 2, 3}
arr := [3]int{1, 2, 3}
arr := [...]int{1, 2, 3}  // Size inferred

// Access
first := arr[0]  // 1
length := len(arr)  // 3
```

#### Slices (Dynamic arrays - most common)

```go
// Create slice
nums := []int{1, 2, 3}
nums := make([]int, 5)         // [0, 0, 0, 0, 0]
nums := make([]int, 0, 10)    // Empty, capacity 10

// Operations
nums = append(nums, 4)         // Add element
nums = append(nums, 5, 6, 7)  // Add multiple
nums = append(nums[:1], nums[2:]...)  // Remove at index 1

// Sub-slice (shares memory!)
sub := nums[1:3]               // Elements at index 1, 2

// Length and capacity
len(nums)  // Number of elements
cap(nums)  // Total capacity
```

**NodeJS Comparison:**
```javascript
const arr = [1, 2, 3];
arr.push(4);                   // Add
arr.splice(1, 1);             // Remove at index
const sub = arr.slice(1, 3);  // Sub-array (copy)

// Go slices share memory, JS slice creates copy
```

#### Maps (Key-value pairs)

```go
// Create map
users := map[string]int{
    "john": 30,
    "jane": 25,
}

// Or using make
users := make(map[string]int)
users["john"] = 30

// Operations
age := users["john"]         // Get value
age, ok := users["john"]     // Check existence (ok = false if not found)
users["bob"] = 35            // Set value
delete(users, "john")        // Delete key

// Iterate
for name, age := range users {
    fmt.Println(name, age)
}
```

**NodeJS Comparison:**
```javascript
const users = { john: 30, jane: 25 };
users.bob = 35;              // Set
delete users.john;           // Delete
const age = users.john;      // Get (undefined if not found)

// Go map returns zero value if key not found
// JavaScript returns undefined
```

---

### Custom Types (from project)

```go
// Type alias
type UserGender string

const (
    Male   UserGender = "MALE"
    Female UserGender = "FEMALE"
    Other  UserGender = "OTHER"
)

// Constructor function
func NewUserGender(key string) UserGender {
    if key == "male" {
        return Male
    }
    if key == "female" {
        return Female
    }
    return Other
}

// Usage
gender := NewUserGender("male")  // UserGender("MALE")
```

---

## 📝 2. Functions in Go

### Basic Function

```go
// Simple function
func greet(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}

// With multiple parameters
func add(a, b int) int {
    return a + b
}

// Different types
func createUser(name string, age int, active bool) {
    // ...
}
```

### Multiple Return Values (Very Common!)

```go
// Return value + error (standard Go pattern)
func GetUser(id uint) (*UserAccount, error) {
    user := &UserAccount{}
    result := db.Where("id = ?", id).First(user)
    if result.Error != nil {
        return nil, result.Error
    }
    return user, nil
}

// Usage
user, err := GetUser(123)
if err != nil {
    log.Error(err)
    return
}
fmt.Println(user.Name)
```

**NodeJS Comparison:**
```javascript
// JavaScript typically uses try/catch or callbacks
async function getUser(id) {
    const user = await db.findById(id);
    if (!user) throw new Error('Not found');
    return user;
}

// Go returns error as second value
// JavaScript throws exceptions
```

### Named Return Values

```go
// Named returns (documenting what's returned)
func divide(a, b float64) (result float64, err error) {
    if b == 0 {
        err = errors.New("division by zero")
        return  // Returns zero values: result=0, err=error
    }
    result = a / b
    return  // Returns result and nil error
}
```

### Variadic Functions

```go
// Accept variable number of arguments
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

// Usage
sum(1, 2, 3)        // 6
sum(1, 2, 3, 4, 5)  // 15

// Pass slice to variadic
nums := []int{1, 2, 3}
sum(nums...)         // Unpack slice
```

### Anonymous Functions / Closures

```go
// Anonymous function
greet := func(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}
fmt.Println(greet("John"))

// Immediately invoked
result := func(x int) int {
    return x * x
}(5)  // 25

// Closure (captures variables)
counter := 0
increment := func() int {
    counter++
    return counter
}
increment()  // 1
increment()  // 2
```

### Function as Types

```go
// Define function type
type HandlerFunc func(*fiber.Ctx) error

// Function that accepts function
func applyMiddleware(handler HandlerFunc, middlewares ...HandlerFunc) HandlerFunc {
    for _, mw := range middlewares {
        handler = mw(handler)
    }
    return handler
}

// Usage in project (Fiber pattern)
func Login(sql *gorm.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Handler logic
    }
}
```

### Real Examples from Project

```go
// From app/controllers/login.go
// Controller returns fiber.Handler (function type)
func Login(sql *gorm.DB, redisToken *redis.Client, mongo *mongo.Database) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Parse request
        body := &models.LoginBody{}
        err := c.BodyParser(&body)
        if err != nil {
            return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)
        }
        
        // Call repository
        connect := service.PostgresQueries{Conn: sql.WithContext(c.Context())}
        userEntity, err := connect.LoginAccountDb(c.Context(), body)
        
        // Return response
        return utils.HTTPSuccessResponse(c, fiber.StatusOK, map[string]interface{}{
            "access_token": token,
        })
    }
}
```

---

## 🏗️ 3. Structs and Methods

### Struct Definition

```go
// Basic struct (like TypeScript interface + class)
type UserAccount struct {
    Id           uint      `gorm:"primaryKey;autoIncrement:true;unique" json:"id"`
    Name         *string   `gorm:"index:idx_user_name,unique" json:"name"`
    Phone        *string   `gorm:"index:idx_user_phone,unique" json:"phone"`
    Uuid         string    `gorm:"column:uuid" json:"uuid"`
    FullName     string    `gorm:"column:fullname" json:"fullname"`
    PasswordHash string    `gorm:"column:password_hash" json:"password_hash"`
    Email        string    `gorm:"type:varchar(100)" json:"email"`
    Address      string    `gorm:"size:1000" json:"address"`
    Dob          string    `json:"dob"`
    Gender       UserGender `sql:"DEFAULT:MALE" json:"gender"`
    CreatedAt    time.Time  `sql:"DEFAULT:'current_timestamp'" json:"-"`
    UpdatedAt    time.Time  `sql:"DEFAULT:'current_timestamp'" json:"-"`
}
```

**NodeJS Comparison:**
```javascript
// TypeScript interface
interface UserAccount {
    id: number;
    name: string | null;
    phone: string | null;
    uuid: string;
    fullname: string;
    password_hash: string;
    email: string;
    // ...
}

// Go struct is both interface AND class
// Go struct tags map to JSON keys and DB columns
```

### Struct Tags

```go
// Tags define metadata for serialization and DB mapping
type User struct {
    Name string `json:"name" gorm:"column:fullname" validate:"required"`
    //              ↓           ↓                      ↓
    //          JSON key    DB column name        Validation rule
}

// Common tag keys:
// json:"field_name"      - JSON serialization
// gorm:"column:name"     - Database column mapping
// validate:"required"    - Validation rules
// bson:"field_name"      - MongoDB BSON
```

### Struct Initialization

```go
// Named fields (preferred)
user := UserAccount{
    Name:  stringPtr("John"),
    Phone: stringPtr("1234567890"),
    Email: "john@example.com",
}

// Positional (order matters, not recommended)
user := UserAccount{"John", "1234567890", "john@example.com"}

// Zero value
user := UserAccount{}  // All fields at zero values

// Pointer to struct
user := &UserAccount{Name: stringPtr("John")}

// Helper for pointer values
func stringPtr(s string) *string {
    return &s
}
```

### Pointer vs Value Receivers

```go
// Value receiver - works on copy
func (u UserAccount) GetName() string {
    return u.FullName  // Returns copy's field
}

// Pointer receiver - modifies original
func (u *UserAccount) SetName(name string) {
    u.FullName = name  // Modifies original struct
}

// When to use which:
// - Use pointer receiver when you need to modify the struct
// - Use pointer receiver for large structs (avoid copying)
// - Use value receiver for small, immutable operations
```

**Real Example from Project:**
```go
// From app/models/user_entity.go

// Value receiver - just reads data
func (user *UserAccount) MappingUserAccount() *UserAccountMap {
    userMap := &UserAccountMap{}
    userMap.Id = user.Id
    userMap.Name = user.Name
    // ... copy fields
    return userMap
}

// Pointer receiver - modifies user
func (user *UserAccount) HashPassword(password string) error {
    var crypter = pbkdf2.SHA256Crypter
    pass, err := crypter.Hash(password)
    if err != nil {
        return err
    }
    user.PasswordHash = pass  // Modifies original
    return nil
}
```

### Custom Table Names

```go
// GORM uses struct name as table name by default
// Override with TableName method

func (UserAccount) TableName() string {
    return "users"
}

func (UserSocialAccount) TableName() string {
    return "user_social_account"
}

func (SmartTvCode) TableName() string {
    return "smart_tv_code"
}
```

---

## 🔌 4. Interfaces

### Implicit Interface Implementation

```go
// Define interface
type Speaker interface {
    Speak() string
}

// Any type with Speak() method satisfies Speaker
type Dog struct{}
func (d Dog) Speak() string { return "Woof!" }

type Cat struct{}
func (c Cat) Speak() string { return "Meow!" }

// Both Dog and Cat satisfy Speaker implicitly
func makeSound(s Speaker) string {
    return s.Speak()
}

// Usage
makeSound(Dog{})  // "Woof!"
makeSound(Cat{})  // "Meow!"
```

**NodeJS Comparison:**
```javascript
// TypeScript requires explicit implements
class Dog implements Speaker {
    speak() { return "Woof!"; }
}

// Go doesn't need implements keyword
// If it has the methods, it satisfies the interface
```

### Common Interfaces in Go

```go
// error interface (built-in)
type error interface {
    Error() string
}

// Any type with Error() method is an error
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// io.Reader
type Reader interface {
    Read(p []byte) (n int, err error)
}

// io.Writer
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

### Small Interfaces (Go Idiom)

```go
// Go prefers small interfaces (1-3 methods)
type UserRepository interface {
    GetUser(ctx context.Context, id uint) (*UserAccount, error)
    CreateUser(ctx context.Context, user *UserAccount) error
}

// Not like Java/C# with huge interfaces
// Go proverb: "The bigger the interface, the weaker the abstraction"
```

### Type Assertions

```go
// Check underlying type
var s Speaker = Dog{}

dog, ok := s.(Dog)
if ok {
    fmt.Println("It's a dog!", dog)
}

// Type switch
switch v := s.(type) {
case Dog:
    fmt.Println("Dog:", v)
case Cat:
    fmt.Println("Cat:", v)
default:
    fmt.Println("Unknown")
}
```

---

## ❌ 5. Error Handling Deep Dive

### Error as Value Pattern

```go
// Errors are values, not exceptions
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Check immediately
result, err := divide(10, 0)
if err != nil {
    fmt.Println("Error:", err)
    return
}
fmt.Println("Result:", result)
```

**NodeJS Comparison:**
```javascript
// JavaScript throws exceptions
function divide(a, b) {
    if (b === 0) throw new Error("division by zero");
    return a / b;
}

try {
    const result = divide(10, 0);
} catch (err) {
    console.log("Error:", err);
}
```

### Custom Error Types

```go
// Define custom error
type AppError struct {
    Code    int
    Message string
    Err     error  // Wrapped error
}

func (e *AppError) Error() string {
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Constructor
func NewAppError(code int, message string, err error) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Err:     err,
    }
}

// Usage
return NewAppError(404, "user not found", nil)
```

### Error Wrapping (fmt.Errorf with %w)

```go
// Wrap errors to preserve chain
func GetUser(id uint) (*UserAccount, error) {
    user, err := db.FindById(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }
    return user, nil
}

// Check wrapped errors
user, err := GetUser(123)
if err != nil {
    var notFound *NotFoundError
    if errors.As(err, &notFound) {
        // Handle not found
    }
}
```

### Panic and Recover

```go
// Panic (like throw in JS, but for unrecoverable errors)
func riskyOperation() {
    panic("something went terribly wrong")
}

// Recover (like catch, but only in deferred functions)
func safeOperation() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered from panic: %v", r)
        }
    }()
    
    riskyOperation()
    return nil
}

// Don't use panic for normal error handling!
// Use error returns instead
```

### Real Error Handling Patterns from Project

```go
// From app/controllers/login.go
func Login(sql *gorm.DB, redisToken *redis.Client, mongo *mongo.Database) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Parse body - check error immediately
        body := &models.LoginBody{}
        err := c.BodyParser(&body)
        if err != nil {
            log.WithError(err).Error("Something went wrong")
            return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)
        }
        
        // Query database - check error
        connect := service.PostgresQueries{Conn: sql.WithContext(c.Context())}
        userEntity, err = connect.LoginAccountDb(c.Context(), body)
        if err != nil {
            log.WithError(err).Error("Something went wrong")
            return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)
        }
        
        // Check nil (user not found)
        if userEntity == nil && err == nil {
            log.WithError(errors.New(utils.USER_NOT_FOUND)).Error("Something went wrong")
            return utils.HTTPFailResponse(c, fiber.StatusBadRequest, errors.New(utils.USER_NOT_FOUND))
        }
        
        // Success
        return utils.HTTPSuccessResponse(c, fiber.StatusOK, data)
    }
}
```

---

## 📦 6. Working with JSON

### Struct Tags

```go
type LoginBody struct {
    MetaData *MetaData `json:"metadata" validate:"required"`
    
    ProviderName        string `json:"provider_name"`
    ProviderAccessToken string `json:"provider_access_token"`
    ProviderTokenType   string `json:"provider_token_type"`
    ClientId            string `json:"client_id"`
    UserName            string `json:"name"`
    Password            string `json:"password"`
}

// Tag options:
// json:"name"           - JSON field name
// json:"-"              - Skip field
// json:"name,omitempty" - Skip if empty
// validate:"required"   - Validation rule
```

### Marshal/Unmarshal

```go
// Struct to JSON
user := UserAccount{Name: "John", Age: 30}
jsonBytes, err := json.Marshal(user)
if err != nil {
    log.Error(err)
}
fmt.Println(string(jsonBytes))  // {"name":"John","age":30}

// JSON to Struct
jsonStr := `{"name":"John","age":30}`
var user UserAccount
err := json.Unmarshal([]byte(jsonStr), &user)
if err != nil {
    log.Error(err)
}
fmt.Println(user.Name)  // "John"
```

### JSON in HTTP Handlers

```go
// Parse JSON request body (Fiber)
func Login(c *fiber.Ctx) error {
    body := &models.LoginBody{}
    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    // Use body...
}

// Return JSON response
return c.Status(200).JSON(fiber.Map{
    "success": true,
    "data":    user,
})

// Using project's response helpers
return utils.HTTPSuccessResponse(c, fiber.StatusOK, user)
return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)
```

### Real Request/Response Models

```go
// Request model
type LoginBody struct {
    MetaData *MetaData `json:"metadata" validate:"required"`
    UserName string    `json:"name"`
    Password string    `json:"password"`
}

type MetaData struct {
    DeviceId   string `json:"device_id" validate:"required"`
    DeviceOS   string `json:"device_os" validate:"required"`
    DeviceName string `json:"device_name"`
    OSVersion  string `json:"os_version"`
    AppVersion string `json:"app_version"`
}

// Response model
type Response struct {
    Success bool        `json:"success"`
    Error   int         `json:"error_code"`
    Message []string    `json:"error_message"`
    Data    interface{} `json:"data"`
}
```

---

## 🔄 7. Context Package

### What is context.Context?

```go
// Context carries deadlines, cancellation signals, and request-scoped values
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key interface{}) interface{}
}
```

### Why It Matters for HTTP Handlers

```go
// Context ties request lifecycle to operations
func GetUser(c *fiber.Ctx) error {
    // Get context from request
    ctx := c.Context()  // Returns fasthttp.RequestCtx
    
    // Pass to database operations
    user, err := db.WithContext(ctx).FindById(id)
    
    // If client disconnects, ctx is cancelled
    // Database query will be interrupted
}
```

### Cancellation and Timeouts

```go
// Timeout for database queries
func (query *PostgresQueries) LoginAccountDb(ctx context.Context, req *models.LoginBody) (*models.UserAccount, error) {
    user := &models.UserAccount{}
    
    // Create context with 2-second timeout
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()  // Always call cancel to release resources
    
    // Query will fail if takes > 2 seconds
    result := query.Conn.WithContext(ctx).First(&user)
    return user, result.Error
}
```

### Real Usage in Project

```go
// From app/repository/postgres_repo.go
func (query *PostgresQueries) GetDeactivateUser(ctx context.Context, id uint) (*models.UserBlockStatus, error) {
    user := &models.UserBlockStatus{}
    
    // Context with timeout
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    // Prepared statement for performance
    session := query.Conn.Session(&gorm.Session{PrepareStmt: true})
    
    // Query with context
    err := session.Where(models.UserBlockStatus{UserID: &id}).WithContext(ctx).First(&user)
    if err.Error != nil {
        if errors.Is(err.Error, gorm.ErrRecordNotFound) {
            return nil, nil  // Not found, no error
        }
        return nil, errors.New(utils.COMMON_ERR)
    }
    return user, nil
}
```

### Context Best Practices

```go
// 1. Always pass context as first parameter
func ProcessRequest(ctx context.Context, data *Data) error {
    // ...
}

// 2. Use context for cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    // This will stop when ctx is cancelled
    select {
    case <-ctx.Done():
        return
    case data := <-dataCh:
        process(data)
    }
}()

// 3. Don't store context in struct
// ❌ Bad
type Service struct {
    ctx context.Context
}

// ✅ Good
func (s *Service) DoSomething(ctx context.Context) error {
    // ...
}
```

---

## 🎯 Quick Reference: Go Concepts cho NodeJS Dev

| Go Concept | NodeJS Equivalent |
|------------|-------------------|
| `var x int` | `let x: number` |
| `x := 10` | `const x = 10` |
| `func add(a, b int) int` | `function add(a: number, b: number): number` |
| `func f() (int, error)` | `function f(): [number, Error]` |
| `type User struct { ... }` | `interface User { ... }` + `class UserImpl` |
| `func (u *User) Save()` | `class User { save() { ... } }` |
| `interface { Read() }` | `interface Readable { read(): void }` |
| `if err != nil { return err }` | `try { ... } catch (e) { throw e }` |
| `go func() { ... }()` | `setTimeout(() => { ... }, 0)` |
| `ch := make(chan int)` | `new EventEmitter()` |
| `ctx.Done()` | `AbortController.signal` |
| `json.Marshal(v)` | `JSON.stringify(v)` |
| `json.Unmarshal(data, &v)` | `JSON.parse(data)` |

---

## 📚 Tiếp theo

- [[01-Project-Structure-Analysis]] - Phân tích project structure
- [[02-Go-vs-NodeJS-Comparison]] - So sánh chi tiết Go vs NodeJS
- [[04-Practical-Patterns-from-Project]] - Patterns thực tế trong project
- [[05-Quick-Reference-CheatSheet]] - Cheat sheet tra cứu nhanh
