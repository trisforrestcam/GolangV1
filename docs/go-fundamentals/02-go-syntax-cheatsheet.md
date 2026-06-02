# Go Syntax Cheatsheet - Dành Cho Node.js Developer

> **Mục tiêu:** So sánh chi tiết cú pháp Go với JavaScript/Node.js, giúp developer có 4 năm kinh nghiệm Node.js nắm vững Go một cách trực quan.
>
> **Yêu cầu đọc trước:** Đã đọc [[01-go-overview]] để hiểu tổng quan về Go.

---

## 1. Variables & Types

### 1.1. Khai báo biến: `var`, `:=`, `const`

Go có **3 cách** khai báo biến, mỗi cách có ngữ nghĩa khác nhau:

```go
// var - explicit declaration, có thể dùng ở package level và function level
var name string = "Tris"       // explicit type + value
var age = 28                    // type inference (Go tự suy ra là int)
var score int                   // zero value: 0

// := - short declaration (CHỈ dùng trong function)
height := 1.75                  // type inference: float64
city := "Hanoi"                 // type inference: string

// const - compile-time constant
const Pi = 3.14159
const MaxRetries = 3
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript
> var name = "Tris";       // function-scoped, hoisted (ít dùng)
> let age = 28;            // block-scoped, mutable
> const city = "Hanoi";    // block-scoped, nhưng VẪN mutable nếu là object/array!
> ```
>
> | Go | JavaScript | Khác biệt chính |
> |---|---|---|
> | `var` | `var` | Go's `var` an toàn hơn, không hoisting, không re-declare |
> | `:=` | `let` | Go's `:=` giống `const` hơn vì **phải assign ngay** |
> | `const` | `const` | Go's `const` là **compile-time**, JS's `const` chỉ là "don't reassign" |

**⚡ Lưu ý quan trọng:** Trong Go, `:=` là cách phổ biến nhất để khai báo biến trong function. Nó vừa declare vừa assign, và Go compiler sẽ **ép bạn phải dùng biến đó** (nếu không sẽ báo lỗi: `declared but not used`).

```go
func example() {
    // ❌ Lỗi: declared but not used
    x := 10

    // ✅ Đúng: dùng _ nếu chưa cần
    _ = 10

    // ✅ Đúng: dùng ngay
    y := 20
    fmt.Println(y)
}
```

### 1.2. Type System: Explicit vs Implicit

Go là **statically typed** với **explicit typing**. Mọi biến có kiểu xác định tại compile time:

```go
var a int = 10       // explicit: int
var b = 10           // implicit: Go suy ra là int
c := 10              // implicit: int

var d float64 = 10   // explicit: float64
// var e int = d     // ❌ Lỗi! Không thể gán float64 cho int
var e int = int(d)   // ✅ Phải convert explicitly

// TypeScript có thể làm được:
// let a: number = 10;
// let b = 10;         // TypeScript cũng suy ra là number
// TypeScript cho phép implicit conversion trong 1 số case
```

> **📦 Node.js Comparison:**
>
> | Đặc điểm | Go | JavaScript/TypeScript |
> |---|---|---|
> | Kiểu kiểm tra | Compile-time (static) | Runtime (JS) / Compile-time (TS) |
> | Type inference | Có, nhưng limited | TS inference rất mạnh |
> | Implicit conversion | **KHÔNG BAO GIỜ** | Rất nhiều (`"5" + 1 = "51"`) |
> | Type casting | Phải dùng explicit function | `"5" * 1 = 5` (implicit) |

### 1.3. Zero Values - Khái niệm CỰC KỲ QUAN TRỌNG

Mỗi kiểu dữ liệu trong Go có **zero value** mặc định. Không bao giờ có `undefined` hay `null`:

```go
// Zero values của các kiểu cơ bản
var i int          // 0
var f float64      // 0.0
var b bool         // false
var s string       // "" (empty string, NOT nil)
var p *int         // nil
var sl []int       // nil
var m map[string]int // nil
var ch chan int     // nil
var fn func()      // nil
var iface error    // nil
var arr [3]int     // [0, 0, 0] (array có zero value cho từng phần tử)
```

> **📦 Node.js Comparison:**
>
> ```javascript
> let i;           // undefined (Go: 0)
> let b;           // undefined (Go: false)
> let s;           // undefined (Go: "")
> let obj;         // undefined (Go: nil cho pointers/slices/maps)
> let arr;         // undefined (Go: zero-initialized array)
>
> // JavaScript có HAI giá trị "nothing":
> undefined        // biến chưa được gán
> null             // explicit "no value"
>
> // Go chỉ có MỘT: nil (cho reference types)
> // và zero values (cho value types)
> ```
>
> **Tại sao zero values quan trọng?**
> - Trong JS, bạn phải check `typeof x !== 'undefined'` hoặc `x != null`
> - Trong Go, `int` luôn có giá trị `0` → không bao giờ crash vì "undefined is not a number"
> - Nhưng `nil` slices/maps/pointers vẫn có thể gây panic nếu dùng sai!

**⚠️ Gotcha: Nil vs Empty**

```go
// Nil slice vs Empty slice - KHÁC NHAU!
var nilSlice []int            // nil slice, len=0, cap=0
emptySlice := []int{}         // empty slice, len=0, cap=0
makeSlice := make([]int, 0)   // empty slice, len=0, cap=0

// Cả 3 đều có len=0, nhưng nilSlice == nil, 2 cái kia != nil
fmt.Println(nilSlice == nil)  // true
fmt.Println(emptySlice == nil) // false

// Khi marshal JSON:
jsonBytes, _ := json.Marshal(nilSlice)   // "null"
jsonBytes2, _ := json.Marshal(emptySlice) // "[]"
```

### 1.4. Type Conversion - Phải EXPLICIT

Go **không cho phép** implicit conversion. Bạn phải convert tường minh:

```go
// ❌ Lỗi: cannot convert
var i int = 42
var f float64 = i          // compile error!
var s string = i           // compile error!

// ✅ Đúng: explicit conversion
var f float64 = float64(i) // 42.0
var s string = strconv.Itoa(i) // "42"
var i2 int = int(f)        // 42 (mất phần thập phân)

// String ↔ []byte cũng phải explicit
str := "Hello"
bytes := []byte(str)       // convert string to byte slice
str2 := string(bytes)      // convert back

// Integer types cũng phải convert
var x int32 = 100
var y int64 = int64(x)     // phải convert tường minh
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript implicit conversions (nguồn gốc của nhiều bug!)
> "5" + 1      // "51" (string concatenation)
> "5" - 1      // 4    (numeric subtraction)
> "5" == 5     // true (loose equality)
> "5" === 5    // false (strict equality)
> true + 1     // 2
> false + "1"  // "false1"
>
> // Go NEVER làm điều này
> // Go: "5" + 1 → compile error (không thể concatenate string với int)
> // Bạn PHẢI viết: "5" + strconv.Itoa(1)
> ```

