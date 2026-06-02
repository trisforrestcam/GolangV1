# 🎓 Go Learning Path - Từ NodeJS đến Go

> **Dành cho:** NodeJS dev muốn học Go từ cơ bản đến nâng cao
> **Tập trung:** Concurrency & Parallelism - điểm mạnh vượt trội của Go

---

## 📖 Danh sách Lessons

| Lesson | File | Nội dung | Thời gian |
|--------|------|----------|-----------|
| 1 | `less1` | Hello World & Basic Syntax | 5 phút |
| 2 | `less2` | Data Types & Operators | 10 phút |
| 3 | `less3` | Functions (multiple returns, closures) | 15 phút |
| 4 | `less4` | Structs & Methods ("Classes" trong Go) | 15 phút |
| 5 | `less5` | **Goroutines** - Concurrency (ĐIỂM MẠNH #1) | 20 phút |
| 6 | `less6` | **Channels** - Communication giữa Goroutines | 20 phút |
| 7 | `less7` | **Real-World Patterns** (Worker Pool, Pub/Sub, Circuit Breaker) | 25 phút |
| 8 | `less8` | HTTP Server với Fiber (Express.js cho Go) | 15 phút |
| 9 | `less9` | JSON Handling | 10 phút |
| 10 | `less10` | Interfaces - Polymorphism | 15 phút |
| 11 | `less11` | Error Handling - Tại sao Go khác biệt | 15 phút |
| 12 | `less12` | **Capstone: Concurrent Web Scraper** | 20 phút |

---

## 🚀 Cách chạy

```bash
# Setup (chỉ cần chạy 1 lần)
cd /home/tris/poc/golangV1
go mod tidy

# Chạy từng lesson
go run ./less1    # Hello World
go run ./less2    # Data Types
go run ./less3    # Functions
go run ./less4    # Structs
go run ./less5    # Goroutines ⭐
go run ./less6    # Channels ⭐
go run ./less7    # Real-World Patterns ⭐
go run ./less8    # HTTP Server (port 3000)
go run ./less9    # JSON
go run ./less10   # Interfaces
go run ./less11   # Error Handling
go run ./less12   # Capstone ⭐
```

---

## 🎯 Lộ trình học gợi ý

### Tuần 1: Foundation
- [ ] Less 1-4: Basic syntax, types, functions, structs
- [ ] Less 9: JSON handling
- [ ] Less 11: Error handling

### Tuần 2: Concurrency (QUAN TRỌNG NHẤT)
- [ ] Less 5: Goroutines - tại sao Go nhanh hơn NodeJS
- [ ] Less 6: Channels - cách goroutines giao tiếp
- [ ] Less 7: Real-world patterns

### Tuần 3: Production
- [ ] Less 8: HTTP server với Fiber
- [ ] Less 10: Interfaces & polymorphism
- [ ] Less 12: Capstone project

---

## 📌 So sánh NodeJS vs Go

### Concurrency
```javascript
// NodeJS: Single-threaded event loop
const results = await Promise.all([
    fetch('/api/1'),
    fetch('/api/2'),
    fetch('/api/3'),
]);
// ⚠️ Chạy trên 1 thread, CPU-intensive sẽ block
```

```go
// Go: Multi-threaded với Goroutines
ch1 := make(chan Result)
ch2 := make(chan Result)
ch3 := make(chan Result)

go func() { ch1 <- fetch("/api/1") }()
go func() { ch2 <- fetch("/api/2") }()
go func() { ch3 <- fetch("/api/3") }()

r1, r2, r3 := <-ch1, <-ch2, <-ch3
// ✅ Chạy trên nhiều threads, TRUE parallelism
```

### Error Handling
```javascript
// NodeJS: try/catch
try {
    const user = await db.findUser(id);
    const token = generateToken(user);
    res.json({ token });
} catch (err) {
    res.status(500).json({ error: err.message });
}
```

```go
// Go: error as return value
user, err := db.FindUser(id)
if err != nil {
    return fiber.NewError(500, "Failed to find user")
}

token, err := generateToken(user)
if err != nil {
    return fiber.NewError(500, "Failed to generate token")
}

return c.JSON(fiber.Map{"token": token})
```

---

## 🔗 Tài liệu tham khảo

- [Go Official Tour](https://go.dev/tour/)
- [Go by Example](https://gobyexample.com/)
- [Fiber Documentation](https://docs.gofiber.io/)
- [Effective Go](https://go.dev/doc/effective_go)
