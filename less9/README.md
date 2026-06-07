# Lesson 9: JSON Handling

## 🎯 Mục tiêu
Master **JSON encoding/decoding** trong Go - từ basic marshal/unmarshal đến streaming, custom marshaling, và performance optimization.

## 🧠 Mental Model - Góc nhìn Expert

### JSON trong Go = Reflection-based
- `encoding/json` dùng reflection để đọc struct tags
- **Performance:** Slower so với libraries khác (easyjson, jsoniter) nhưng đủ cho 90% use cases
- **Trade-off:** Standard library > Performance cho đến khi profiling chứng minh ngược lại

### Struct Tags là Metadata
```go
type User struct {
    Name     string `json:"name"`           // JSON key = "name"
    Age      int    `json:"age,omitempty"`  // Bỏ qua nếu zero value
    Password string `json:"-"`              // Hoàn toàn ignore
}
```
- Tags là **compile-time metadata** cho reflection
- `json:"-"` = ignore hoàn toàn
- `omitempty` = skip nếu zero value

### `interface{}` = `any`
```go
var data interface{} = user  // Có thể hold bất kỳ type nào
// Hoặc
data := any(user)           // Go 1.18+
```
- `interface{}` và `any` là **hoàn toàn giống nhau**
- Dùng `any` cho code rõ ràng hơn (Go 1.18+)

## 📚 Kiến thức cốt lõi

### Marshal / Unmarshal
```go
// Struct → JSON
jsonBytes, err := json.Marshal(user)

// JSON → Struct
var user User
err := json.Unmarshal(jsonBytes, &user)

// Pretty print
prettyJSON, err := json.MarshalIndent(user, "", "  ")
```

### Map ↔ JSON
```go
// Map → JSON
data := map[string]interface{}{
    "name": "John",
    "age": 30,
}
jsonBytes, _ := json.Marshal(data)

// JSON → Map
var result map[string]interface{}
json.Unmarshal(jsonBytes, &result)

// ⚠️ JSON numbers → float64 trong map!
age := result["age"].(float64)  // Không phải int!
```

### Streaming (Decoder/Encoder)
```go
// Decode từ stream (file, network)
decoder := json.NewDecoder(r)
var user User
err := decoder.Decode(&user)

// Encode ra stream
encoder := json.NewEncoder(w)
err := encoder.Encode(user)
```

### Slice of Structs
```go
jsonArray := `[{"id":1,"name":"A"},{"id":2,"name":"B"}]`
var products []Product
err := json.Unmarshal([]byte(jsonArray), &products)
```

## ⚠️ Sai lầm NodeJS dev hay mắc

1. **Quên dấu `&` trong unmarshal**
```go
// ❌ Sai - Không modify được
var user User
json.Unmarshal(data, user)   // Truyền value, không phải pointer!

// ✅ Đúng
var user User
json.Unmarshal(data, &user)  // Truyền pointer
```

2. **Nhầm lẫn number trong map**
```go
// ❌ Panic: interface conversion
var data map[string]interface{}
json.Unmarshal(jsonBytes, &data)
age := data["age"].(int)  // ❌ Lỗi! JSON numbers = float64

// ✅ Đúng
age := int(data["age"].(float64))
```

3. **Không handle error**
```go
// ❌ Tệ
json.Unmarshal(data, &user)

// ✅ Luôn check
if err := json.Unmarshal(data, &user); err != nil {
    return fmt.Errorf("failed to unmarshal user: %w", err)
}
```

4. **Dùng struct tags sai**
```go
// ❌ Sai syntax
Name string `json:name`     // Thiếu dấu nháy

// ✅ Đúng
Name string `json:"name"`
```

## 🏆 Best Practices (Expert Level)

### 1. Consistent struct tags
```go
type User struct {
    ID        int       `json:"id" db:"id" validate:"required"`
    Name      string    `json:"name" db:"name" validate:"required"`
    Email     string    `json:"email" db:"email" validate:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

### 2. Custom marshaling
```go
type CustomTime struct {
    time.Time
}

func (ct CustomTime) MarshalJSON() ([]byte, error) {
    return []byte(`"` + ct.Format("02/01/2006 15:04:05") + `"`), nil
}

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
    t, err := time.Parse(`"02/01/2006 15:04:05"`, string(data))
    if err != nil {
        return err
    }
    ct.Time = t
    return nil
}
```

### 3. Dùng `json.RawMessage` cho lazy parsing
```go
type Event struct {
    Type string          `json:"type"`
    Data json.RawMessage `json:"data"`  // Parse sau
}

var event Event
json.Unmarshal(data, &event)

switch event.Type {
case "user":
    var user User
    json.Unmarshal(event.Data, &user)
case "order":
    var order Order
    json.Unmarshal(event.Data, &order)
}
```

### 4. Streaming cho large data
```go
// Decode từ HTTP request body
decoder := json.NewDecoder(r.Body)
decoder.DisallowUnknownFields()  // Error nếu có field không xác định

var user User
if err := decoder.Decode(&user); err != nil {
    return err
}
```

### 5. Validation sau unmarshal
```go
func parseUser(data []byte) (*User, error) {
    var user User
    if err := json.Unmarshal(data, &user); err != nil {
        return nil, err
    }
    
    if user.Name == "" {
        return nil, errors.New("name is required")
    }
    if user.Age < 0 {
        return nil, errors.New("age must be positive")
    }
    
    return &user, nil
}
```

## 🔗 Liên kết
- **Previous:** [Lesson 8: Fiber HTTP Server](../less8/README.md)
- **Next:** [Lesson 10: Interfaces, Polymorphism, Generics](../less10/README.md)
- **Docs:** [go-fundamentals/03-struct-interface-error-handling.md](../../docs/go-fundamentals/03-struct-interface-error-handling.md)

---
*"JSON trong Go đơn giản nhưng powerful. Struct tags cho phép bạn control serialization hoàn toàn. Đừng dùng map[string]interface{} nếu có thể dùng struct - type safety là friend, không phải enemy."*