### 1.5. Constants - Compile-time Constants

Go constants khác biệt lớn so với JavaScript's `const`:

```go
// Constants trong Go là compile-time constants
const Pi = 3.14159265358979
const MaxUsers = 1000

// Typed constants
const Greeting string = "Hello"
const Timeout float64 = 30.0

// Untyped constants (linh hoạt hơn)
const Untyped = 42        // có thể dùng ở bất kỳ numeric context nào
var f float64 = Untyped   // OK! Vì Untyped là untyped constant
var i int = Untyped       // OK!

// iota - enum-like pattern
type Direction int
const (
    North Direction = iota  // 0
    East                    // 1
    South                   // 2
    West                    // 3
)

// iota với expression
const (
    _  = iota             // skip 0
    KB = 1 << (10 * iota) // 1 << 10 = 1024
    MB = 1 << (10 * iota) // 1 << 20 = 1048576
    GB = 1 << (10 * iota) // 1 << 30
    TB = 1 << (10 * iota) // 1 << 40
)

// Constants không thể dùng với expression runtime
const x = len("hello")   // ✅ OK: len() của string literal là compile-time
const y = len(someVar)   // ❌ Lỗi: someVar là runtime
```

> **📦 Node.js Comparison:**
>
> | | Go `const` | JS `const` |
> |---|---|---|
> | Kiểu | Compile-time constant | Runtime binding |
> | Object/Array | Không có object const | `const obj = {}` nhưng vẫn mutate được |
> | Performance | Compiler thay thế giá trị trực tiếp | Vẫn là runtime variable |
> | Enum pattern | `iota` rất mạnh | Phải dùng object hoặc TypeScript enum |

### 1.6. Các kiểu dữ liệu cơ bản

```go
// Boolean
var b bool = true       // true hoặc false

// Integer
var i1 int              // platform-dependent: 32 hoặc 64 bit
var i2 int8             // -128 to 127
var i3 int16            // -32768 to 32767
var i4 int32            // ~±2.1 tỷ
var i5 int64            // ~±9.2×10^18
var u1 uint             // unsigned int
var u2 uint8            // = byte (0 to 255)
var u3 uint16
var u4 uint32
var u5 uint64

// Float
var f1 float32          // single precision
var f2 float64          // double precision (default)

// String & Rune
var s string            // UTF-8 encoded, immutable
var r rune              // = int32, represents a Unicode code point
var b byte              // = uint8, represents a single byte

// Complex (ít dùng trong web dev)
var c complex128        // 128-bit complex number
```

---

## 2. Collections Deep Dive

### 2.1. Array - Fixed Size

Array trong Go có **kích thước cố định**, khác hoàn toàn so với JavaScript:

```go
// Khai báo array với size cố định
var nums [3]int = [3]int{1, 2, 3}
names := [3]string{"Alice", "Bob", "Charlie"}

// Size là MỘT PHẦN CỦA KIỂU
var a [3]int
var b [5]int
// a và b có KHÁC kiểu! Không thể gán a = b

// Ellipsis - để Go tự đếm
colors := [...]string{"red", "green", "blue"} // [3]string

// Zero value: mỗi phần tử là zero value của kiểu đó
var zeros [3]int   // [0, 0, 0]
var bools [2]bool  // [false, false]

// Truy cập và modify
nums[0] = 10
fmt.Println(len(nums)) // 3

// Copy array (VALUE COPY - không phải reference!)
original := [3]int{1, 2, 3}
copied := original     // copied là BẢN SAO hoàn toàn
copied[0] = 999
fmt.Println(original[0]) // 1 (không bị thay đổi!)
fmt.Println(copied[0])   // 999
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript Array là DYNAMIC
> const arr = [1, 2, 3];
> arr.push(4);           // [1, 2, 3, 4] - OK
> arr.length = 100;      // có thể thay đổi size
>
> // Go Array là FIXED SIZE
> var arr [3]int = [3]int{1, 2, 3}
> // arr.push(4) → không có push method!
> // arr có luôn length = 3
>
> // JavaScript arrays là reference type
> const a = [1, 2, 3];
> const b = a;           // b là reference đến cùng array
> b[0] = 999;
> console.log(a[0]);     // 999 (bị thay đổi!)
>
> // Go arrays là value type
> a := [3]int{1, 2, 3}
> b := a                 // b là BẢN SAO
> b[0] = 999
> fmt.Println(a[0])      // 1 (không bị thay đổi)
> ```

### 2.2. Slice - Dynamic Array (Cực kỳ quan trọng!)

Slice là collection **phổ biến nhất** trong Go. Hiểu slice internals sẽ giúp bạn viết Go giỏi:

```go
// Tạo slice
nums := []int{1, 2, 3}           // literal
zeros := make([]int, 5)           // [0,0,0,0,0], len=5, cap=5
withCap := make([]int, 3, 10)     // [0,0,0], len=3, cap=10
var nilSlice []int                 // nil, len=0, cap=0

// Thao tác cơ bản
nums = append(nums, 4)             // thêm phần tử
nums = append(nums, 5, 6, 7)      // thêm nhiều
nums = append(nums, otherSlice...) // merge 2 slices

// Length và Capacity
fmt.Println(len(nums))             // số phần tử hiện tại
fmt.Println(cap(nums))             // dung lượng allocated

// Sub-slicing (tạo slice mới TRỎ VÀNG DATA GỐC!)
original := []int{10, 20, 30, 40, 50}
sub := original[1:3]               // [20, 30], len=2, cap=4
sub[0] = 999
fmt.Println(original[1])           // 999! (shared underlying array!)

// Để tạo bản SAO ĐỘC LẬP:
copied := make([]int, len(original))
copy(copied, original)             // deep copy
```

**Slice Internal Structure:**

Một slice trong Go thực chất là **struct 3 fields**:

