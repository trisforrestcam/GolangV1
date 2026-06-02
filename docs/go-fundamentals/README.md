# Learn Go - Dành Cho Senior Node.js Developer

> Series tài liệu học Go từ cơ bản đến nâng cao, được viết dành riêng cho developer có kinh nghiệm Node.js/TypeScript.

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
