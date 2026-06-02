# Struct, Interface, Error Handling - Hướng Dẫn Cho Node.js Developer

> **Mục tiêu:** Hiểu sâu struct, interface, error handling - ba trụ cột quan trọng nhất trong Go production code. So sánh liên tục với Node.js/TypeScript.

---

## 1. Struct - Nền Tảng Data Model

### 1.1. Khởi tạo Struct

```go
type User struct {
	ID    int
	Email string
	Name  string
	Age   int
}

// Cách 1: Struct literal (phổ biến nhất)
u1 := User{ID: 1, Email: "alice@example.com", Name: "Alice", Age: 30}

// Cách 2: Field-by-field
var u2 User
u2.ID = 2
u2.Email = "bob@example.com"

// Cách 3: new() - trả về pointer, tất cả zero value
u3 := new(User) // &{0 ""  0}
u3.Email = "charlie@example.com"
```

**So sánh Node.js:**
```javascript
// JS - object literal, không type safety
const user = { id: 1, email: "alice@example.com", name: "Alice" };
```

### 1.2. Zero Value - Quan Trọng

| Type | Zero value |
|------|-----------|
| `int`, `float64` | `0`, `0.0` |
| `string` | `""` |
| `bool` | `false` |
| `pointer`, `map`, `slice`, `func`, `interface` | `nil` |
| `struct` | mỗi field là zero value của nó |

```go
var u User
fmt.Println(u) // {0 ""  0} - sẵn sàng dùng, không null/undefined

// Node.js
let user; // undefined - dùng sẽ crash
let user = {}; // {} - user.name -> undefined (không crash nhưng không an toàn)
```

> **Design principle:** Tận dụng zero value hợp lý. Ví dụ: `sync.Mutex{}` sẵn sàng dùng, `bytes.Buffer{}` sẵn sàng ghi.

### 1.3. Value vs Pointer Semantics - CỰC KỲ QUAN TRỌNG

Trong JS, object luôn là reference. Trong Go, **struct là value type** - gán là copy!

```go
a := User{ID: 1, Name: "Alice"}
b := a           // COPY!
b.Name = "Bob"
fmt.Println(a.Name) // "Alice" - không bị thay đổi

// Pointer - trỏ đến cùng data
c := &a
c.Name = "Charlie"
fmt.Println(a.Name) // "Charlie" - BỊ thay đổi
```

```javascript
// Node.js
let a = { id: 1, name: "Alice" };
let b = a;           // REFERENCE
b.name = "Bob";
console.log(a.name); // "Bob" - BỊ thay đổi
```

**Quy tắc chọn:**

| Dùng pointer (`*T`) | Dùng value (`T`) |
|---------------------|-------------------|
| Method cần mutate state | Struct nhỏ, immutable |
| Struct lớn (tránh copy) | Data chỉ đọc |
| Interface cần consistent | Không cần thay đổi |

```go
// Hàm nhận value - chỉ là copy
func printUser(u User) {
	u.Name = "Changed" // chỉ thay đổi copy
}

// Hàm nhận pointer - thay đổi gốc
func updateUser(u *User, name string) {
	u.Name = name
}
```

### 1.4. Embedded Structs - Composition Over Inheritance

Go **không có inheritance**. Dùng **composition**:

```go
type AuditFields struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID    int
	Email string
	AuditFields  // Embedded - không có field name
}

u := User{ID: 1, Email: "alice@example.com"}
u.CreatedAt = time.Now() // Field promotion - truy cập trực tiếp
```

**So sánh TypeScript:**
```typescript
// TS - intersection type
type User = { id: number; email: string } & AuditFields;
// Hoặc class extends
class User extends AuditFields { ... }
```

**Embedded method promotion:**
```go
type Logger struct{ prefix string }
func (l Logger) Log(msg string) { fmt.Printf("[%s] %s\n", l.prefix, msg) }

type Server struct {
	Logger  // Method Log() được promote
	address string
}

s := Server{Logger: Logger{"HTTP"}, address: ":8080"}
s.Log("started") // [HTTP] started
```