```
┌─────────────────────────────────────┐
│           Slice Header              │
├─────────────┬───────────┬───────────┤
│  Pointer    │  Length   │  Capacity │
│  (8 bytes)  │  (8 bytes)│  (8 bytes)│
└──────┬──────┴───────────┴───────────┘
       │
       ▼
   ┌───┬───┬───┬───┬───┬───┬───┬───┐
   │ 0 │ 1 │ 2 │ 3 │ 4 │ 5 │ 6 │ 7 │  ← underlying array
   └───┴───┴───┴───┴───┴───┴───┴───┘
```

```go
// Illustration
a := make([]int, 3, 8)  // Pointer→[0,0,0,...], len=3, cap=8
b := a[1:2]             // Pointer→a[1], len=1, cap=7
c := a[:2]              // Pointer→a[0], len=2, cap=8
```

**Append Behavior - Capacity Growth:**

```go
s := make([]int, 0)
fmt.Println(len(s), cap(s)) // 0 0

s = append(s, 1)
fmt.Println(len(s), cap(s)) // 1 1

s = append(s, 2)
fmt.Println(len(s), cap(s)) // 2 2

s = append(s, 3)
fmt.Println(len(s), cap(s)) // 3 4  ← capacity doubled!

s = append(s, 4)
fmt.Println(len(s), cap(s)) // 4 4

s = append(s, 5)
fmt.Println(len(s), cap(s)) // 5 8  ← doubled again!

// Quy tắc: capacity thường gấp đôi khi slice đầy
// Nhưng ĐỪNG phụ thuộc vào implementation detail này!
```

**⚠️ Gotcha: Append với sub-slice**

```go
original := []int{1, 2, 3, 4, 5}
sub := original[1:3]         // [2, 3], len=2, cap=4

// append vào sub - có thể ghi đè original!
sub = append(sub, 999)
fmt.Println(original)        // [1, 2, 3, 999, 5] ← original bị modify!
fmt.Println(sub)             // [2, 3, 999]

// Nếu append vượt capacity → tạo array mới → original an toàn
sub = append(sub, 888, 777, 666) // exceeds cap, new array allocated
fmt.Println(original)            // [1, 2, 3, 999, 5] (không thay đổi thêm)
```

> **📦 Node.js Comparison:**
>
> | Feature | Go Slice | JS Array |
> |---|---|---|
> | Size | Dynamic (nhưng pre-allocate tốt hơn) | Dynamic |
> | Internal structure | Pointer + Len + Cap | Object với length property |
> | Capacity pre-alloc | `make([]T, 0, 100)` | Không có (JS tự manage) |
> | Sub-slicing | Tạo view vào data gốc (shared!) | `slice()` tạo shallow copy |
> | Append | `append(s, x)` trả slice mới | `push(x)` mutate in-place |
> | Copy semantics | `copy(dst, src)` | `structuredClone()` hoặc spread |

**Khi nào pre-allocate capacity:**

```go
// ❌ Bad: append nhiều lần, nhiều lần reallocate
var results []int
for i := 0; i < 10000; i++ {
    results = append(results, i*2)
}

// ✅ Good: pre-allocate capacity
results := make([]int, 0, 10000)
for i := 0; i < 10000; i++ {
    results = append(results, i*2)
}
```

### 2.3. Map - Hash Map

Map trong Go tương đương `Object` hoặc `Map` trong JavaScript:

```go
// Tạo map
scores := map[string]int{
    "Alice": 95,
    "Bob":   87,
}

// Tạo bằng make
users := make(map[string]int)    // empty map
users["Alice"] = 95

// Truy cập
score := scores["Alice"]         // 95
missing := scores["Unknown"]     // 0 (zero value, không crash!)

// Check tồn tại - COMMA OK idiom
score, ok := scores["Alice"]     // score=95, ok=true
score2, ok2 := scores["Unknown"] // score2=0, ok2=false

if ok {
    fmt.Println("Found:", score)
}

// Delete
delete(scores, "Bob")

// Iterate
for name, score := range scores {
    fmt.Printf("%s: %d\n", name, score)
}

// Length
fmt.Println(len(scores))
```

**⚠️ Nil Map vs Empty Map:**

```go
// Nil map - CRASH nếu write!
var nilMap map[string]int
// nilMap["key"] = 1        // PANIC: assignment to entry in nil map
_ = nilMap["key"]            // OK: read nil map returns zero value
fmt.Println(len(nilMap))     // OK: returns 0
delete(nilMap, "key")        // OK: no-op

// Empty map - OK
emptyMap := make(map[string]int)
emptyMap["key"] = 1          // OK

// ⚠️ Go maps KHÔNG có guaranteed iteration order!
m := map[string]int{"a": 1, "b": 2, "c": 3}
for k, v := range m {
    // Thứ tự có thể khác nhau MỖI LẦN CHẠY!
    _ = k; _ = v
}
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript Object
> const obj = { name: "Alice", age: 30 };
> obj.email = "alice@example.com";  // add property
> delete obj.age;                    // delete property
>
> // JavaScript Map
> const map = new Map();
> map.set("name", "Alice");
> map.get("name");           // "Alice"
> map.has("name");           // true (giống ok pattern)
> map.delete("name");
> map.size;                  // length
>
> // Go Map
> m := map[string]int{"a": 1}
> m["b"] = 2                 // add
> delete(m, "a")             // delete
> _, ok := m["a"]            // check exists
> len(m)                     // length
> ```
>
> | Feature | Go Map | JS Object | JS Map |
> |---|---|---|---|
> | Key types | Comparable types | String/Symbol | Any type |
> | Order | Không guaranteed | Insertion order | Insertion order |
> | Zero value access | Returns zero value | `undefined` | `undefined` |
> | Nil/Null | Panic nếu write nil map | N/A | N/A |
> | Concurrency | **NOT safe** (dùng `sync.Map` hoặc mutex) | N/A (single-thread) | N/A |

### 2.4. Slice of Structs vs Map - Khi nào dùng gì?

```go
// Slice of structs - khi cần ordered collection hoặc iterate theo thứ tự
type User struct {
    ID   int
    Name string
}
users := []User{
    {ID: 1, Name: "Alice"},
    {ID: 2, Name: "Bob"},
}

// Map - khi cần lookup nhanh theo key
userMap := map[int]User{
    1: {ID: 1, Name: "Alice"},
    2: {ID: 2, Name: "Bob"},
}

// Lookup performance:
// Map: O(1) average
// Slice linear search: O(n)
// Slice binary search (sorted): O(log n)
```

---

## 3. Control Flow

### 3.1. `for` - Loop duy nhất trong Go

Go **chỉ có `for`**, không có `while`, `do-while`, `for...in`, `for...of`:

