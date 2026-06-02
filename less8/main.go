// ============================================
// LESSON 8: HTTP Server với Fiber (Express.js cho Go)
// ============================================
// Chạy: go run ./less8
// Mục tiêu: Tạo HTTP server với Fiber - giống Express.js

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// ===== Models =====
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ===== In-memory "database" =====
var users = []User{
	{ID: 1, Name: "John Doe", Email: "john@example.com", Age: 30},
	{ID: 2, Name: "Jane Doe", Email: "jane@example.com", Age: 25},
	{ID: 3, Name: "Bob Smith", Email: "bob@example.com", Age: 35},
}

func main() {
	fmt.Println("=== Lesson 8: HTTP Server với Fiber ===")
	fmt.Println()
	fmt.Println("📌 Fiber = Express.js cho Go")
	fmt.Println("   - API giống hệt Express")
	fmt.Println("   - Performance gấp 10x Express")
	fmt.Println()

	// ===== 1. Create App =====
	app := fiber.New(fiber.Config{
		AppName:      "Go API Demo",
		BodyLimit:    4 * 1024 * 1024, // 4MB
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	// ===== 2. Middleware (giống Express middleware) =====
	app.Use(recover.New())         // Panic recovery
	app.Use(logger.New())          // Request logging
	app.Use(requestid.New())       // Request ID
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
	}))

	// ===== 3. Routes =====
	
	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	// ===== GET Routes =====
	
	// GET /users - List all users
	// Express: app.get('/users', (req, res) => { ... })
	app.Get("/users", func(c *fiber.Ctx) error {
		// Query params: ?page=1&limit=10
		page := c.Query("page", "1")
		limit := c.Query("limit", "10")
		
		return c.JSON(fiber.Map{
			"success": true,
			"data":    users,
			"page":    page,
			"limit":   limit,
			"total":   len(users),
		})
	})

	// GET /users/:id - Get user by ID
	// Express: app.get('/users/:id', (req, res) => { ... })
	app.Get("/users/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid user ID",
			})
		}

		for _, user := range users {
			if user.ID == id {
				return c.JSON(fiber.Map{
					"success": true,
					"data":    user,
				})
			}
		}

		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "User not found",
		})
	})

	// ===== POST Routes =====
	
	// POST /users - Create new user
	// Express: app.post('/users', (req, res) => { ... })
	app.Post("/users", func(c *fiber.Ctx) error {
		user := new(User)
		
		// Parse JSON body
		if err := c.BodyParser(user); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid request body",
			})
		}

		// Validate
		if user.Name == "" || user.Email == "" {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   "Name and email are required",
			})
		}

		// Add to "database"
		user.ID = len(users) + 1
		users = append(users, *user)

		return c.Status(201).JSON(fiber.Map{
			"success": true,
			"data":    user,
		})
	})

	// POST /login - Login endpoint
	app.Post("/login", func(c *fiber.Ctx) error {
		req := new(LoginRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid request body",
			})
		}

		// Simple auth check (demo only!)
		if req.Username == "admin" && req.Password == "password" {
			return c.JSON(fiber.Map{
				"success": true,
				"data": LoginResponse{
					Token: "demo-jwt-token-12345",
					User: User{
						ID:    0,
						Name:  "Admin",
						Email: "admin@example.com",
					},
				},
			})
		}

		return c.Status(401).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid credentials",
		})
	})

	// ===== PUT Routes =====
	
	// PUT /users/:id - Update user
	app.Put("/users/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid user ID",
			})
		}

		update := new(User)
		if err := c.BodyParser(update); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid request body",
			})
		}

		for i, user := range users {
			if user.ID == id {
				users[i].Name = update.Name
				users[i].Email = update.Email
				users[i].Age = update.Age
				return c.JSON(fiber.Map{
					"success": true,
					"data":    users[i],
				})
			}
		}

		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "User not found",
		})
	})

	// ===== DELETE Routes =====
	
	// DELETE /users/:id - Delete user
	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid user ID",
			})
		}

		for i, user := range users {
			if user.ID == id {
				users = append(users[:i], users[i+1:]...)
				return c.JSON(fiber.Map{
					"success": true,
					"message": "User deleted",
				})
			}
		}

		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "User not found",
		})
	})

	// ===== Group Routes (giống Express Router) =====
	
	// api := app.Group("/api/v1")
	// api.Get("/users", getUsers)
	// api.Post("/users", createUser)

	// ===== Start Server =====
	fmt.Println("🚀 Server starting on http://localhost:3000")
	fmt.Println()
	fmt.Println("📌 Endpoints:")
	fmt.Println("   GET    /health")
	fmt.Println("   GET    /users?page=1&limit=10")
	fmt.Println("   GET    /users/:id")
	fmt.Println("   POST   /users")
	fmt.Println("   POST   /login")
	fmt.Println("   PUT    /users/:id")
	fmt.Println("   DELETE /users/:id")
	fmt.Println()
	fmt.Println("📌 Test với curl:")
	fmt.Println("   curl http://localhost:3000/health")
	fmt.Println("   curl http://localhost:3000/users")
	fmt.Println("   curl http://localhost:3000/users/1")
	fmt.Println(`   curl -X POST http://localhost:3000/users -H "Content-Type: application/json" -d '{"name":"New User","email":"new@example.com","age":28}'`)
	fmt.Println(`   curl -X POST http://localhost:3000/login -H "Content-Type: application/json" -d '{"username":"admin","password":"password"}'`)
	fmt.Println()

	log.Fatal(app.Listen(":3000"))
}
