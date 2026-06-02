# 🔄 Go vs NodeJS - So sánh Toàn diện cho Dev 4 năm kinh nghiệm

> **Dành cho:** NodeJS/JavaScript dev muốn học Go nhanh
> **Mục tiêu:** Hiểu Go qua lăng kính những gì bạn đã biết

---

## 🧠 1. Mental Model Shifts

### Package System

```javascript
// NodeJS - Module system
const express = require('express');           // CommonJS
import express from 'express';                // ES Modules
import { Router } from 'express';             // Named import

// package.json defines dependencies
// node_modules/ stores packages
```

```go
// Go - Package system
package main                                  // Package declaration (first line)

import (
    "fmt"                                     // Standard library
    "github.com/gofiber/fiber/v2"             // Third-party
    "vtvcab-identity-api/app/models"          // Internal package
)

// go.mod defines dependencies
// No node_modules - packages stored in GOPATH/pkg/mod
```

**Key Differences:**
- Go packages are organized by directory, not by file
- Exported names start with Uppercase (`UserAccount`)
- Unexported names start with lowercase (`userAccount`)
- No default exports - everything is explicit

---

### Static Typing vs Dynamic Typing

```javascript
// NodeJS - Dynamic typing
let user = { name: "John", age: 30 };
user = "now I'm a string";  // ✅ No error
user = 42;                  // ✅ No error

// TypeScript adds types but still compiles to JS
interface User {
    name: string;
    age: number;
}
```

```go
// Go - Static typing
var user string = "John"
user = 42  // ❌ Compile error: cannot use 42 (type int) as type string

// Structs define types
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
```

**Key Differences:**
- Go catches type errors at compile time, not runtime
- No `any` type (use `interface{}` for generic, but avoid)
- Type conversions must be explicit: `string(42)` doesn't work like JS

---

### Goroutines vs Async/Await vs Promises

```javascript
// NodeJS - Async patterns
// Promise
fetchUser(id)
    .then(user => fetchOrders(user.id))
    .then(orders => processOrders(orders))
    .catch(err => handleError(err));

// Async/Await
async function processUser(id) {
    try {
        const user = await fetchUser(id);
        const orders = await fetchOrders(user.id);
        return await processOrders(orders);
    } catch (err) {
        handleError(err);
    }
}

// Parallel execution
const [user, orders] = await Promise.all([
    fetchUser(id),
    fetchOrders(id)
]);
```

```go
// Go - Goroutines + Channels

// Goroutine (lightweight thread)
go func() {
    fetchUser(id)
}()

// Channel communication
userCh := make(chan User)
errCh := make(chan error)

go func() {
    user, err := fetchUser(id)
    if err != nil {
        errCh <- err
        return
    }
    userCh <- user
}()

select {
case user := <-userCh:
    // process user
case err := <-errCh:
    // handle error
}

// WaitGroup for parallel execution
var wg sync.WaitGroup
wg.Add(2)

go func() {
    defer wg.Done()
    fetchUser(id)
}()

go func() {
    defer wg.Done()
    fetchOrders(id)
}()

wg.Wait()
```

**Key Differences:**
- Goroutines are cheaper than OS threads (~2KB vs ~1MB)
- Channels are typed communication pipes (not events)
- No event loop - Go uses real parallelism
- `select` handles multiple channels (giống `Promise.race`)

---

### Interfaces vs Duck Typing

```javascript
// NodeJS - Duck typing (if it walks like a duck...)
class Dog {
    speak() { return "Woof!"; }
}

class Cat {
    speak() { return "Meow!"; }
}

// Both work - no interface needed
function makeSound(animal) {
    return animal.speak();
}
```

```go
// Go - Interfaces are implicit
type Speaker interface {
    Speak() string
}

type Dog struct{}
func (d Dog) Speak() string { return "Woof!" }

type Cat struct{}
func (c Cat) Speak() string { return "Meow!" }

// Both satisfy Speaker interface implicitly
func makeSound(s Speaker) string {
    return s.Speak()
}
```

**Key Differences:**
- Go interfaces are satisfied implicitly (no `implements` keyword)
- Interfaces are small (usually 1-3 methods)
- Type assertions: `dog, ok := s.(Dog)`

