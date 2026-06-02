// ============================================
// LESSON 10: Interfaces - Polymorphism trong Go
// ============================================
// Chạy: go run ./less10
// Mục tiêu: Hiểu Interfaces - implicit implementation, polymorphism

package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Println("=== Lesson 10: Interfaces ===")
	fmt.Println()

	// ===== 1. Basic Interface =====
	fmt.Println("--- 1. Basic Interface ---")
	fmt.Println("📌 Go interfaces are IMPLICIT - no 'implements' keyword")
	fmt.Println()

	var shape Shape

	shape = Circle{Radius: 5}
	fmt.Printf("Circle area: %.2f\n", shape.Area())
	fmt.Printf("Circle perimeter: %.2f\n", shape.Perimeter())

	shape = Rectangle{Width: 4, Height: 6}
	fmt.Printf("Rectangle area: %.2f\n", shape.Area())
	fmt.Printf("Rectangle perimeter: %.2f\n", shape.Perimeter())

	// ===== 2. Interface as Function Parameter =====
	fmt.Println("\n--- 2. Interface as Function Parameter ---")
	
	printShapeInfo(Circle{Radius: 10})
	printShapeInfo(Rectangle{Width: 5, Height: 8})

	// ===== 3. Multiple Interfaces =====
	fmt.Println("\n--- 3. Multiple Interfaces ---")
	
	// Circle satisfies both Shape and Describable
	c := Circle{Radius: 7}
	
	var s Shape = c
	var d Describable = c
	
	fmt.Printf("Shape area: %.2f\n", s.Area())
	fmt.Printf("Description: %s\n", d.Describe())

	// ===== 4. Empty Interface (any) =====
	fmt.Println("\n--- 4. Empty Interface (any) ---")
	fmt.Println("📌 interface{} = any - accepts ANY type")
	fmt.Println()

	printAnything(42)
	printAnything("hello")
	printAnything(true)
	printAnything([]int{1, 2, 3})
	printAnything(map[string]int{"a": 1})

	// ===== 5. Type Assertion =====
	fmt.Println("\n--- 5. Type Assertion ---")
	
	var val interface{} = "hello"
	
	// Type assertion
	str, ok := val.(string)
	if ok {
		fmt.Printf("String value: %s\n", str)
	}
	
	num, ok := val.(int)
	if ok {
		fmt.Printf("Int value: %d\n", num)
	} else {
		fmt.Println("Not an int!")
	}

	// ===== 6. Type Switch =====
	fmt.Println("\n--- 6. Type Switch ---")
	
	values := []interface{}{42, "hello", true, 3.14, nil}
	
	for _, v := range values {
		switch val := v.(type) {
		case int:
			fmt.Printf("  int: %d\n", val)
		case string:
			fmt.Printf("  string: %s\n", val)
		case bool:
			fmt.Printf("  bool: %v\n", val)
		case float64:
			fmt.Printf("  float64: %.2f\n", val)
		case nil:
			fmt.Println("  nil")
		default:
			fmt.Printf("  unknown: %v\n", val)
		}
	}

	// ===== 7. Interface Composition =====
	fmt.Println("\n--- 7. Interface Composition ---")
	
	var rw ReadWriter = &FileHandler{name: "data.txt"}
	
	data := []byte("Hello, World!")
	n, err := rw.Write(data)
	if err != nil {
		fmt.Printf("Write error: %v\n", err)
	} else {
		fmt.Printf("Wrote %d bytes\n", n)
	}
	
	readData, err := rw.Read()
	if err != nil {
		fmt.Printf("Read error: %v\n", err)
	} else {
		fmt.Printf("Read: %s\n", string(readData))
	}

	// ===== 8. Real-World Example: Payment Processor =====
	fmt.Println("\n--- 8. Real-World: Payment Processor ---")
	
	processors := []PaymentProcessor{
		&CreditCardProcessor{},
		&PayPalProcessor{},
		&CryptoProcessor{},
	}

	amount := 100.00
	for _, p := range processors {
		err := p.ProcessPayment(amount)
		if err != nil {
			fmt.Printf("  %s failed: %v\n", p.Name(), err)
		}
	}

	// ===== 9. Interface với Generics (Go 1.18+) =====
	fmt.Println("\n--- 9. Interface với Generics ---")
	
	// Find trong slice
	numbers := []int{1, 2, 3, 4, 5}
	found := Find(numbers, func(n int) bool {
		return n > 3
	})
	fmt.Printf("First number > 3: %d\n", found)

	// Map transformation
	strings := Map(numbers, func(n int) string {
		return fmt.Sprintf("#%d", n)
	})
	fmt.Printf("Mapped: %v\n", strings)

	// Filter
	evens := Filter(numbers, func(n int) bool {
		return n%2 == 0
	})
	fmt.Printf("Even numbers: %v\n", evens)

	// ===== 10. Nil Interface =====
	fmt.Println("\n--- 10. Nil Interface ---")
	
	var s2 Shape
	fmt.Printf("Nil interface: %v\n", s2)  // <nil>
	
	if s2 == nil {
		fmt.Println("Interface is nil")
	}

	fmt.Println("\n✅ Lesson 10 hoàn thành!")
}

// ===== INTERFACES =====

// Basic interface
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Another interface
type Describable interface {
	Describe() string
}

// Composed interface
type ReadWriter interface {
	Read() ([]byte, error)
	Write(data []byte) (int, error)
}

// ===== TYPES =====

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

func (c Circle) Describe() string {
	return fmt.Sprintf("Circle with radius %.2f", c.Radius)
}

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// ===== FUNCTIONS =====

func printShapeInfo(s Shape) {
	fmt.Printf("  Area: %.2f, Perimeter: %.2f\n", s.Area(), s.Perimeter())
}

func printAnything(val interface{}) {
	fmt.Printf("  %T: %v\n", val, val)
}

// ===== FILE HANDLER =====

type FileHandler struct {
	name    string
	content []byte
}

func (f *FileHandler) Read() ([]byte, error) {
	return f.content, nil
}

func (f *FileHandler) Write(data []byte) (int, error) {
	f.content = data
	return len(data), nil
}

// ===== PAYMENT PROCESSORS =====

type PaymentProcessor interface {
	Name() string
	ProcessPayment(amount float64) error
}

type CreditCardProcessor struct{}

func (p *CreditCardProcessor) Name() string { return "Credit Card" }
func (p *CreditCardProcessor) ProcessPayment(amount float64) error {
	fmt.Printf("  💳 Processing $%.2f via Credit Card\n", amount)
	return nil
}

type PayPalProcessor struct{}

func (p *PayPalProcessor) Name() string { return "PayPal" }
func (p *PayPalProcessor) ProcessPayment(amount float64) error {
	fmt.Printf("  🅿️ Processing $%.2f via PayPal\n", amount)
	return nil
}

type CryptoProcessor struct{}

func (p *CryptoProcessor) Name() string { return "Crypto" }
func (p *CryptoProcessor) ProcessPayment(amount float64) error {
	fmt.Printf("  ₿ Processing $%.2f via Crypto\n", amount)
	return nil
}

// ===== GENERICS HELPERS =====

func Find[T any](slice []T, predicate func(T) bool) T {
	for _, item := range slice {
		if predicate(item) {
			return item
		}
	}
	var zero T
	return zero
}

func Map[T, U any](slice []T, transform func(T) U) []U {
	result := make([]U, len(slice))
	for i, item := range slice {
		result[i] = transform(item)
	}
	return result
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}
