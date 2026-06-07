# Lesson 11: Error Handling Patterns

## 🎯 Mục tiêu
Master **Error Handling** - điểm khác biệt lớn nhất giữa Go và NodeJS. Hiểu tại sao Go dùng "error as value" và cách xử lý error như một Go expert.

## 🧠 Mental Model - Góc nhìn Expert

### Go: Error as Value
```go
result, err := doSomething()
if err != nil {
    return err
}
```
- Error là **value bình thường**, không phải exception
- **Explicit handling** - Không thể ignore (trừ khi explicit `_`)
- Buộc developer **phải handle** error tại mọi level

### NodeJS: Exception-based
```js
try {
    const result = await doSomething()
} catch (err) {
    // Handle error
}
```
- Exception có thể bắt ở bất kỳ level nào
- Dễ "swallow" error hoặc handle sai level

### Triết lý: "Errors are Values"
- Go không phân biệt "expected errors" và "unexpected errors"
- Mọi lỗi đều được xử lý giống nhau
- Điều này làm code **dễ đọc hơn** và **dễ maintain hơn**

## 📚 Kiến thức cốt lõi

### Basic Error Handling
```go
// Tạo error
err := errors.New("something went wrong")

// Wrap error
err := fmt.Errorf("failed to process: %w", originalErr)

// Check error
if err != nil {
    return err
}
```

### Error Wrapping
```go
// Wrap với context
if err != nil {
    return fmt.Errorf("failed to fetch user %d: %w", userID, err)
}

// Unwrap
if errors.Unwrap(err) == sql.ErrNoRows {
    // ...
}
```

### Sentinel Errors
```go
var ErrNotFound = errors.New("not found")
var ErrInvalidInput = errors.New("invalid input")

func findUser(id int) (*User, error) {
    if id <= 0 {
        return nil, ErrInvalidInput
    }
    // ...
    return nil, ErrNotFound
}

// Check sentinel error
if errors.Is(err, ErrNotFound) {
    // Handle not found
}
```

### Custom Error Types
```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// Check type
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Printf("Field %s: %s\n", valErr.Field, valErr.Message)
}
```

### Multiple Errors
```go
// Collect multiple validation errors
func validateUser(u *User) []error {
    var errs []error
    
    if u.Name == "" {
        errs = append(errs, errors.New("name is required"))
    }
    if u.Email == "" {
        errs = append(errs, errors.New("email is required"))
    }
    
    return errs
}
```

## ⚠️ Sai lầm NodeJS dev hay mắc

1. **Ignore errors**
```go
// ❌ Tệ nhất
result, _ := doSomething()  // Ignore error!

// ✅ Luôn handle
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

2. **Dùng panic như throw**
```go
// ❌ Sai - Panic là cho unrecoverable errors
if user == nil {
    panic("user is nil")
}

// ✅ Đúng
if user == nil {
    return errors.New("user is nil")
}
```

3. **Không wrap error**
```go
// ❌ Mất context
if err != nil {
    return err  // Không biết lỗi từ đâu!
}

// ✅ Đúng
if err != nil {
    return fmt.Errorf("failed to process order %d: %w", orderID, err)
}
```

4. **Nesting quá sâu**
```go
// ❌ Pyramid of Doom
result, err := doA()
if err != nil {
    // ...
} else {
    result2, err := doB(result)
    if err != nil {
        // ...
    } else {
        // ...
    }
}

// ✅ Early return
result, err := doA()
if err != nil {
    return err
}

result2, err := doB(result)
if err != nil {
    return err
}
```

## 🏆 Best Practices (Expert Level)

### 1. Early return (Guard clause)
```go
func processUser(u *User) error {
    if u == nil {
        return errors.New("user is nil")
    }
    if !u.Active {
        return errors.New("user is inactive")
    }
    if u.Age < 18 {
        return errors.New("user must be 18+")
    }
    
    // Main logic
    return nil
}
```

### 2. Wrap errors với context
```go
func fetchUser(ctx context.Context, id int) (*User, error) {
    user, err := db.QueryContext(ctx, "SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch user %d: %w", id, err)
    }
    return user, nil
}
```

### 3. Dùng sentinel errors cho expected conditions
```go
var ErrUserNotFound = errors.New("user not found")

func findUser(id int) (*User, error) {
    // ...
    return nil, ErrUserNotFound
}

// Check
user, err := findUser(id)
if errors.Is(err, ErrUserNotFound) {
    return nil, fiber.StatusNotFound
}
```

### 4. Custom error types cho structured data
```go
type APIError struct {
    Code    int
    Message string
    Details map[string]string
}

func (e *APIError) Error() string {
    return fmt.Sprintf("API error %d: %s", e.Code, e.Message)
}

// Usage
return &APIError{
    Code:    400,
    Message: "validation failed",
    Details: map[string]string{"email": "invalid format"},
}
```

### 5. Error handling trong HTTP handlers
```go
func handleError(c *fiber.Ctx, err error) error {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        return c.Status(apiErr.Code).JSON(apiErr)
    }
    
    if errors.Is(err, ErrUserNotFound) {
        return c.Status(404).JSON(fiber.Map{"error": "User not found"})
    }
    
    return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
}
```

## 🔗 Liên kết
- **Previous:** [Lesson 10: Interfaces, Polymorphism, Generics](../less10/README.md)
- **Next:** [Lesson 12: Capstone - Concurrent Web Scraper](../less12/README.md)
- **Docs:** [go-fundamentals/03-struct-interface-error-handling.md](../../docs/go-fundamentals/03-struct-interface-error-handling.md)

---
*"Error handling trong Go explicit hơn NodeJS - nhưng đó là feature. Bạn không thể ignore error, không thể swallow exception. Code Go dài hơn một chút nhưng reliability cao hơn rất nhiều."*