---

### Error Handling

```javascript
// NodeJS - Try/Catch
try {
    const user = await fetchUser(id);
    const result = await processUser(user);
    return result;
} catch (err) {
    console.error(err);
    throw new Error('Failed to process user');
}
```

```go
// Go - Error as return value
user, err := fetchUser(id)
if err != nil {
    return fmt.Errorf("failed to fetch user: %w", err)
}

result, err := processUser(user)
if err != nil {
    return fmt.Errorf("failed to process user: %w", err)
}

return result
```

**Key Differences:**
- Go errors are values, not exceptions
- Check errors immediately after function call
- `fmt.Errorf` with `%w` wraps errors (giống Error cause chain)
- `panic/recover` exists but rare (giống uncaught exceptions)

---

### Pointers vs References

```javascript
// NodeJS - Pass by reference (objects)
const user = { name: "John" };
const ref = user;        // Same object
ref.name = "Jane";
console.log(user.name);  // "Jane"

// Primitives are pass by value
let x = 10;
let y = x;
y = 20;
console.log(x);  // 10
```

```go
// Go - Explicit pointers
user := User{Name: "John"}
ref := &user      // Pointer to user
ref.Name = "Jane"
fmt.Println(user.Name)  // "Jane"

// Function with pointer receiver
func (u *User) UpdateName(name string) {
    u.Name = name  // Modifies original
}

// Function with value receiver
func (u User) GetName() string {
    return u.Name  // Works on copy
}
```

**Key Differences:**
- `&` gets address (reference), `*` dereferences
- Struct methods can use pointer or value receivers
- Pointers enable mutation, values are copies
- No pointer arithmetic (unlike C)

---

## 📝 2. Syntax Translation Table

### Variable Declaration

```javascript
// NodeJS
const name = "John";           // Constant
let count = 0;                 // Mutable
var old = "don't use";         // Function-scoped
const [a, b] = [1, 2];        // Destructuring
const { name, age } = user;   // Object destructuring
```

```go
// Go
name := "John"                 // Short declaration (type inferred)
var count int = 0              // Explicit type
var name string                // Zero value ("")
const name = "John"            // Compile-time constant
a, b := 1, 2                  // Multiple assignment

// No destructuring - use explicit fields
name := user.Name
age := user.Age
```

---

### Functions

```javascript
// NodeJS - Arrow functions, callbacks
const add = (a, b) => a + b;

const greet = (name) => {
    return `Hello, ${name}!`;
};

// Callback pattern
fetchData(url, (err, data) => {
    if (err) handleError(err);
    console.log(data);
});

// Default parameters
function createUser(name, role = "user") {
    return { name, role };
}
```

```go
// Go - Functions
func add(a, b int) int {
    return a + b
}

func greet(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}

// Multiple return values (common pattern)
func fetchData(url string) (Data, error) {
    data, err := http.Get(url)
    if err != nil {
        return Data{}, err  // Return zero value + error
    }
    return data, nil
}

// Default parameters - use functional options pattern
type UserConfig struct {
    Role string
}

func WithRole(role string) func(*UserConfig) {
    return func(c *UserConfig) { c.Role = role }
}

func createUser(name string, opts ...func(*UserConfig)) {
    config := &UserConfig{Role: "user"}  // Default
    for _, opt := range opts {
        opt(config)
    }
}
```

---

### Structs vs Objects/Classes

```javascript
// NodeJS - Objects and Classes
const user = {
    name: "John",
    age: 30,
    greet() { return `Hi, I'm ${this.name}`; }
};

class User {
    constructor(name, age) {
        this.name = name;
        this.age = age;
    }
    
    greet() {
        return `Hi, I'm ${this.name}`;
    }
}

// Inheritance
class Admin extends User {
    constructor(name, age, role) {
        super(name, age);
        this.role = role;
    }
}
```

```go
// Go - Structs
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

// Method with value receiver
func (u User) Greet() string {
    return fmt.Sprintf("Hi, I'm %s", u.Name)
}

// Method with pointer receiver (can modify)
func (u *User) SetName(name string) {
    u.Name = name
}

// No inheritance - use composition
type Admin struct {
    User                    // Embedded struct
    Role string `json:"role"`
}

