# Lesson 2: Data Types & Operators

## 🎯 Mục tiêu
Hiểu sâu **type system** của Go - đây là nền tảng để viết code Go hiệu quả và tránh memory issues.

## 🧠 Mental Model - Góc nhìn Expert

### Go là Statically Typed với Type Inference
- Mọi biến đều có type cố định tại compile time
- Không có `any` mặc định (trừ khi bạn explicit dùng `interface{}`)
- Điều này giúp compiler tối ưu memory layout và phát hiện lỗi sớm

### Integer Types - Kiểm soát Memory
```go
int8   // -128 to 127        (1 byte)
int16  // -32768 to 32767    (2 bytes)
int32  // ~-2B to ~2B        (4 bytes)
int64  // rất lớn             (8 bytes)
int    // Platform dependent (32/64 bit)
```
- **Expert tip:** Luôn chọn type nhỏ nhất đủ dùng để tiết kiệm memory trong structs lớn

### Strings là Immutable & UTF-8
```go
s := "Hello"
// s[0] = 'h'  // ❌ Lỗi: cannot assign to s[0]
```
- Strings trong Go là **read-only slices of bytes**
- Index truy cập byte, không phải rune (character)

### Type Safety là Feature
```go
type Celsius float64
type Fahrenheit float64

var c Celsius = 100
var f Fahrenheit = c  // ❌ Lỗi: cannot use c (type Celsius) as type Fahrenheit
```
- Go coi các type aliases là **types hoàn toàn khác nhau**
- Điều này ngăn bug đơn vị trong systems lớn

## 📚 Kiến thức cốt lõi

### Numeric Types
```go
// Integer
var i int = 42          // Platform dependent
var i8 int8 = 127       // Signed 8-bit
var u8 uint8 = 255      // Unsigned 8-bit (byte)

// Float
var f32 float32 = 3.14  // 32-bit (ít dùng)
var f64 float64 = 3.14  // 64-bit (default, nên dùng)

// Complex (hiếm dùng)
var c complex64 = 1 + 2i
```

### Bitwise Operators
```go
m & n    // AND
m | n     // OR
m ^ n     // XOR
^m        // NOT
m << 1   // Left shift
m >> 1   // Right shift
```
- **Expert use case:** Flags, permissions, bitmasks, protocol parsing

### String Operations
```go
len(s)        // Byte length (không phải character count!)
s[0]          // Byte đầu tiên (byte, không phải rune)
s[0:5]        // Slice (bytes 0-4)
s + " world"  // Concatenation
```

## ⚠️ Sai lầm NodeJS dev hay mắc

1. **Hiểu nhầm string length**
```go
// ❌ Sai
s := "Xin chào"
fmt.Println(len(s))  // 11 (bytes), không phải 8 (characters)

// ✅ Đúng
fmt.Println(utf8.RuneCountInString(s))  // 8 characters
```

2. **Integer division surprise**
```go
// ❌ Surprise
result := 5 / 2      // result = 2 (integer division!)

// ✅ Đúng
result := 5.0 / 2.0  // result = 2.5
// hoặc
result := float64(5) / float64(2)
```

3. **Không hiểu zero values**
```go
var i int     // 0
var s string  // ""
var b bool    // false
var p *int    // nil

// ❌ Sai
if s == nil { ... }  // string không thể nil

// ✅ Đúng
if s == "" { ... }
```

## 🏆 Best Practices (Expert Level)

### 1. Luôn dùng `int` trừ khi có lý do cụ thể
```go
// ✅ Tốt - 99% trường hợp
count := 10

// ✅ Chỉ khi cần optimize memory hoặc protocol constraints
var flags uint8
var port uint16
```

### 2. Dùng `float64` thay vì `float32`
```go
// ✅ Tốt
var price float64 = 99.99

// ❌ Tránh trừ khi cần tiết kiệm memory (graphics, embedded)
var price float32 = 99.99
```

### 3. Raw strings cho multi-line
```go
// ✅ Tốt
query := `SELECT id, name, email
          FROM users
          WHERE active = true`

// ❌ Khó đọc
query := "SELECT id, name, email\n" +
         "FROM users\n" +
         "WHERE active = true"
```

### 4. Type aliases cho domain concepts
```go
// ✅ Expert pattern
type UserID int64
type Email string
type Money int64 // cents, tránh float cho tiền

func processPayment(userID UserID, amount Money) error {
    // Không thể nhầm lẫn thứ tự arguments
}
```

### 5. Dùng `rune` cho Unicode
```go
// ✅ Đúng
s := "Xin chào"
for _, r := range s {  // range trả về rune (Unicode code point)
    fmt.Printf("%c ", r)
}
```

## 🔗 Liên kết
- **Previous:** [Lesson 1: Hello World & Basic Syntax](../less1/README.md)
- **Next:** [Lesson 3: Functions, Closures, Defer](../less3/README.md)
- **Docs:** [go-fundamentals/02-go-syntax-cheatsheet.md](../../docs/go-fundamentals/02-go-syntax-cheatsheet.md)

---
*"Trong Go, type không chỉ là annotation - nó là contract. Hiểu type system là hiểu 50% triết lý của Go."*