```go
// C-style for (giống JavaScript)
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// While-style (không có init và post statement)
n := 1
for n < 100 {
    n *= 2
}

// Infinite loop
for {
    // break hoặc return để thoát
    break
}

// For với init statement (scoped)
for i := 0; i < 5; i++ {
    // i chỉ tồn tại trong scope này
}
// fmt.Println(i)  // ❌ Lỗi: i undefined

// range - iterate qua collections
nums := []int{10, 20, 30}

for i, v := range nums {
    fmt.Printf("index=%d, value=%d\n", i, v)
}

// Chỉ cần index
for i := range nums {
    fmt.Println(i) // 0, 1, 2
}

// Chỉ cần value (dùng _)
for _, v := range nums {
    fmt.Println(v) // 10, 20, 30
}

// range qua string - iterate RUNES (Unicode code points), không phải bytes!
for i, ch := range "Hello, 世界" {
    fmt.Printf("index=%d, char=%c\n", i, ch)
}
// index=0, char=H
// index=1, char=e
// ...
// index=7, char=世  (UTF-8: 3 bytes, index jumped from 7)
// index=10, char=界

// range qua map
m := map[string]int{"a": 1, "b": 2}
for k, v := range m {
    fmt.Printf("%s: %d\n", k, v)
}

// range qua channel
ch := make(chan int, 3)
ch <- 1; ch <- 2; ch <- 3
close(ch)
for v := range ch {
    fmt.Println(v) // 1, 2, 3
}
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript có NHIỀU loop
> for (let i = 0; i < 10; i++) {}
> while (condition) {}
> do {} while (condition);
> for (const key in obj) {}        // for...in
> for (const item of array) {}     // for...of
> array.forEach((item, i) => {})   // callback
> array.map(x => x * 2)           // transform
> array.filter(x => x > 5)        // filter
> array.reduce((acc, x) => acc + x, 0) // reduce
>
> // Go CHỈ CÓ for + range
> // Không có forEach, map, filter, reduce trên collections
> // (Go 1.21+ có slices package với các function này)
> ```

### 3.2. `if` với Init Statement

Go cho phép **khai báo biến ngay trong `if`**:

```go
// Nếu err chỉ dùng trong if-else block:
if err := doSomething(); err != nil {
    fmt.Println("Error:", err)
    return err
}
// err không tồn tại ở đây

// So sánh với pattern thường dùng:
err := doSomething()
if err != nil {
    fmt.Println("Error:", err)
    return err
}

// if-else chain
if score >= 90 {
    fmt.Println("A")
} else if score >= 80 {
    fmt.Println("B")
} else {
    fmt.Println("C")
}

// ⚠️ Opening brace PHẢI cùng dòng với if
if condition {
    // ...
}                // ✅
// ❌ if condition
// {
// }             // compile error!
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript không có init statement trong if
> // Phải khai báo trước
> const result = doSomething();
> if (result !== null) {
>     console.log(result);
> }
>
> // Go pattern gọn hơn:
> if result := doSomething(); result != nil {
>     fmt.Println(result)
> }
> ```

### 3.3. `switch` - Mạnh hơn JavaScript

Go switch **không fallthrough mặc định** (trái ngược JS):

```go
// Basic switch - KHÔNG fallthrough
switch day {
case "Monday":
    fmt.Println("Start of week")
    // Không fallthrough! Không cần break
case "Friday":
    fmt.Println("TGIF")
default:
    fmt.Println("Regular day")
}

// Multiple values trong 1 case
switch day {
case "Saturday", "Sunday":
    fmt.Println("Weekend")
default:
    fmt.Println("Weekday")
}

// Switch không cần expression (giống if-else chain)
score := 85
switch {
case score >= 90:
    fmt.Println("A")
case score >= 80:
    fmt.Println("B")
default:
    fmt.Println("C")
}

// Type switch (rất hữu ích cho interfaces)
func describe(i interface{}) {
    switch v := i.(type) {
    case int:
        fmt.Printf("Integer: %d\n", v)
    case string:
        fmt.Printf("String: %s\n", v)
    case bool:
        fmt.Printf("Boolean: %v\n", v)
    default:
        fmt.Printf("Unknown type: %T\n", v)
    }
}

// Fallthrough (explicit) - hiếm khi dùng
switch n := 2; n {
case 1:
    fmt.Println("One")
case 2:
    fmt.Println("Two")
    fallthrough              // BẮT BUỘC vào case tiếp theo
case 3:
    fmt.Println("Three")    // Sẽ in "Three" vì fallthrough
default:
    fmt.Println("Other")
}
// Output:
// Two
// Three
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript switch CÓ fallthrough mặc định!
> switch (day) {
>     case "Monday":
>         console.log("Start of week");
>         // ⚠️ Falls through to next case!
>     case "Tuesday":
>         console.log("Also Tuesday");
>         break;  // Phải nhớ break!
> }
>
> // Go switch - clean hơn nhiều
> switch day {
> case "Monday":
>     fmt.Println("Start of week")
>     // Không fallthrough, không cần break
> case "Tuesday":
>     fmt.Println("Tuesday")
> }
> ```

### 3.4. Labeled Break/Continue, Goto

```go
// Labeled break - thoát outer loop từ inner loop
outer:
    for i := 0; i < 5; i++ {
        for j := 0; j < 5; j++ {
            if j == 3 {
                break outer  // thoát cả outer loop
            }
            fmt.Printf("(%d, %d) ", i, j)
        }
    }

// Labeled continue
outer:
    for i := 0; i < 5; i++ {
        for j := 0; j < 5; j++ {
            if j == 3 {
                continue outer  // skip sang iteration tiếp của outer
            }
            fmt.Printf("(%d, %d) ", i, j)
        }
    }

// goto - tồn tại nhưng RẤT ÍT dùng
// Chỉ dùng trong generated code hoặc very specific patterns
goto end
fmt.Println("This won't print")
end:
fmt.Println("Jumped here")
```

> **📦 Node.js Comparison:**
>
> JavaScript có labeled break/continue nhưng rất ít dùng. Go cũng vậy.
> Go KHÔNG có `do...while`, `for...in`, `for...of` - chỉ có `for`.

---

## 4. Functions - Deep Dive

### 4.1. Multiple Return Values - Paradigm Shift

Đây là **một trong những feature quan trọng nhất** của Go:

```go
// Single return (giống JS)
func add(a, b int) int {
    return a + b
}

// Multiple returns - CỰC KỲ PHỔ BIẾN trong Go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Go convention: error LUÔN là return cuối
result, err := divide(10, 3)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result) // 3.333...

// Ignore return value với _
_, err = divide(10, 3) // chỉ lấy error
_ = err

// Return nhiều values
func swap(a, b string) (string, string) {
    return b, a
}
x, y := swap("hello", "world") // x="world", y="hello"
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript: return object destructuring
> function divide(a, b) {
>     if (b === 0) return { result: null, error: "division by zero" };
>     return { result: a / b, error: null };
> }
> const { result, error } = divide(10, 3);
> if (error) throw new Error(error);
>
> // JavaScript: callback pattern (old style)
> fs.readFile(path, (err, data) => {
>     if (err) throw err;
>     console.log(data);
> });
>
> // Go pattern - much cleaner
> data, err := os.ReadFile(path)
> if err != nil {
>     log.Fatal(err)
> }
> fmt.Println(data)
> ```

### 4.2. Named Return Values

```go
// Named returns - ít dùng nhưng cần biết
func divide(a, b float64) (result float64, err error) {
    if b == 0 {
        err = errors.New("division by zero")
        return // naked return: trả về result (0) và err
    }
    result = a / b
    return // naked return: trả về result và nil
}

// Có thể dùng naked return trong short functions
// NHƯNG trong functions dài, explicit return rõ ràng hơn

// Named returns chỉ là CỬA SỔ DOCUMENTATION
// Bạn vẫn có thể return explicit values
func example() (x int, y int) {
    return 10, 20  // OK, ignored x, y assignments
}
```

### 4.3. Variadic Functions

```go
// Variadic - nhận nhiều arguments (giống rest params trong JS)
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

sum(1, 2, 3)          // 6
sum(1, 2, 3, 4, 5)    // 15

// Pass slice vào variadic function
nums := []int{1, 2, 3}
sum(nums...)           // unpack slice

// Variadic với các tham số khác
func logMessage(level string, messages ...string) {
    for _, msg := range messages {
        fmt.Printf("[%s] %s\n", level, msg)
    }
}
logMessage("INFO", "Server started", "Listening on :8080")
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript rest parameters
> function sum(...nums) {
>     return nums.reduce((a, b) => a + b, 0);
> }
>
> // Go variadic
> func sum(nums ...int) int {
>     total := 0
>     for _, n := range nums {
>         total += n
>     }
>     return total
> }
> ```

### 4.4. Functions as First-Class Citizens

```go
// Function type
type MathFunc func(int, int) int

// Function as variable
var add = func(a, b int) int {
    return a + b
}

// Function as parameter
func apply(fn MathFunc, a, b int) int {
    return fn(a, b)
}
result := apply(add, 5, 3) // 8

// Anonymous function (lambda)
multiply := func(a, b int) int {
    return a * b
}
fmt.Println(multiply(4, 5)) // 20

// Immediately invoked function
func() {
    fmt.Println("IIFE!")
}()

// Function returning function (factory pattern)
func makeGreeter(prefix string) func(string) string {
    return func(name string) string {
        return prefix + ", " + name + "!"
    }
}
hello := makeGreeter("Hello")
hi := makeGreeter("Hi")
fmt.Println(hello("Alice")) // "Hello, Alice!"
fmt.Println(hi("Bob"))      // "Hi, Bob!"
```

### 4.5. Closures

Go closures hoạt động tương tự JavaScript nhưng có **scoping khác**:

```go
// Closure capture variable
func counter() func() int {
    count := 0
    return func() int {
        count++        // capture 'count'
        return count
    }
}

c := counter()
fmt.Println(c()) // 1
fmt.Println(c()) // 2
fmt.Println(c()) // 3

// ⚠️ Gotcha: loop variable capture (TRƯỚC Go 1.22)
// Go 1.22+: loop variables là per-iteration (giống JavaScript let)
funcs := make([]func(), 5)
for i := 0; i < 5; i++ {
    funcs[i] = func() {
        fmt.Println(i) // Go 1.22+: in 0,1,2,3,4
                        // Go < 1.22: in 5,5,5,5,5 (bug!)
    }
}
for _, f := range funcs {
    f()
}
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript closures - quen thuộc với JS devs
> function counter() {
>     let count = 0;
>     return function() {
>         count++;
>         return count;
>     };
> }
>
> // JavaScript loop capture với let (block-scoped)
> const funcs = [];
> for (let i = 0; i < 5; i++) {
>     funcs.push(() => console.log(i)); // 0,1,2,3,4
> }
>
> // Go >= 1.22 tương tự với let
> // Go < 1.22: loop vars là per-loop, không per-iteration
> ```

### 4.6. Defer - Cleanup Pattern

`defer` schedule một function chạy **khi surrounding function return**:

```go
func readFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()  // Chạy khi function return

    // ... process file ...
    // file.Close() tự động gọi dù function return ở đâu
    return nil
}

// Multiple defers - LIFO order (stack)
func example() {
    defer fmt.Println("first defer")
    defer fmt.Println("second defer")
    defer fmt.Println("third defer")
    fmt.Println("function body")
}
// Output:
// function body
// third defer
// second defer
// first defer

// Defer với anonymous function
func process() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered: %v", r)
        }
    }()
    // ... risky code ...
    return nil
}

// Defer tính toán arguments NGAY, nhưng chạy SAU
func example2() {
    x := 10
    defer fmt.Println(x)  // x được capture là 10
    x = 20
    fmt.Println(x)        // 20
}
// Output:
// 20
// 10  (defer in ra giá trị capture: 10)
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript: try/finally
> let file;
> try {
>     file = await fs.promises.open(path);
>     // ... process file ...
> } finally {
>     if (file) await file.close();
> }
>
> // Go: defer (cleaner!)
> file, err := os.Open(path)
> if err != nil { return err }
> defer file.Close()
> // ... process file ...
> ```

### 4.7. Panic & Recover

```go
// Panic - program crash (chỉ dùng khi KHÔNG THỂ recover)
func mightPanic() {
    panic("something went terribly wrong")
}

// Recover - bắt panic trong deferred function
func safeDivide(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered from panic: %v", r)
        }
    }()

    return a / b, nil  // panics if b == 0
}

// ⚠️ Panic KHÔNG PHẢI error handling!
// Error handling: return error
// Panic: unrecoverable situation (programming error)
// recover: "last resort" defense (middleware, test frameworks)

// Go convention:
// - Errors: return (error) → caller handle
// - Panics: programming bugs → let it crash, fix the code
// - Recover: middleware, test harness, top-level handlers
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript: throw/catch
> try {
>     riskyOperation();
> } catch (error) {
>     console.error(error);
> }
>
> // Go: errors return, panics are rare
> result, err := riskyOperation()
> if err != nil {
>     return err  // "errors are values"
> }
>
> // Go: panic/recover (like last resort try/catch)
> defer func() {
>     if r := recover(); r != nil {
>         log.Printf("recovered: %v", r)
>     }
> }()
> ```

