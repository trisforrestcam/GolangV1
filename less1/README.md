# Lesson 1: Hello World & Basic Syntax

## 🎯 Mục tiêu
Nắm vững **nền tảng vững chắc** của Go - không chỉ là syntax mà là cách Go "nghĩ" về code.

## 🧠 Mental Model - Góc nhìn Expert

### Go là ngôn ngữ "Explicit over Implicit"
- Không có magic, không có implicit behavior
- Mọi thứ đều phải rõ ràng: type conversion, error handling, visibility
- Đây là **feature, không phải bug** - code dễ đọc, dễ maintain

### Zero Values - Triết lý "No Null"
```go
var i int     // 0 (không phải undefined)
var s string  // "" (không phải null)
var b bool    // false
```
- Go **không có null/undefined** cho primitive types
- Đây là lý do Go ít bug null pointer hơn NodeJS

### Short Declaration `:=` - Syntactic Sugar có chủ đích
```go
// Chỉ dùng := khi KHAI BÁO biến mới
city := "HCM"           // ✅ Khai báo
// city := "HN"         // ❌ Lỗi: no new variables

city = "HN"             // ✅ Gán lại (không dùng :=)
```

## 📚 Kiến thức cốt lõi

| Khái niệm | Go | NodeJS tương đương |
|-----------|-----|-------------------|
| Khai báo biến | `var x int = 10` | `let x = 10` |
| Short declaration | `x := 10` | `const x = 10` (không hoàn toàn giống) |
| Constants | `const Pi = 3.14` | `const PI = 3.14` |
| Type conversion | `float64(x)` (explicit) | `Number(x)` hoặc implicit |
| Zero value | Tự động | `undefined` / `null` |

### Format Verbs (fmt.Printf)
```go
%v   // Default format
%+v  // With field names
%#v  // Go syntax representation
%T   // Type
%s   // String
%d   // Integer
%f   // Float
%t   // Boolean
%p   // Pointer
```

## ⚠️ Sai lầm NodeJS dev hay mắc

1. **Implicit type conversion**
```go
// ❌ Sai
var x int = 10
var y float64 = x    // Lỗi: cannot use x (type int) as type float64

// ✅ Đúng
var y float64 = float64(x)
```

2. **Không hiểu zero values**
```go
// ❌ Sai
var count int
if count == nil { ... }  // int không thể so sánh với nil

// ✅ Đúng
if count == 0 { ... }
```

3. **Dùng `:=` sai chỗ**
```go
// ❌ Sai
var name string = "John"
name := "Jane"  // Lỗi: no new variables on left side

// ✅ Đúng
name = "Jane"
```

## 🏆 Best Practices (Expert Level)

### 1. Luôn dùng short declaration `:=` khi có thể
```go
// ✅ Tốt
name := "John"
age := 30

// ❌ Không cần thiết
var name string = "John"
var age int = 30
```

### 2. Nhóm variables có liên quan
```go
// ✅ Tốt
var (
    name    string = "John"
    age     int    = 30
    isAdmin bool   = true
)
```

### 3. Constants cho giá trị không đổi
```go
const (
    MaxConnections = 100
    DefaultTimeout = 30 * time.Second
    Version        = "1.0.0"
)
```

### 4. Chỉ dùng `var` khi cần zero value hoặc cần scope package-level
```go
// ✅ Package-level variables
var db *sql.DB

// ✅ Cần zero value cụ thể
var counter int  // Bắt đầu từ 0

// ✅ Trong function, dùng :=
func process() {
    name := getName()  // ✅
}
```

## 🔗 Liên kết
- **Next:** [Lesson 2: Data Types & Operators](../less2/README.md)
- **Docs:** [go-fundamentals/01-go-mental-model-for-nodejs-dev.md](../../docs/go-fundamentals/01-go-mental-model-for-nodejs-dev.md)

---
*"Go là ngôn ngữ của sự rõ ràng. Nếu bạn thấy phiền vì phải explicit type conversion, bạn chưa hiểu triết lý của Go."*