**Naming conflict:**
```go
type C struct {
	A  // có Name
	B  // có Name
}
c := C{}
// c.Name = "conflict" // COMPILE ERROR - ambiguous
c.A.Name = "from A" // OK
```

### 1.5. Struct Tags

```go
type User struct {
	ID        int       `json:"id" db:"id" validate:"required"`
	Email     string    `json:"email" db:"email" validate:"required,email"`
	Password  string    `json:"-" db:"password"`          // bỏ qua khi JSON
	Age       int       `json:"age,omitempty" db:"age"`   // bỏ nếu zero value
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GORM tags
type Product struct {
	ID    uint    `gorm:"primaryKey;autoIncrement"`
	Name  string  `gorm:"column:product_name;size:200;not null"`
	Price float64 `gorm:"column:unit_price;type:decimal(10,2)"`
}
```

### 1.6. Anonymous Structs

```go
// One-off data
response := struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}{Status: 200, Message: "OK"}

// Parse JSON nhanh
var result struct {
	Items []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"items"`
}
```

### 1.7. Struct Comparison

```go
type Point struct{ X, Y int }
p1 := Point{1, 2}
p2 := Point{1, 2}
fmt.Println(p1 == p2) // true

// Struct chứa slice/map KHÔNG so sánh được bằng ==
// Phải dùng reflect.DeepEqual()
```

### 1.8. Constructor Pattern (NewXxx)

```go
type Config struct {
	Host     string
	Port     int
	Debug    bool
	LogLevel string
}

func NewConfig(host string, port int) *Config {
	return &Config{
		Host:     host,
		Port:     port,
		Debug:    false,
		LogLevel: "info",
	}
}

cfg := NewConfig("localhost", 8080)
```

**Constructor với validation:**
```go
type Email string

func NewEmail(raw string) (Email, error) {
	if !strings.Contains(raw, "@") {
		return "", fmt.Errorf("invalid email: %s", raw)
	}
	return Email(strings.ToLower(strings.TrimSpace(raw))), nil
}
```

---

## 2. Methods

### 2.1. Value vs Pointer Receiver

```go
type Counter struct{ Value int }

func (c Counter) Get() int    { return c.Value }        // Value receiver
func (c *Counter) Inc()       { c.Value++ }             // Pointer receiver
func (c *Counter) Reset()     { c.Value = 0 }
```

**Lỗi thường gặp:**
```go
func (c Counter) BrokenInc() {
	c.Value++ // chỉ thay đổi copy - BUG!
}
```

### 2.2. Method Sets - Quan Trọng Cho Interface

| Receiver type | Method set |
|---------------|-----------|
| `T` (value) | Chỉ `(t T)` |
| `*T` (pointer) | Cả `(t T)` và `(t *T)` |

```go
type Writer interface {
	Write(data []byte) error
}

type File struct{ name string }
func (f *File) Write(data []byte) error { return nil }

var w Writer = File{"test.txt"}     // COMPILE ERROR!
var w Writer = &File{"test.txt"}    // OK
```

### 2.3. Method Chaining

```go
type QueryBuilder struct {
	table      string
	conditions []string
	orderBy    string
	limit      int
}

func (q *QueryBuilder) Where(cond string) *QueryBuilder {
	q.conditions = append(q.conditions, cond)
	return q
}
func (q *QueryBuilder) OrderBy(field string) *QueryBuilder {
	q.orderBy = field; return q
}
func (q *QueryBuilder) Limit(n int) *QueryBuilder {
	q.limit = n; return q
}

// Fluent API
query := NewQueryBuilder("users").
	Where("age > 18").
	Where("active = true").
	OrderBy("created_at DESC").
	Limit(10).
	Build()
```

---

## 3. Interfaces - Paradigm Shift

### 3.1. Implicit Satisfaction - Không Cần `implements`

```go
type Greeter interface {
	Greet() string
}