// Usage
admin := Admin{
    User: User{Name: "John", Age: 30},
    Role: "superadmin",
}
admin.Greet()  // Inherited method
```

---

### Arrays/Slices vs Arrays

```javascript
// NodeJS - Arrays
const arr = [1, 2, 3];
arr.push(4);           // Add to end
arr.splice(1, 1);      // Remove at index
const sliced = arr.slice(1, 3);  // Sub-array
const mapped = arr.map(x => x * 2);
const filtered = arr.filter(x => x > 2);
```

```go
// Go - Slices (dynamic arrays)
arr := []int{1, 2, 3}
arr = append(arr, 4)                // Add to end
arr = append(arr[:1], arr[2:]...)   // Remove at index
sliced := arr[1:3]                  // Sub-slice (shares memory!)

// No built-in map/filter - use loops
mapped := make([]int, len(arr))
for i, x := range arr {
    mapped[i] = x * 2
}

filtered := []int{}
for _, x := range arr {
    if x > 2 {
        filtered = append(filtered, x)
    }
}
```

---

### Maps vs Objects/Map

```javascript
// NodeJS - Objects and Maps
const obj = { name: "John", age: 30 };
obj.email = "john@example.com";    // Add property
delete obj.age;                    // Delete property
const keys = Object.keys(obj);
const values = Object.values(obj);

// Map for non-string keys
const map = new Map();
map.set(1, "one");
map.get(1);
```

```go
// Go - Maps
m := map[string]interface{}{
    "name": "John",
    "age":  30,
}
m["email"] = "john@example.com"    // Add key
delete(m, "age")                   // Delete key

// Check if key exists
val, ok := m["name"]
if !ok {
    // key doesn't exist
}

// Iterate
for key, val := range m {
    fmt.Println(key, val)
}
```

---

### Error Handling Patterns

```javascript
// NodeJS - Error handling
class AppError extends Error {
    constructor(message, statusCode) {
        super(message);
        this.statusCode = statusCode;
    }
}

// Throw
throw new AppError("Not found", 404);

// Catch
try {
    const result = await riskyOperation();
} catch (err) {
    if (err instanceof AppError) {
        res.status(err.statusCode).json({ error: err.message });
    } else {
        res.status(500).json({ error: "Internal server error" });
    }
}
```

```go
// Go - Error handling
type AppError struct {
    Message    string
    StatusCode int
}

func (e *AppError) Error() string {
    return e.Message
}

// Return error
return &AppError{Message: "Not found", StatusCode: 404}

// Check error
result, err := riskyOperation()
if err != nil {
    var appErr *AppError
    if errors.As(err, &appErr) {
        return c.Status(appErr.StatusCode).JSON(fiber.Map{
            "error": appErr.Message,
        })
    }
    return c.Status(500).JSON(fiber.Map{
        "error": "Internal server error",
    })
}
```

---

### HTTP Handlers

```javascript
// NodeJS (Express)
app.post('/api/login', async (req, res) => {
    try {
        const { username, password } = req.body;
        const user = await userService.login(username, password);
        const token = jwt.sign({ id: user.id }, SECRET);
        res.json({ success: true, token });
    } catch (err) {
        res.status(400).json({ error: err.message });
    }
});

// Middleware
app.use(authMiddleware);
app.get('/api/profile', authMiddleware, profileController);
```

```go
// Go (Fiber)
func Login(service *UserService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        body := &models.LoginBody{}
        if err := c.BodyParser(&body); err != nil {
            return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)
        }
        
        user, err := service.Login(body.UserName, body.Password)
        if err != nil {
            return utils.HTTPFailResponse(c, fiber.StatusBadRequest, err)
        }
        
        token, err := generateToken(user)
        if err != nil {
            return utils.HTTPFailResponse(c, fiber.StatusInternalServerError, err)
        }
        
        return utils.HTTPSuccessResponse(c, fiber.StatusOK, fiber.Map{
            "token": token,
        })
    }
}

// Middleware
app.Use(authMiddleware)
app.Get("/api/profile", authMiddleware, profileController)
```

---

## 📦 3. Package System

### Dependency Management

```javascript
// NodeJS
// package.json
{
    "dependencies": {
        "express": "^4.18.2",
        "mongoose": "^7.0.0"
    },
    "devDependencies": {
        "jest": "^29.0.0"
    }
}