### 4.8. `init()` Function

```go
// init() chạy TRƯỚC main()
// Mỗi package có thể có nhiều init()
// Không được gọi trực tiếp

package main

import "fmt"

var config map[string]string

func init() {
    // Package-level initialization
    config = make(map[string]string)
    config["env"] = "production"
    fmt.Println("init() called")
}

func main() {
    fmt.Println("main() called")
    fmt.Println(config["env"])
}
// Output:
// init() called
// main() called
// production

// init() trong imported packages chạy TRƯỚC
// Import order determines init order
```

---

## 5. Pointers - Dành Cho JavaScript Developers

### 5.1. Pointer là gì?

Pointer là biến **chứa địa chỉ bộ nhớ** của biến khác:

```go
x := 42       // x chứa giá trị 42
p := &x       // p chứa ĐỊA CHỈ của x (& là "address-of" operator)
fmt.Println(p)  // 0xc0000b2000 (địa chỉ bộ nhớ)
fmt.Println(*p) // 42 (* là "dereference" - lấy giá trị tại địa chỉ)

*p = 100       // Thay đổi giá trị TẠI ĐỊA CHỈ mà p trỏ đến
fmt.Println(x) // 100 (x bị thay đổi!)
```

**Visual:**

```
x := 42
p := &x

Stack memory:
┌─────────────┐
│ x: 42       │ ← address: 0x1000
├─────────────┤
│ p: 0x1000   │ ← p trỏ đến x
└─────────────┘
      │
      └────→ x (address 0x1000, value 42)
```

### 5.2. `&` và `*` Operators

```go
// & (address-of) - lấy địa chỉ của biến
x := 42
p := &x          // p là *int, giá trị là địa chỉ của x

// * (dereference) - lấy giá trị tại địa chỉ
val := *p        // val = 42
*p = 100         // thay đổi giá trị tại địa chỉ
fmt.Println(x)   // 100

// Zero value của pointer là nil
var p *int       // p = nil
fmt.Println(p)   // <nil>
// fmt.Println(*p) // PANIC: nil pointer dereference
```

### 5.3. Tại sao Go dùng Pointers?

```go
// 1. Tránh copy large structs
type LargeStruct struct {
    Data [1024 * 1024]byte  // 1MB data
}

func processBad(ls LargeStruct) {     // COPY 1MB mỗi lần gọi!
    // ...
}

func processGood(ls *LargeStruct) {   // chỉ copy 8 bytes (pointer)
    // ...
}

// 2. Mutation - cho phép function thay đổi caller's data
func incrementBad(x int) {
    x++           // chỉ thay đổi local copy
}

func incrementGood(x *int) {
    *x++          // thay đổi giá trị gốc
}

// 3. Optional/nullable semantics
type User struct {
    Name  string
    Email *string   // có thể nil (optional)
}

user := User{Name: "Alice", Email: nil}       // Email không được set
user2 := User{Name: "Bob", Email: &[]string{"bob@example.com"}[0]}
```

### 5.4. Khi nào dùng Pointer vs Value

```go
// Dùng POINTER khi:
// 1. Cần mutate (thay đổi) struct
// 2. Struct lớn (tránh copy overhead)
// 3. Cần represent "optional" (nil)
// 4. Implement interface với pointer receiver

// Dùng VALUE khi:
// 1. Struct nhỏ (int, bool, small struct)
// 2. Immutable data
// 3. Maps, slices (đã là reference types)
// 4. Concurrent access (value an toàn hơn)

// Example decision:
type Point struct { X, Y int }           // Dùng VALUE (nhỏ, immutable)
type User struct { ID int; Name string }  // Dùng POINTER (có thể mutate)
type Config struct { ... }                // Dùng POINTER (lớn, mutable)
```

### 5.5. Không có Pointer Arithmetic

```go
// Go KHÔNG cho phép pointer arithmetic (khác C/C++)
p := &x
// p++      // ❌ compile error
// p + 1    // ❌ compile error
// *(p + 1) // ❌ compile error

// Go KHÔNG có manual memory management
// Không có malloc/free, new/delete
// Garbage collector lo hết
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript: KHÔNG CÓ POINTERS!
> let x = 42;
> let y = x;   // y = 42 (copy value)
> y = 100;
> console.log(x); // 42 (không bị thay đổi)
>
> // Objects/Arrays là reference
> let obj = { name: "Alice" };
> let ref = obj;  // ref TRỎ ĐẾN CÙNG object
> ref.name = "Bob";
> console.log(obj.name); // "Bob" (bị thay đổi!)
>
> // Go pointers explicit hơn
> x := 42
> p := &x       // explicit: p là pointer
> *p = 100      // explicit: dereference để modify
> ```
>
> | Feature | Go | JavaScript |
> |---|---|---|
> | Primitive types | Value type | Value type |
> | Objects/Structs | Value type (copy) | Reference type |
> | Mutation | Explicit pointers | Implicit references |
> | Null pointers | `nil` (typed) | `null`, `undefined` |
> | Memory management | GC (nhưng explicit pointers) | GC (no pointers) |

### 5.6. Pointer Receiver vs Value Receiver

```go
type User struct {
    ID   int
    Name string
}

// Value receiver - hoạt động trên BẢN SAO
func (u User) GetName() string {
    return u.Name
}

// Pointer receiver - hoạt động trên GỐC
func (u *User) SetName(name string) {
    u.Name = name  // thay đổi original
}

// Usage:
user := User{ID: 1, Name: "Alice"}

name := user.GetName()    // OK: value receiver
user.SetName("Bob")       // OK: pointer receiver (Go tự lấy &user)

// Quy tắc:
// - Dùng pointer receiver khi cần mutate
// - Dùng pointer receiver khi struct lớn
// - Dùng value receiver khi immutable
// - Nhất quán: nếu 1 method dùng pointer receiver, tất cả nên dùng
```

---

## 6. Structs vs Objects/Classes

### 6.1. Definition & Initialization