type Person struct{ Name string }

// Tự động implement Greeter - không cần keyword!
func (p Person) Greet() string {
	return "Hello, I'm " + p.Name
}

var g Greeter = Person{Name: "Alice"} // OK
```

```typescript
// TypeScript - explicit
interface Greeter { greet(): string; }
class Person implements Greeter { ... }
```

> **Tư duy mới:** Interface define ở **consumer side**, không phải provider side.

### 3.2. Empty Interface `any`

```go
func PrintAnything(v any) {
	fmt.Printf("%v (type: %T)\n", v, v)
}

PrintAnything(42)
PrintAnything("hello")
PrintAnything([]int{1, 2})
```

### 3.3. Type Assertion và Type Switch

```go
func process(v any) {
	// Type assertion an toàn
	str, ok := v.(string)
	if ok { fmt.Println("String:", str) }

	// Type switch - elegant
	switch val := v.(type) {
	case string:
		fmt.Println("String:", val)
	case int:
		fmt.Println("Int:", val)
	default:
		fmt.Printf("Unknown: %T\n", val)
	}
}
```

**Check interface support:**
```go
type Logger interface { Log(msg string) }
type LoggerWithLevels interface {
	Logger
	Debug(msg string)
	Warn(msg string)
}

func setup(logger Logger) {
	logger.Log("started")
	if leveled, ok := logger.(LoggerWithLevels); ok {
		leveled.Debug("debug mode")
	}
}
```

### 3.4. Interface Composition

```go
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
type Closer interface { Close() error }

type ReadWriter interface { Reader; Writer }
type ReadWriteCloser interface { Reader; Writer; Closer }
```

### 3.5. Small Interfaces - Triết Lý Go

```go
// 1 method
type Stringer interface { String() string }
type Reader interface { Read(p []byte) (n int, err error) }
type error interface { Error() string }

// Implement Stringer
type Money struct{ Amount float64; Currency string }
func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.Amount, m.Currency)
}
fmt.Println(Money{99.5, "USD"}) // "99.50 USD"
```

### 3.6. Khi Nào Define Interface

**Anti-pattern - quá sớm:**
```go
type UserRepository interface {
	Create(user *User) error
	FindByID(id int) (*User, error)
	FindByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id int) error
	List(limit, offset int) ([]*User, error)
}
```

**Nên làm - concrete trước, interface ở consumer:**
```go
// Concrete implementation
type UserRepo struct { db *sql.DB }
func (r *UserRepo) FindByID(ctx context.Context, id int) (*User, error) { ... }
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*User, error) { ... }

// Interface nhỏ ở consumer - chỉ method cần
type UserFinder interface {
	FindByID(ctx context.Context, id int) (*User, error)
}

type UserService struct { repo UserFinder }
```

### 3.7. Interface Segregation

```go
// Tách nhỏ theo nhu cầu
type Creator[T any] interface { Create(ctx context.Context, entity T) error }
type Finder[T any] interface { FindByID(ctx context.Context, id int) (T, error) }
type Deleter interface { Delete(ctx context.Context, id int) error }

// Compose khi cần
type UserRepository[T any] interface {
	Creator[T]; Finder[T]; Deleter
}
```

---

## 4. Error Handling - Exhaustive

### 4.1. Error As Value

```javascript
// Node.js
try {
	const user = await getUser(id);
	const posts = await getPosts(user.id);
} catch (err) { console.error(err); }
```

```go
// Go - error là expected flow
user, err := getUser(ctx, id)
if err != nil {
	return fmt.Errorf("get user: %w", err)
}
posts, err := getPosts(ctx, user.ID)
if err != nil {
	return fmt.Errorf("get posts: %w", err)
}
```

### 4.2. `error` Interface

```go
type error interface { Error() string }
```

Bất kỳ type nào implement `Error() string` đều là error:
```go
type ValidationError struct {
	Field   string
	Message string
}
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation '%s': %s", e.Field, e.Message)
}
```

### 4.3. `errors.New` vs `fmt.Errorf`

```go
var ErrNotFound = errors.New("not found")      // sentinel error
err := fmt.Errorf("user %d not found", id)     // formatted
return fmt.Errorf("database query: %w", err)   // wrap error
```

### 4.4. Error Wrapping `%w` (Go 1.13+)

```go
func GetUser(ctx context.Context, id int) (*User, error) {
	user, err := db.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetUser: find user %d: %w", id, err)
	}
	return user, nil
}