// Install
npm install express
npm install --save-dev jest

// Lock file
package-lock.json
```

```go
// Go
// go.mod
module vtvcab-identity-api

go 1.21

require (
    github.com/gofiber/fiber/v2 v2.36.0
    gorm.io/gorm v1.25.5
)

// Install
go get github.com/gofiber/fiber/v2
go get -u github.com/gofiber/fiber/v2@latest  // Update

// Lock file
go.sum

// Clean unused
go mod tidy
```

### Import Paths

```javascript
// NodeJS
const express = require('express');              // npm package
const utils = require('./utils');                // Local file
const { User } = require('../models/user');      // Named import
```

```go
// Go
import (
    "fmt"                                        // Standard library
    "github.com/gofiber/fiber/v2"                // Third-party
    "vtvcab-identity-api/app/models"             // Internal package
    configs "vtvcab-identity-api/settings/config" // Alias import
)
```

---

## ⚡ 4. Concurrency Comparison

### Callbacks/Promises vs Goroutines

```javascript
// NodeJS - Event Loop based
// Single-threaded, non-blocking I/O

// Sequential
const user = await fetchUser(id);
const orders = await fetchOrders(user.id);

// Parallel
const [user, orders] = await Promise.all([
    fetchUser(id),
    fetchOrders(id)
]);

// Callback hell (old pattern)
fetchUser(id, (err, user) => {
    fetchOrders(user.id, (err, orders) => {
        processOrders(orders, (err, result) => {
            // ...
        });
    });
});
```

```go
// Go - Goroutines + Channels
// Multi-threaded, true parallelism

// Sequential
user, _ := fetchUser(id)
orders, _ := fetchOrders(user.ID)

// Parallel with WaitGroup
var wg sync.WaitGroup
var user User
var orders []Order

wg.Add(2)
go func() {
    defer wg.Done()
    user, _ = fetchUser(id)
}()
go func() {
    defer wg.Done()
    orders, _ = fetchOrders(id)
}()
wg.Wait()

// Channel communication
userCh := make(chan User)
ordersCh := make(chan []Order)

go func() { userCh <- fetchUser(id) }()
go func() { ordersCh <- fetchOrders(id) }()

user := <-userCh
orders := <-ordersCh
```

### When to Use Which?

| Scenario | NodeJS | Go |
|----------|--------|-----|
| I/O bound (API calls, DB queries) | Async/await | Goroutines |
| CPU intensive | Worker threads | Goroutines (true parallel) |
| Real-time (WebSocket) | Event emitter | Channels |
| Background tasks | Bull queue | Goroutines |

---

## 🖼️ 5. Framework Comparison

### Express.js vs Fiber v2

```javascript
// Express.js
const express = require('express');
const app = express();

// Middleware
app.use(express.json());
app.use(cors());

// Routes
app.post('/api/login', async (req, res) => {
    const { username, password } = req.body;
    // ...
    res.json({ success: true });
});

// Error handling
app.use((err, req, res, next) => {
    res.status(500).json({ error: err.message });
});

app.listen(3000);
```

```go
// Fiber v2
app := fiber.New()

// Middleware
app.Use(cors.New())

// Routes
app.Post("/api/login", func(c *fiber.Ctx) error {
    body := &models.LoginBody{}
    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }
    // ...
    return c.JSON(fiber.Map{"success": true})
})

// Error handling
app.Use(func(c *fiber.Ctx) error {
    return c.Status(500).JSON(fiber.Map{"error": "Internal error"})
})

app.Listen(":3000")
```

### Key Differences

| Feature | Express | Fiber |
|---------|---------|-------|
| Request body | `req.body` | `c.BodyParser(&body)` |
| Query params | `req.query` | `c.Query("key")` |
| URL params | `req.params` | `c.Params("key")` |
| Headers | `req.headers` | `c.Get("key")` |
| Response | `res.json()` | `c.JSON()` |
| Status code | `res.status(200)` | `c.Status(200)` |
| Next middleware | `next()` | `c.Next()` |
| Error | `next(err)` | `return err` |

---

## ⚠️ 6. Common "Gotchas" cho NodeJS Devs

### 1. Unused Variables Cause Compile Errors

```go
// ❌ Won't compile
func main() {
    name := "John"  // Error: name declared but not used
}