```go
// Struct definition (giống class definition)
type User struct {
    ID        int
    Name      string
    Email     string
    IsActive  bool
}

// Initialization styles:
u1 := User{ID: 1, Name: "Alice", Email: "alice@example.com", IsActive: true}  // named fields
u2 := User{1, "Bob", "bob@example.com", false}                                // positional (ít dùng)
u3 := User{Name: "Charlie"}                                      // partial: other fields zero value
u4 := User{}                                                       // zero value: all fields zero

// Pointer to struct
u5 := &User{Name: "Dave"}  // returns *User

// Access fields
fmt.Println(u1.Name)
u1.Name = "Alice Updated"

// Nil pointer check
var u6 *User              // u6 = nil
if u6 != nil {
    fmt.Println(u6.Name)
}
```

### 6.2. Anonymous Structs

```go
// Anonymous struct - dùng 1 lần
point := struct {
    X, Y int
}{10, 20}
fmt.Println(point.X, point.Y)

// Hữu ích cho test data hoặc JSON responses
response := struct {
    Status  string
    Message string
}{
    Status:  "success",
    Message: "User created",
}
data, _ := json.Marshal(response)
fmt.Println(string(data))
```

### 6.3. Embedded Structs (Composition)

Go không có inheritance, dùng **composition** qua embedding:

```go
// Base struct
type Address struct {
    Street  string
    City    string
    Country string
}

func (a Address) FullAddress() string {
    return fmt.Sprintf("%s, %s, %s", a.Street, a.City, a.Country)
}

// Embedded struct
type User struct {
    Name    string
    Email   string
    Address        // embedded (không field name)
}

// Usage:
user := User{
    Name:  "Alice",
    Email: "alice@example.com",
    Address: Address{
        Street:  "123 Main St",
        City:    "Hanoi",
        Country: "Vietnam",
    },
}

// Truy cập embedded fields trực tiếp
fmt.Println(user.City)        // "Hanoi" (promoted field)
fmt.Println(user.Address.City) // "Hanoi" (explicit)

// Embedded methods được promote
fmt.Println(user.FullAddress()) // "123 Main St, Hanoi, Vietnam"

// Method override
func (u User) FullAddress() string {
    return fmt.Sprintf("%s <%s>: %s", u.Name, u.Email, u.Address.FullAddress())
}
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript: inheritance (class-based)
> class Animal {
>     constructor(name) { this.name = name; }
>     speak() { return `${this.name} speaks`; }
> }
> class Dog extends Animal {
>     speak() { return `${this.name} barks`; }
> }
>
> // Go: composition (embedding)
> type Animal struct { Name string }
> type Dog struct {
>     Animal  // embedded
>     Breed string
> }
> ```
>
> | Feature | Go | JavaScript |
> |---|---|---|
> | Inheritance | ❌ Không có | ✅ class extends |
> | Composition | ✅ Struct embedding | ✅ Mixins, composition |
> | Method override | ✅ Same method name | ✅ Override in subclass |
> | Interface impl | Implicit (duck typing) | Explicit (implements) |

### 6.4. Struct Tags

Tags là metadata cho structs, dùng rất nhiều trong encoding/ORM:

```go
type User struct {
    ID        int    `json:"id" db:"user_id" validate:"required"`
    Name      string `json:"name" db:"user_name" validate:"required,min=2"`
    Email     string `json:"email" db:"email" validate:"required,email"`
    Password  string `json:"-" db:"password"`               // "-" = skip trong JSON
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// JSON tags
data, _ := json.Marshal(user)
// {"id":1,"name":"Alice","email":"alice@example.com","created_at":"..."}

// JSON unmarshal
var u User
json.Unmarshal(data, &u)

// Common tags:
// json:"field_name"       - JSON encoding/decoding
// db:"column_name"        - Database column mapping
// validate:"required"     - Validation rules
// xml:"field_name"        - XML encoding
// yaml:"field_name"       - YAML encoding
// bson:"field_name"       - MongoDB
// gorm:"column:user_name" - GORM ORM

// Custom tag parsing với reflect
import "reflect"

t := reflect.TypeOf(User{})
field, _ := t.FieldByName("Name")
fmt.Println(field.Tag.Get("json"))    // "name"
fmt.Println(field.Tag.Get("validate")) // "required,min=2"
```

> **📦 Node.js Comparison:**
>
> ```javascript
> // JavaScript: decorators (Stage 3 proposal) hoặc class-validator
> class User {
>     @IsString() @MinLength(2)
>     name: string;
>
>     @IsEmail()
>     email: string;
>
>     @Exclude()
>     password: string;
> }
>
> // Go tags - compile-time metadata, rất nhẹ
> type User struct {
>     Name     string `json:"name" validate:"required,min=2"`
>     Email    string `json:"email" validate:"required,email"`
>     Password string `json:"-"`
> }
> ```

### 6.5. Methods trên Structs

```go
type Rectangle struct {
    Width  float64
    Height float64
}

// Value receiver - không thay đổi original
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
    return 2 * (r.Width + r.Height)
}

// Pointer receiver - có thể thay đổi
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}

// Method với multiple return values
func (r Rectangle) Dimensions() (float64, float64) {
    return r.Width, r.Height
}

// Usage:
rect := Rectangle{Width: 10, Height: 5}
fmt.Println(rect.Area())       // 50
rect.Scale(2)
fmt.Println(rect.Area())       // 200

// Go tự động lấy address khi gọi pointer receiver method
rect.Area()    // Go tự convert: (&rect).Area() nếu cần
rect.Scale(2)  // Go tự convert: (&rect).Scale(2)
```

---

## 7. Package System

### 7.1. Package Visibility

Go dùng **capitalization** để control visibility (không có `public`/`private` keywords):

```go
package user

// Exported (public) - chữ cái đầu VIẾT HOA
type User struct {    // Exported type
    ID   int          // Exported field
    Name string       // Exported field
    email string      // Unexported field (private)
}

func NewUser(name, email string) *User {  // Exported function
    return &User{Name: name, email: email}
}

func (u *User) GetName() string {  // Exported method
    return u.Name
}

func (u *User) validate() bool {  // Unexported method
    return u.email != ""
}

// Unexported (private) - chữ cái đầu viết thường
type config struct {              // Unexported type
    dbURL string                  // Unexported field
}

func newConfig() *config {        // Unexported function
    return &config{}
}
```

### 7.2. Package Initialization Order

```
1. Import packages (theo dependency order)
2. Package-level variables initialized
3. init() functions chạy (trong mỗi package)
4. main.init() chạy
5. main.main() chạy
```

