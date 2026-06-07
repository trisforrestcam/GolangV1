# Lesson 4: Structs, Methods, Composition

## 🎯 Mục tiêu
Hiểu sâu **Data Modeling** trong Go - structs, methods, composition. Đây là thay thế cho Classes trong OOP.

## 🧠 Mental Model - Góc nhìn Expert

### Go KHÔNG có Classes - Và đó là điều tốt
- Không có inheritance (extends)
- Không có constructors
- Không có access modifiers (public/private)
- Thay vào đó: **Structs + Methods + Composition**

### Composition over Inheritance
```go
type Employee struct {
    User                    // Embedded struct (composition)
    Department string
    Salary     float64
}
```
- Employee "has a" User (composition)
- Employee "is a" User trong behavior (method promotion)
- **Không** phải "extends" hay "is a" trong type hierarchy

### Visibility - Quy tắc đơn giản
```go
type User struct {
    Name  string  // Public (viết hoa)
    email string  // Private (viết thường) - chỉ trong package
}
```
- **Viết hoa = Public**, **viết thường = Private**
- Không có `public`, `private`, `protected`
- Visibility dựa trên **package**, không phải class

### Value vs Pointer Receiver
```go
// Value Receiver - Không modify original
func (u User) Display() string { ... }

// Pointer Receiver - Có thể modify original
func (u *User) UpdateEmail(email string) { ... }
```
- **Rule of thumb:**
  - Methods chỉ đọc → Value receiver
  - Methods modify state → Pointer receiver
  - Struct lớn → Pointer receiver (tránh copy)

## 📚 Kiến thức cốt lõi

### Constructor Pattern
```go
// Go không có constructors - dùng factory functions
func NewUser(id uint, name, email string, age int) *User {
    return &User{
        ID:        id,
        Name:      name,
        Email:     email,
        Age:       age,
        CreatedAt: time.Now(),
    }
}
```

### Struct Tags
```go
type User struct {
    ID    int    `json:"id" gorm:"primaryKey"`
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"email"`
}
```
- Tags là metadata cho reflection (JSON, DB, validation, v.v.)
- Dùng backticks `` ` ``

### Method Sets
```go
type Printer interface {
    Print()
}

// Value receiver
type Book struct{}
func (b Book) Print() {}    // Cả *Book và Book đều thỏa mãn Printer

// Pointer receiver  
type Magazine struct{}
func (m *Magazine) Print() {}  // Chỉ *Magazine thỏa mãn Printer
```

## ⚠️ Sai lầm NodeJS dev hay mắc

1. **Tưởng struct = class**
```go
// ❌ Sai - Không có constructors
user := new(User)  // Có thể dùng nhưng zero values

// ✅ Đúng - Factory function
user := NewUser(1, "John", "john@example.com", 30)
```

2. **Nhầm lẫn pointer vs value**
```go
// ❌ Không hiệu quả với struct lớn
func (u User) Process() { ... }  // Copy toàn bộ struct!

// ✅ Tốt hơn
func (u *User) Process() { ... }  // Chỉ copy pointer
```

3. **Không hiểu method promotion**
```go
type Employee struct {
    User  // Embedded
}

// ✅ Có thể gọi trực tiếp
emp.Name  // Thay vì emp.User.Name

// ❌ Nhưng type không phải "is a"
var u User = emp  // ❌ Lỗi: cannot use emp (type Employee) as type User
```

## 🏆 Best Practices (Expert Level)

### 1. Constructor với validation
```go
func NewUser(name, email string, age int) (*User, error) {
    if name == "" {
        return nil, errors.New("name is required")
    }
    if age < 0 || age > 150 {
        return nil, errors.New("invalid age")
    }
    return &User{
        Name:      name,
        Email:     email,
        Age:       age,
        CreatedAt: time.Now(),
    }, nil
}
```

### 2. Dùng pointer receiver cho consistency
```go
// ✅ Tốt - Tất cả methods cùng receiver type
type Counter struct {
    count int
}

func (c *Counter) Increment() { c.count++ }
func (c *Counter) Decrement() { c.count-- }
func (c *Counter) Value() int { return c.count }
```

### 3. Embedding cho behavior reuse
```go
// ✅ Expert pattern - Mixin-like
type Logger struct {
    mu     sync.Mutex
    output io.Writer
}

func (l *Logger) Log(msg string) { ... }

type Service struct {
    Logger  // Embed behavior
    config Config
}

// Service có method Log() từ Logger
svc.Log("starting...")
```

### 4. Struct tags chuẩn hóa
```go
type User struct {
    ID        int       `json:"id" db:"id" validate:"required"`
    Name      string    `json:"name" db:"name" validate:"required,min=2,max=100"`
    Email     string    `json:"email" db:"email" validate:"required,email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

### 5. Avoid nil pointer
```go
// ✅ Check nil trước khi dùng
func (u *User) Display() string {
    if u == nil {
        return "<nil user>"
    }
    return u.Name
}
```

## 🔗 Liên kết
- **Previous:** [Lesson 3: Functions, Closures, Defer](../less3/README.md)
- **Next:** [Lesson 5: Goroutines, WaitGroup, Mutex](../less5/README.md)
- **Docs:** [go-fundamentals/03-struct-interface-error-handling.md](../../docs/go-fundamentals/03-struct-interface-error-handling.md)

---
*"Go không thiếu OOP - Go có OOP đúng nghĩa: composition over inheritance. Nếu bạn nhớ class inheritance, hãy quên nó đi."*
