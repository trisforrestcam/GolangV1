// ============================================
// LESSON 4: Structs & Methods - "Classes" trong Go
// ============================================
// Chạy: go run ./less4
// Mục tiêu: Hiểu Structs, Methods, và Composition

package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== Lesson 4: Structs & Methods ===")

	// 1. Basic Struct - Giống class trong JS/TS
	fmt.Println("\n--- 1. Basic Struct ---")
	
	user := User{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		CreatedAt: time.Now(),
	}
	fmt.Printf("User: %+v\n", user)

	// 2. Struct với Pointer
	fmt.Println("\n--- 2. Struct Pointer ---")
	
	userPtr := &User{
		ID:   2,
		Name: "Jane Doe",
	}
	fmt.Printf("User pointer: %+v\n", userPtr)
	fmt.Printf("User name: %s\n", userPtr.Name)  // Auto dereference

	// 3. Constructor Pattern - Go không có constructor
	fmt.Println("\n--- 3. Constructor Pattern ---")
	
	admin := NewUser(3, "Admin", "admin@example.com", 25)
	fmt.Printf("Admin: %+v\n", admin)

	// 4. Methods - Value Receiver (không modify)
	fmt.Println("\n--- 4. Value Receiver Methods ---")
	
	fmt.Printf("Is adult: %v\n", user.IsAdult())
	fmt.Printf("Display: %s\n", user.Display())

	// 5. Methods - Pointer Receiver (có thể modify)
	fmt.Println("\n--- 5. Pointer Receiver Methods ---")
	
	user.UpdateEmail("newemail@example.com")
	fmt.Printf("Updated email: %s\n", user.Email)

	user.IncrementAge()
	fmt.Printf("Updated age: %d\n", user.Age)

	// 6. Struct Methods với validation
	fmt.Println("\n--- 6. Validation ---")
	
	err := user.SetAge(-5)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	
	err = user.SetAge(25)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Age set to: %d\n", user.Age)
	}

	// 7. Struct Embedding (Composition over Inheritance)
	fmt.Println("\n--- 7. Struct Embedding ---")
	
	employee := Employee{
		User: User{
			ID:    100,
			Name:  "Employee One",
			Email: "emp@company.com",
			Age:   35,
		},
		Department: "Engineering",
		Salary:     100000,
	}
	
	// Access embedded struct fields
	fmt.Printf("Employee name: %s\n", employee.Name)
	fmt.Printf("Department: %s\n", employee.Department)
	
	// Access embedded struct methods
	fmt.Printf("Is adult: %v\n", employee.IsAdult())
	fmt.Printf("Display: %s\n", employee.Display())

	// 8. Interface satisfaction
	fmt.Println("\n--- 8. Interfaces ---")
	
	var printable Printable = user
	printable.Print()
	
	var printable2 Printable = employee
	printable2.Print()

	// 9. Slice of Structs
	fmt.Println("\n--- 9. Slice of Structs ---")
	
	users := []User{
		{ID: 1, Name: "Alice", Age: 25},
		{ID: 2, Name: "Bob", Age: 30},
		{ID: 3, Name: "Charlie", Age: 35},
	}
	
	for _, u := range users {
		fmt.Printf("  %s (age %d)\n", u.Name, u.Age)
	}

	// 10. Struct tags (JSON, DB, etc.)
	fmt.Println("\n--- 10. Struct Tags ---")
	
	profile := Profile{
		Name:    "John",
		Age:     30,
		Phone:   "0901234567",
		private: "hidden",  // Unexported field
	}
	fmt.Printf("Profile: %+v\n", profile)
	// JSON tags sẽ được dùng trong Lesson 12 (JSON)

	fmt.Println("\n✅ Lesson 4 hoàn thành!")
}

// ===== STRUCT DEFINITIONS =====

// Basic struct - Giống class
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" validate:"required"`
	Email     string    `json:"email" validate:"email"`
	Age       int       `json:"age" validate:"min=0,max=150"`
	CreatedAt time.Time `json:"created_at"`
}

// Constructor function - Go convention
func NewUser(id uint, name, email string, age int) *User {
	return &User{
		ID:        id,
		Name:      name,
		Email:     email,
		Age:       age,
		CreatedAt: time.Now(),
	}
}

// ===== METHODS =====

// Value Receiver - Không modify original
func (u User) IsAdult() bool {
	return u.Age >= 18
}

func (u User) Display() string {
	return fmt.Sprintf("%s <%s> (age: %d)", u.Name, u.Email, u.Age)
}

// Pointer Receiver - Có thể modify original
func (u *User) UpdateEmail(email string) {
	u.Email = email
}

func (u *User) IncrementAge() {
	u.Age++
}

func (u *User) SetAge(age int) error {
	if age < 0 {
		return fmt.Errorf("invalid age: %d (must be >= 0)", age)
	}
	u.Age = age
	return nil
}

// ===== EMBEDDING (COMPOSITION) =====

// Employee embeds User - "Inheritance" trong Go
type Employee struct {
	User                    // Embedded struct
	Department string       `json:"department"`
	Salary     float64      `json:"salary"`
}

// Employee overrides User's Display method
func (e Employee) Display() string {
	return fmt.Sprintf("%s [%s] - $%.2f", e.Name, e.Department, e.Salary)
}

// ===== INTERFACES =====

// Interface - Implicit satisfaction
type Printable interface {
	Print()
}

func (u User) Print() {
	fmt.Printf("[User] %s\n", u.Display())
}

func (e Employee) Print() {
	fmt.Printf("[Employee] %s\n", e.Display())
}

// ===== STRUCT TAGS =====

type Profile struct {
	Name    string `json:"name" db:"name" validate:"required"`
	Age     int    `json:"age" db:"age"`
	Phone   string `json:"phone" db:"phone"`
	private string // Unexported - not accessible outside package
}
