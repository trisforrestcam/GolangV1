# 📚 docs-node-go - Học Go cho NodeJS Developer

> **Dành cho:** Dev 4 năm kinh nghiệm NodeJS muốn code Go project hiệu quả
> **Project:** vtvcab-identity-api (Fiber v2 + Postgres + Redis + MongoDB + RabbitMQ)

---

## 📖 Mục lục

| # | File | Nội dung | Thời gian đọc |
|---|------|----------|---------------|
| 1 | [[01-Project-Structure-Analysis]] | Phân tích architecture, data flow, dependencies | ~30 phút |
| 2 | [[02-Go-vs-NodeJS-Comparison]] | So sánh chi tiết Go vs NodeJS với ví dụ thực tế | ~45 phút |
| 3 | [[03-Go-Fundamentals-for-NodeJS-Devs]] | Types, functions, structs, interfaces, error handling | ~45 phút |
| 4 | [[04-Practical-Patterns-from-Project]] | Patterns thực tế: Controllers, Repository, Middleware | ~40 phút |
| 5 | [[05-Quick-Reference-CheatSheet]] | Cheat sheet tra cứu nhanh hàng ngày | Bookmark! |
| 6 | [[06-Fiber-Framework-Complete-Guide]] | Fiber v2 framework chi tiết (Express.js cho Go) | ~60 phút |
| 7 | [[07-Senior-Go-Dev-Production-Patterns]] | **Senior-level patterns**: DI, Service Layer, Security, Testing | ~90 phút |

---

## 🗺️ Lộ trình học

### Tuần 1: Foundation
- [ ] Đọc [[02-Go-vs-NodeJS-Comparison]] - Hiểu mental model shift
- [ ] Đọc [[03-Go-Fundamentals-for-NodeJS-Devs]] - Nắm core concepts

### Tuần 2: Project Context
- [ ] Đọc [[01-Project-Structure-Analysis]] - Hiểu architecture
- [ ] Bookmark [[05-Quick-Reference-CheatSheet]] - Tra cứu hàng ngày

### Tuần 3: Deep Dive
- [ ] Đọc [[04-Practical-Patterns-from-Project]] - Học từ code thật
- [ ] Đọc [[06-Fiber-Framework-Complete-Guide]] - Hiểu sâu Fiber framework

### Tuần 4: Hands-on
- [ ] Chạy `go vet ./...` để hiểu code
- [ ] Sửa một bug nhỏ trong `controllers/login.go`
- [ ] Thêm một endpoint mới trong `router/v1/api.go`

---

## 💡 Key Concepts cho NodeJS Dev

| NodeJS | Go |
|--------|-----|
| `async/await` | Goroutines + Channels |
| `try/catch` | `if err != nil` |
| `class` | `struct` + methods |
| `interface` (TS) | `interface` (implicit) |
| `express()` | `fiber.New()` |
| `npm install` | `go get` |

---

## 🔗 Links hữu ích

- [Fiber v2 Docs](https://docs.gofiber.io/) - Giống Express.js
- [Go by Example](https://gobyexample.com/) - Học qua ví dụ
- [Effective Go](https://go.dev/doc/effective_go) - Best practices
- [Go Tour](https://go.dev/tour/) - Interactive tutorial