// Error chain: HandleRequest: GetUser: find user 42: sql: connection refused
```

### 4.5. `errors.Is` và `errors.As`

```go
var ErrUserNotFound = errors.New("user not found")

user, err := FindUser(42)
if errors.Is(err, ErrUserNotFound) {  // check trong error chain
	fmt.Println("User not found!")
}

// errors.As - extract typed error
var httpErr *HTTPError
if errors.As(err, &httpErr) {
	fmt.Printf("HTTP %d: %s\n", httpErr.StatusCode, httpErr.Message)
}

// Switch nhiều loại error
switch {
case errors.As(err, &httpErr):    // handle HTTP
case errors.As(err, &timeoutErr): // handle timeout
case errors.As(err, &validationErr): // handle validation
default: // unknown
}
```

### 4.6. Sentinel Errors và Custom Error Types

```go
// Sentinel errors
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrUnauthorized  = errors.New("unauthorized")
)

// Custom error type
type ValidationError struct {
	Errors map[string]string
}
func (e *ValidationError) Error() string {
	var msgs []string
	for field, msg := range e.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", field, msg))
	}
	return "validation: " + strings.Join(msgs, "; ")
}
```

### 4.7. Error Patterns

**Check early, return early:**
```go
func ProcessOrder(ctx context.Context, order *Order) error {
	if order == nil { return errors.New("order is nil") }

	user, err := userRepo.FindByID(ctx, order.UserID)
	if err != nil { return fmt.Errorf("find user: %w", err) }

	if !user.IsActive { return fmt.Errorf("user %d inactive", user.ID) }
	// ... continue
	return nil
}
```

**Error accumulation:**
```go
func ValidateBatch(items []Item) error {
	var errs []error
	for i, item := range items {
		if err := validate(item); err != nil {
			errs = append(errs, fmt.Errorf("item[%d]: %w", i, err))
		}
	}
	return errors.Join(errs...) // Go 1.20+
}
```

### 4.8. Panic và Recover

**Chỉ dùng panic khi:**
```go
// 1. Initialization fatal errors
func init() {
	cfg, err := loadConfig()
	if err != nil { panic("cannot load config: " + err.Error()) }
}

// 2. Programmer errors (bugs)
if b == 0 { panic("division by zero") }

// 3. Must pattern
var validEmail = regexp.MustCompile(`^[a-z]+@[a-z]+\.[a-z]+$`)
```

**Recover:**
```go
func SafeHandler(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	fn()
	return nil
}
```

> **Rule:** Panic cho flow control trong business logic = **sai**.

### 4.9. Error Trong Goroutines - Common Pitfall

```go
// SAI - error bị mất
for _, id := range ids {
	go func(id int) {
		err := process(id)
		if err != nil { log.Printf("err: %v", err) }
	}(id)
}

// ĐÚNG - errgroup
g, ctx := errgroup.WithContext(context.Background())
for _, id := range ids {
	id := id
	g.Go(func() error { return processWithContext(ctx, id) })
}
if err := g.Wait(); err != nil {
	return fmt.Errorf("processing: %w", err)
}
```

### 4.10. Try/Catch vs If Err

| Aspect | Node.js try/catch | Go if err != nil |
|--------|------------------|-----------------|
| Flow | Exception - jumps up | Normal return value |
| Performance | Expensive (stack trace) | Cheap |
| Unhandled | Runtime crash | Compile error |
| Propagation | Automatic | Explicit |

```javascript
// Node.js - implicit propagation
const user = await getUser(id); // throws -> caller handles
```

```go
// Go - explicit propagation
user, err := GetUser(ctx, id)
if err != nil { return nil, fmt.Errorf("get user: %w", err) }
```

---

## 5. Generics (Go 1.18+)

### 5.1. Generic Structs

```go
type Stack[T any] struct { items []T }

