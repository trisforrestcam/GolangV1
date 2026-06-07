# Lesson 10: Interfaces, Polymorphism, Generics

## 🎯 Mục tiêu
Master **Interfaces** - feature quan trọng nhất của Go. Hiểu implicit implementation, composition, và generics (Go 1.18+).

## 🧠 Mental Model - Góc nhìn Expert

### Interfaces là "Contracts" - Không phải "Inheritance"
```go
type Writer interface {
    Write([]byte) (int, error)
}

// io.Writer là interface phổ biến nhất
// File, Buffer, Network connection - tất cả đều implement Writer
```
- Interface định nghĩa **behavior**, không phải **data**
- Không có `implements` keyword - **implicit implementation**

### Implicit Implementation = Decoupling
```go
// Go: Không cần khai báo "implements"
type File struct{}
func (f File) Write(p []byte) (int, error) { ... }
// File TỰ ĐỘNG implement io.Writer

// Java/C#: Phải explicit
class File implements Writer { ... }
```
- **Lợi ích:** Decoupling hoàn toàn - package A định nghĩa interface, package B implement
- Không cần import interface definition

### Empty Interface = Any Type
```go
var anything interface{} = 42
anything = "hello"
anything = []int{1, 2, 3}

// Go 1.18+
var anything any = 42
```
- `interface{}` hoặc `any` có thể hold bất kỳ type nào
- Dùng **type assertion** hoặc **type switch** để lấy giá trị

### Interface Composition
```go
type Reader interface {
    Read([]byte) (int, error)
}

type Writer interface {
    Write([]byte) (int, error)
}

type ReadWriter interface {
    Reader  // Embed interface
    Writer  // Embed interface
}
```

## 📚 Kiến thức cốt lõi

### Type Assertion
```go
var val interface{} = "hello"

// Safe assertion
str, ok := val.(string)
if ok {
    fmt.Println(str)
}

// Unsafe assertion (panic nếu sai)
str := val.(string)  // Panic nếu val không phải string
```

### Type Switch
```go
switch v := val.(type) {
case int:
    fmt.Printf("Integer: %d\n", v)
case string:
    fmt.Printf("String: %s\n", v)
case bool:
    fmt.Printf("Boolean: %v\n", v)
default:
    fmt.Printf("Unknown: %T\n", v)
}
```

### Generics (Go 1.18+)
```go
// Generic function
func Find[T any](slice []T, predicate func(T) bool) T {
    for _, item := range slice {
        if predicate(item) {
            return item
        }
    }
    var zero T
    return zero
}

// Generic type constraint
func Min[T comparable](a, b T) T {
    if a < b {  // ❌ Lỗi: comparable không support <
        return a
    }
    return b
}

// Dùng constraints package
import "golang.org/x/exp/constraints"

func Min[T constraints.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}
```

### Nil Interface vs Nil Pointer
```go
// ❌ Trap!
var p *MyStruct = nil
var i MyInterface = p

fmt.Println(i == nil)  // false! i chứa (*MyStruct)(nil), không phải nil interface

// ✅ Check nil pointer
if p == nil {
    // ...
}
```

## ⚠️ Sai lầm NodeJS dev hay mắc

1. **Tưởng interface = class**
```go
// ❌ Sai - Interface không có data
type Animal interface {
    Name string  // ❌ Không thể có field!
    Speak() string
}

// ✅ Đúng - Interface chỉ có methods
type Animal interface {
    Speak() string
}
```

2. **Type assertion không check**
```go
// ❌ Panic nếu sai type
str := val.(string)

// ✅ Safe
str, ok := val.(string)
if !ok {
    return errors.New("not a string")
}
```

3. **Overuse empty interface**
```go
// ❌ Mất type safety
func process(data interface{}) {
    // ...
}

// ✅ Dùng generics hoặc specific interface
func process(data io.Reader) {
    // ...
}
```

## 🏆 Best Practices (Expert Level)

### 1. Interface nhỏ, focused
```go
// ✅ Tốt - Small interface
type Reader interface {
    Read([]byte) (int, error)
}

type Writer interface {
    Write([]byte) (int, error)
}

// ❌ Không tốt - Fat interface
type ReadWriteCloser interface {
    Read([]byte) (int, error)
    Write([]byte) (int, error)
    Close() error
}
```

### 2. Return interfaces, accept structs (hoặc ngược lại?)
```go
// ✅ Accept interfaces (flexible)
func process(r io.Reader) { ... }

// ✅ Return structs (concrete)
func newReader() *bufio.Reader { ... }
```
- **Accept interfaces, return structs** là general rule
- Nhưng đôi khi return interface cũng OK nếu cần hide implementation

### 3. Interface embedding
```go
// ✅ Tốt - Compose interfaces
type ReadWriter interface {
    io.Reader
    io.Writer
}
```

### 4. Generics cho reusable code
```go
// ✅ Generic helper functions
func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
    var result []T
    for _, v := range slice {
        if predicate(v) {
            result = append(result, v)
        }
    }
    return result
}
```

### 5. Avoid `interface{}` nếu có thể
```go
// ❌ Mất type safety
func process(data interface{}) { ... }

// ✅ Dùng generics
func process[T any](data T) { ... }

// ✅ Hoặc specific interface
func process(data io.Reader) { ... }
```

## 🔗 Liên kết
- **Previous:** [Lesson 9: JSON Handling](../less9/README.md)
- **Next:** [Lesson 11: Error Handling Patterns](../less11/README.md)
- **Docs:** [go-fundamentals/03-struct-interface-error-handling.md](../../docs/go-fundamentals/03-struct-interface-error-handling.md)

---
*"Interfaces trong Go là implicit - điều này tạo ra decoupling tự nhiên. Bạn không cần 'implements' keyword, không cần mock framework phức tạp. Interface nhỏ, focused là key của Go design."*
