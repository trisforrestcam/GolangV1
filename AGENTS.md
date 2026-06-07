# Agent Notes — GolangV1

## What this repo is
Vietnamese-language Go tutorial for NodeJS developers. 12 standalone lessons (`less1/`–`less12/`) plus docs.

## Module
- **Name:** `golangV1`
- **Go version:** 1.21
- **Only dependency:** `github.com/gofiber/fiber/v2 v2.52.5` (used only in `less8/`)

## How to run
Each lesson is an independent `package main` program:

```bash
go run ./less1   # or ./less2, ./less3, ... ./less12
go run .         # root main.go (simple demo, not primary)
```

## Structure
```
main.go          # root demo (structs)
less1/           # Hello World, basic syntax
less2/           # Data types & operators
less3/           # Functions, closures, defer
less4/           # Structs, methods, composition
less5/           # Goroutines, WaitGroup, Mutex
less6/           # Channels, select, pipeline
less7/           # Real-world concurrency patterns
less8/           # Fiber HTTP server (uses fiber/v2)
less9/           # JSON encoding/decoding
less10/          # Interfaces, polymorphism, generics
less11/          # Error handling patterns
less12/          # Capstone: concurrent web scraper
docs/            # Vietnamese docs: go-fundamentals/, nodejs-to-go/
```

## Testing
- **No tests exist.** The repo has zero `*_test.go` files. Do not assume test files or run `go test`.

## Important conventions
- Comments and output strings are in **Vietnamese** (e.g., `✅ Lesson 5 hoàn thành!`).
- Each `lessN/main.go` starts with a header comment showing the exact run command.
- Root `main.go` is **not** the main application; it is a minimal struct demo.
- `less8/main.go` spins up a local Fiber server on `:3000` with CRUD routes (`/users`, `/login`, etc.).

## Commands
```bash
go mod tidy               # after adding dependencies
go run ./lessN            # run a specific lesson
go run .                  # run root demo
go build ./lessN          # build a lesson binary
go vet ./...              # static analysis (no tests)
```