func (s *Stack[T]) Push(item T) { s.items = append(s.items, item) }
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 { var zero T; return zero, false }
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, true
}

intStack := &Stack[int]{}
intStack.Push(1)
```

### 5.2. Type Constraints

```go
type Number interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64
}

func Sum[T Number](numbers []T) T {
	var total T
	for _, n := range numbers { total += n }
	return total
}
```

### 5.3. Generic Utilities

```go
func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item { return true }
	}
	return false
}

func Map[T any, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice { result[i] = fn(v) }
	return result
}

func Filter[T any](slice []T, fn func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if fn(v) { result = append(result, v) }
	}
	return result
}
```

---

## 6. Design Patterns

### 6.1. Repository Pattern

```go
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id int) (*User, error)
}

type PostgresUserRepo struct{ db *sql.DB }
func NewPostgresUserRepo(db *sql.DB) *PostgresUserRepo {
	return &PostgresUserRepo{db: db}
}

// In-memory cho testing
type InMemoryUserRepo struct{ users map[int]*User }
```

### 6.2. Options Pattern (Functional Options)

```go
type Server struct {
	host         string
	port         int
	readTimeout  time.Duration
	writeTimeout time.Duration
}

type Option func(*Server)

func WithPort(port int) Option         { return func(s *Server) { s.port = port } }
func WithTimeout(r, w time.Duration) Option {
	return func(s *Server) { s.readTimeout = r; s.writeTimeout = w }
}

func NewServer(host string, opts ...Option) *Server {
	s := &Server{host: host, port: 8080, readTimeout: 10 * time.Second}
	for _, opt := range opts { opt(s) }
	return s
}

// Clean, readable
srv := NewServer("localhost",
	WithPort(3000),
	WithTimeout(5*time.Second, 30*time.Second),
)
```

```javascript
// Node.js - plain object
const server = new Server("localhost", { port: 3000, timeout: 5000 });
```

### 6.3. Middleware Pattern

```go
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error
type Middleware func(Handler) Handler

func LoggingMiddleware(next Handler) Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		start := time.Now()
		err := next(ctx, w, r)
		log.Printf("%s %s %dms", r.Method, r.URL.Path, time.Since(start).Milliseconds())
		return err
	}
}

func RecoveryMiddleware(next Handler) Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()
		return next(ctx, w, r)
	}
}

func Chain(h Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

handler := Chain(GetUserHandler, RecoveryMiddleware, LoggingMiddleware, AuthMiddleware)
```

---

## 7. Điều Dễ Sai Từ Node Sang Go

| Sai lầm | Nên làm |
|---------|---------|
| Lạm dụng interface quá sớm | Concrete type trước, interface ở consumer |
| Dùng pointer mọi nơi | Chỉ khi cần mutate hoặc struct lớn |
| Over-engineering | Go ưa simplicity, YAGNI |
| Dùng panic cho business logic | Error return, panic chỉ cho fatal |
| Bỏ qua error check | Luôn `if err != nil` |
| Define interface ở provider | Define ở consumer side |

---

## 8. Checklist

- [ ] Struct zero value hữu ích?
- [ ] Value hay pointer receiver?
- [ ] Interface đủ nhỏ? (1-3 methods)
- [ ] Interface ở consumer side?
- [ ] Error wrap với context?
- [ ] Goroutines handle error đúng?
- [ ] Có cần generics không?

---

> **Tóm lại:** Go ưu tiên **explicit over implicit**, **composition over inheritance**, **errors as values**. Hiểu sâu struct-interface-error là foundation cho mọi Go code production.