// ✅ Fix: use it or blank identifier
func main() {
    name := "John"
    fmt.Println(name)
    
    // Or use blank identifier
    _ = someFunction()  // Discard return value
}
```

### 2. No Semicolons Needed

```go
// ❌ Unnecessary but valid
func main() {
    name := "John";
    fmt.Println(name);
}

// ✅ Preferred
func main() {
    name := "John"
    fmt.Println(name)
}
```

### 3. Exported vs Unexported (Capitalization)

```go
// Exported (public) - starts with Uppercase
type UserAccount struct { ... }
func (u *UserAccount) GetName() string { ... }

// Unexported (private) - starts with lowercase
type userAccount struct { ... }
func (u *userAccount) getName() string { ... }

// Same for package-level variables
var PublicVar = "accessible everywhere"
var privateVar = "only in this package"
```

### 4. Zero Values

```go
// Go has zero values (no undefined/null for basic types)
var s string    // ""
var i int       // 0
var f float64   // 0.0
var b bool      // false
var p *User     // nil
var sl []int    // nil
var m map[string]int  // nil

// Check for nil
if sl == nil {
    sl = []int{}
}
```

### 5. Package-level vs Function-level Scope

```go
package main

// Package-level variables
var globalDB *gorm.DB
const maxRetries = 3

func main() {
    // Function-level variables
    localVar := "only in main"
    
    // Closure can access both
    go func() {
        fmt.Println(globalDB)  // ✅
        fmt.Println(localVar)  // ✅
    }()
}
```

### 6. String Formatting

```go
// ❌ JavaScript style
console.log(`Hello, ${name}! You are ${age} years old.`);

// ✅ Go style
fmt.Printf("Hello, %s! You are %d years old.\n", name, age)

// Or using Sprintf (returns string)
msg := fmt.Sprintf("Hello, %s! You are %d years old.", name, age)

// Modern Go (1.22+)
msg := fmt.Sprintf("Hello, %s! You are %d years old.", name, age)
```

### 7. No Array Destructuring

```go
// ❌ JavaScript style
const [first, second] = array;

// ✅ Go style
first := arr[0]
second := arr[1]

// Multiple return values
name, err := getName()  // This works!
```

### 8. No Optional Chaining

```go
// ❌ JavaScript style
const city = user?.address?.city;

// ✅ Go style
var city string
if user != nil && user.Address != nil {
    city = user.Address.City
}

// Or use helper functions
city := derefString(user.Address.City)
```

---

## 🎯 Quick Reference: NodeJS → Go Mental Map

| NodeJS Concept | Go Equivalent |
|----------------|---------------|
| `npm install` | `go get` |
| `package.json` | `go.mod` |
| `require()` / `import` | `import` |
| `class` | `struct` + methods |
| `interface` (TypeScript) | `interface` (implicit) |
| `async/await` | Goroutines + Channels |
| `Promise.all()` | `sync.WaitGroup` + goroutines |
| `try/catch` | `if err != nil` |
| `null` / `undefined` | `nil` (for pointers, slices, maps) |
| `console.log()` | `fmt.Println()` / `log.Info()` |
| `express()` | `fiber.New()` |
| `req.body` | `c.BodyParser(&body)` |
| `res.json()` | `c.JSON()` |
| `module.exports` | Exported names (Uppercase) |
| `this` | Receiver (`func (u *User)`) |
| `extends` | Embedding (composition) |
| `Array.map()` | `for` loop |
| `Array.filter()` | `for` loop + `append` |
| `Object.keys()` | `for key := range map` |

---

## 📚 Tiếp theo

- [[01-Project-Structure-Analysis]] - Phân tích project structure
- [[03-Go-Fundamentals-for-NodeJS-Devs]] - Go fundamentals chi tiết
- [[04-Practical-Patterns-from-Project]] - Patterns thực tế
- [[05-Quick-Reference-CheatSheet]] - Cheat sheet tra cứu nhanh
