# Learn Go - Dành Cho Senior Node.js Developer

> Series tài liệu học Go từ cơ bản đến nâng cao, được viết dành riêng cho developer có kinh nghiệm Node.js/TypeScript.

---

## 📊 Tiến độ học tập

> Đây là bảng theo dõi **duy nhất** cho cả bộ tài liệu. Tick vào từng mốc — tiến độ được lưu trên trình duyệt và đồng bộ ở mọi trang qua nút nổi **📊 Tiến độ** ở góc phải dưới. Checklist chi tiết từng ngày vẫn nằm trong bài `07 · Fast Track Roadmap`.

### Giai đoạn 1 · Nền tảng (Ngày 1)

- [ ] Hiểu mental model Go vs Node.js — file `01`
- [ ] Nắm syntax: `var`, `:=`, `const`, slice, map, `for`, `if`, `switch` — file `02`
- [ ] Viết được `struct` + method + pointer cơ bản
- [ ] Chạy được `go run`, `go fmt` và tạo module mới

### Giai đoạn 2 · Struct, Interface, Error (Ngày 2)

- [ ] Thiết kế struct, embedding và constructor
- [ ] Hiểu interface + polymorphism trong Go
- [ ] Xử lý lỗi với `errors.Is`, `errors.As` và custom error — file `03`

### Giai đoạn 3 · HTTP API (Ngày 3–4)

- [ ] Dựng được `/health` + CRUD bằng `net/http` hoặc Fiber — file `05`
- [ ] Viết middleware logging + error handling
- [ ] Test API bằng curl/Postman
- [ ] Viết unit test table-driven + `httptest` — file `06`

### Giai đoạn 4 · Concurrency (Ngày 5)

- [ ] Dùng goroutine + `WaitGroup` — file `04`
- [ ] Dùng channel + `select`
- [ ] Dùng `context.WithTimeout` để cancel
- [ ] Chạy `go run -race .` không báo lỗi

### Giai đoạn 5 · Project Layout & Capstone (Ngày 6–7)

- [ ] Tách `cmd/`, `internal/`, viết repository interface
- [ ] Wire dependency trong `main.go` + graceful shutdown
- [ ] Hoàn thành 1 mini project kèm test, `go vet ./...`, `go build ./...`

### 🚦 Mốc sẵn sàng: Phân tích & code project

> **Đọc/phân tích project thật:** xong Giai đoạn 1 → 3 (khoảng Ngày 3).
> **Sửa code trong project thật:** xong thêm Giai đoạn 4 (khoảng Ngày 5).
> **Tự code một project hoàn chỉnh:** xong Giai đoạn 5 (khoảng Ngày 6–7).

- [ ] Xong Giai đoạn 1 → 3: đọc hiểu và tự dựng được REST API CRUD
- [ ] Đọc hiểu một repo Go lạ trong khoảng 1 buổi, không cần dịch từng dòng
- [ ] Tự viết CRUD + test table-driven mà không copy nguyên mẫu
- [ ] Giải thích được khi nào dùng goroutine / channel / context
- [ ] Chạy thành thạo `go test`, `go vet`, `go build` và đọc được lỗi
- [ ] Đã phân tích ít nhất 1 project mẫu — vtvcab-identity-api (series Node.js → Go, bài `01`)

---

## 📚 Nội Dung

| # | File | Mô tả | Lines |
|---|------|--------|-------|
| 01 | [[01-go-mental-model-for-nodejs-dev]] | Mental Model - Map Go sang Node.js | ~1800 |
| 02 | [[02-go-syntax-cheatsheet]] | Syntax toàn tập - So sánh chi tiết | ~2000 |
| 03 | [[03-struct-interface-error-handling]] | Struct, Interface, Error Handling | ~850 |
| 04 | [[04-concurrency-context]] | **Concurrency - Goroutine, Channel, Context** | ~3400 |
| 05 | [[05-http-api-project-layout]] | HTTP API, Middleware, Project Layout | ~1900 |
| 06 | [[06-tooling-testing]] | Tooling, Testing, Debugging, Profiling | ~2000 |
| 07 | [[07-fast-track-roadmap]] | 7-Day Roadmap - Học Go trong 1 tuần | ~400 |

**Tổng cộng: ~12,000+ lines** nội dung chi tiết với code examples và so sánh Node.js.

---

## 🎯 Đọc Theo Thứ Tự

```
01 (Mental Model) → 02 (Syntax) → 03 (Struct/Interface/Error)
                                    ↓
                              04 (Concurrency) ← FOCUS!
                                    ↓
                              05 (HTTP API) → 06 (Testing) → 07 (Roadmap)
```

---

## 🔥 Phần Quan Trọng Nhất

**[[04-concurrency-context]]** - Đây là phần Go THỰC SỰ mạnh hơn Node.js:

- Concurrency vs Parallelism (Node chỉ có concurrency, Go có cả hai)
- Goroutine vs Promise/async-await
- Channel vs EventEmitter
- Context vs AbortController
- Worker Pool, Pipeline, Fan-out/Fan-in patterns

---

## 💡 Dành Cho Ai?

- ✅ Node.js/TypeScript developer muốn học Go
- ✅ Backend developer muốn hiểu concurrency thực sự
- ✅ Developer muốn build high-performance services
- ✅ Người muốn deploy single binary thay vì Node runtime

---

## 🚀 Bắt Đầu Nhanh

1. Đọc [[01-go-mental-model-for-nodejs-dev]] để hiểu tổng quan
2. Làm theo [[07-fast-track-roadmap]] để học theo lộ trình
3. Focus vào [[04-concurrency-context]] - đây là game changer

---

> **Nguyên tắc:** Mỗi file đều có "Node.js Comparison" boxes để bạn map trực tiếp kiến thức Node.js sang Go.
