# Lesson 3: Functions - Nền tảng của Go

## 🎯 Mục tiêu
Master **function design** trong Go - từ basic syntax đến higher-order functions, closures, và defer. Đây là nền tảng cho idiomatic Go.

## 🧠 Mental Model - Góc nhìn Expert

### Functions là First-Class Citizens
- Functions có thể: gán vào biến, truyền làm parameter, return từ function khác
- Điều này tạo ra **composition** thay vì **inheritance**

### Multiple Return Values - Triết lý "Error as Value"
```go
result, err := doSomething()
if err != nil {
    return err
}
```
- Go **KHÔNG** dùng exceptions (try/catch)
- Error là **value bình thường**, không phải control flow đặc biệt
- Điều này buộc developer **phải handle error** (không thể ignore)

### Defer - Resource Management tự động
```go
f, err := os.Open("file.txt")
if err != nil {
    return err
}
defer f.Close()  // ✅ Chạy khi function return
```
- `defer` chạy theo thứ tự **LIFO** (Last In, First Out)
- Đảm bảo cleanup **luôn** xảy ra, kể cả khi panic

### Named Return Values - Cẩn thận khi dùng
```go
func getDimensions() (width, height int) {
    width = 1920   // Gán trực tiếp
    height = 1080
    return         // Naked return - trả về named values
}
```
- **Expert advice:** Chỉ dùng named returns khi cải thiện readability rõ rệt
- Trong functions phức tạp, naked return gây confusion

## 📚 Kiến thức cốt lõi

### Function Signature
```go
func name(param type) returnType       // 1 param
func name(a, b int) int                // Multiple params cùng type
func name() (int, error)               // Multiple returns
func name(a int, b ...string)          // Variadic
```

### Closure - Function nhớ scope
```go
func makeCounter() func() int {
    count := 0
    return func() int {
        count++  // Capture biến count từ outer scope
        return count
    }
}

counter := makeCounter()
counter() // 1
counter() // 2
counter() // 3
```

### Higher-Order Functions
```go
func filter(nums []int, predicate func(int) bool) []int {
    var result []int
    for _, n := range nums {
        if predicate(n) {
            result = append(result, n)
        }
    }
    return result
}
```

### Panic & Recover
```go
func safeOperation() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered: %v", r)
        }
    }()
    // ... code có thể panic
}
```
- `panic` = unrecoverable error (giống uncaught exception)
- `recover` chỉ hoạt động trong `defer`
- **Expert rule:** Chỉ dùng panic cho lỗi **lập trình** (bug), không dùng cho lỗi **runtime** thông thường

## ⚠️ Sai lầm NodeJS dev hay mắc

1. **Ignore errors**
```go
// ❌ Tệ nhất
result, _ := doSomething()  // Ignore error

// ✅ Luôn check
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

2. **Dùng panic như throw**
```go
// ❌ Sai
if user == nil {
    panic("user is nil")  // Không dùng panic cho business logic!
}

// ✅ Đúng
if user == nil {
    return errors.New("user not found")
}
```

3. **Defer trong loop (Performance trap)**
```go
// ❌ Chậm + tốn memory
for _, file := range files {
    f, _ := os.Open(file)
    defer f.Close()  // Defer stack up, không chạy đến cuối function!
}

// ✅ Đúng
for _, file := range files {
    func() {
        f, _ := os.Open(file)
        defer f.Close()
        // ... process
    }()
}
```

## 🏆 Best Practices (Expert Level)

### 1. Error handling: Check ngay lập tức
```go
// ✅ Idiomatic Go
f, err := os.Open("file.txt")
if err != nil {
    return err
}
defer f.Close()
```

### 2. Early return (Guard clause)
```go
// ✅ Tốt
func process(user *User) error {
    if user == nil {
        return errors.New("user is nil")
    }
    if !user.Active {
        return errors.New("user is inactive")
    }
    // ... main logic
}

// ❌ Khó đọc (Pyramid of Doom)
func process(user *User) error {
    if user != nil {
        if user.Active {
            // ... main logic
        }
    }
}
```

### 3. Function size: Một function = một việc
```go
// ✅ Tốt - Mỗi function rõ ràng
func validateUser(u *User) error { ... }
func saveUser(u *User) error { ... }
func notifyUser(u *User) error { ... }

func createUser(u *User) error {
    if err := validateUser(u); err != nil {
        return err
    }
    if err := saveUser(u); err != nil {
        return err
    }
    return notifyUser(u)
}
```

### 4. Dùng closures cho dependency injection
```go
// ✅ Expert pattern
type Handler func(ctx context.Context) error

func withTimeout(h Handler, timeout time.Duration) Handler {
    return func(ctx context.Context) error {
        ctx, cancel := context.WithTimeout(ctx, timeout)
        defer cancel()
        return h(ctx)
    }
}
```

### 5. Defer cho resource cleanup
```go
// ✅ Luôn dùng defer cho cleanup
tx, err := db.Begin()
if err != nil {
    return err
}
defer tx.Rollback()  // Safe: Rollback sau commit là no-op

if err := tx.Exec(...); err != nil {
    return err
}

return tx.Commit()  // Defer Rollback sẽ chạy nhưng là no-op
```

## 🔗 Liên kết
- **Previous:** [Lesson 2: Data Types & Operators](../less2/README.md)
- **Next:** [Lesson 4: Structs, Methods, Composition](../less4/README.md)
- **Docs:** [go-fundamentals/03-struct-interface-error-handling.md](../../docs/go-fundamentals/03-struct-interface-error-handling.md)

---
*"Trong Go, function không chỉ là block code - nó là unit of composition. Master functions là master cách Go tổ chức logic."*