```go
// file: database/db.go
package database

import "fmt"

var DB *Database

func init() {
    fmt.Println("database init()")
    DB = connect()
}

// file: main.go
package main

import (
    "fmt"
    _ "myapp/database"  // import for side effects (init)
)

func init() {
    fmt.Println("main init()")
}

func main() {
    fmt.Println("main main()")
}
// Output:
// database init()
// main init()
// main main()
```

### 7.3. Internal Packages

```
myapp/
├── go.mod
├── main.go
├── internal/          # Chỉ package trong myapp mới import được
│   ├── auth/
│   │   └── auth.go
│   └── config/
│       └── config.go
├── pkg/               # Public packages, ai cũng import được
│   └── utils/
│       └── utils.go
└── api/
    └── handler.go
```

### 7.4. Module Path Conventions

```go
// go.mod
module github.com/username/myproject

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    go.uber.org/zap v1.26.0
)

// Import paths
import (
    "github.com/username/myproject/internal/auth"
    "github.com/username/myproject/pkg/utils"
    "github.com/gin-gonic/gin"
)

// Naming conventions:
// - Module path: github.com/user/project
// - Package name: short, lowercase, no underscores
// - file names: snake_case.go
// - test files: xxx_test.go
```

---

## 8. Generics (Go 1.18+)

### 8.1. Type Parameters

```go
// Không có generics: phải viết lại cho mỗi kiểu
func MaxInt(a, b int) int {
    if a > b { return a }
    return b
}
func MaxFloat(a, b float64) float64 {
    if a > b { return a }
    return b
}

// Có generics: viết 1 lần, dùng mọi kiểu
func Max[T constraints.Ordered](a, b T) T {
    if a > b { return a }
    return b
}

Max(1, 2)           // T = int
Max(1.5, 2.5)       // T = float64
Max("abc", "def")   // T = string

// Type parameters cho structs
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}

// Usage:
intStack := &Stack[int]{}
intStack.Push(1)
intStack.Push(2)
val, ok := intStack.Pop() // val=2, ok=true

stringStack := &Stack[string]{}
stringStack.Push("hello")
```

### 8.2. Type Constraints

```go
import "golang.org/x/exp/constraints"

// Built-in constraints
type Number interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~float32 | ~float64
}

// ~ means "underlying type" (includes type aliases)
type Celsius float64
type Fahrenheit float64

func Convert[T ~float64](temp T) T {
    return T(float64(temp)*9/5 + 32)
}
Convert(Celsius(100))    // OK: Celsius underlying type là float64

// Custom constraint
type Printable interface {
    String() string
}

func PrintAll[T Printable](items []T) {
    for _, item := range items {
        fmt.Println(item.String())
    }
}

// Comparable constraint
func Contains[T comparable](slice []T, target T) bool {
    for _, item := range slice {
        if item == target {
            return true
        }
    }
    return false
}

// Ordered constraint (supports <, >, <=, >=)
func Min[T constraints.Ordered](a, b T) T {
    if a < b { return a }
    return b
}
```

### 8.3. Khi nào dùng vs khi nào tránh

```go
// ✅ NÊN dùng generics:
// - Data structures generic (Stack, Queue, Set)
// - Utility functions (Map, Filter, Reduce, Contains)
// - Sorting, searching algorithms
// - Repository patterns (generic CRUD)

// ❌ TRÁNH dùng generics:
// - Khi interface-based approach đủ
// - Khi chỉ dùng 1-2 kiểu cụ thể
// - Khi code trở nên khó đọc

// Ví dụ: interface-based (đơn giản hơn)
type Stringer interface {
    String() string
}

func Print(s Stringer) {
    fmt.Println(s.String())
}

// vs generic (linh hoạt hơn nhưng phức tạp hơn)
func Print[T fmt.Stringer](s T) {
    fmt.Println(s.String())
}
```

> **📦 Node.js Comparison:**
>
> ```typescript
> // TypeScript generics
> function max<T extends number | string>(a: T, b: T): T {
>     return a > b ? a : b;
> }
>
> class Stack<T> {
>     private items: T[] = [];
>     push(item: T): void { this.items.push(item); }
>     pop(): T | undefined { return this.items.pop(); }
> }
>
> // Go generics (tương tự nhưng compile-time)
> func Max[T constraints.Ordered](a, b T) T {
>     if a > b { return a }
>     return b
> }
>
> type Stack[T any] struct {
>     items []T
> }
> ```
>
> | Feature | Go Generics | TypeScript Generics |
> |---|---|---|
> | Type erasure | Compile-time (không runtime overhead) | Runtime (type info erased) |
> | Constraint | Interface-based | extends keyword |
> | Performance | Zero-cost abstraction | Runtime overhead |
> | Complexity | Đơn giản hơn | Nhiều features hơn |

---

## Quick Reference Table

| Concept | JavaScript/Node.js | Go |
|---|---|---|
| Variable declaration | `let`, `const`, `var` | `var`, `:=`, `const` |
| Type system | Dynamic (TS: static) | Static, explicit |
| Null/undefined | `null`, `undefined` | `nil` (reference types), zero values |
| Arrays | Dynamic, reference type | Fixed-size arrays, dynamic slices |
| Objects/Maps | `{}`, `new Map()` | `map[K]V`, structs |
| Loop | `for`, `while`, `do-while` | `for` only |
| Error handling | `try/catch/finally` | `return error`, `panic/recover` |
| Classes | `class` | Structs + methods |
| Inheritance | `extends` | Composition (embedding) |
| Interface | `implements` (explicit) | Implicit (duck typing) |
| Generics | TypeScript: `<T>` | `<T constraints>` |
| Null safety | Optional chaining `?.` | Explicit nil check |
| Async/await | `async/await` | Goroutines + channels |
| Package system | `npm`, `import/export` | `go mod`, packages |

---

## Tips cho Node.js Developers

1. **Embrace explicit error handling**: `if err != nil` pattern sẽ là bạn thân
2. **Nil checks everywhere**: Luôn check `nil` trước khi dùng pointers/slices/maps
3. **Value semantics by default**: Structs là value types, dùng pointers khi cần mutation
4. **Composition over inheritance**: Dùng embedding, không có `extends`
5. **`:=` is your friend**: Short declaration rất tiện, nhưng hiểu scope
6. **Zero values are your friend**: Không cần `new` hay constructors phức tạp
7. **Defer for cleanup**: Dùng `defer` thay vì `try/finally`
8. **No classes, no `this`**: Methods có receivers, không có `this` context

---

**Next:** [[03-error-handling-patterns]] - Chi tiết về error handling trong Go
