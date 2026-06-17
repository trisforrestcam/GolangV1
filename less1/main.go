// ============================================
// LESSON 1: Hello World & Basic Syntax
// ============================================
// Chạy: go run ./less1
// Mục tiêu: Hiểu cấu trúc cơ bản của Go

package main

import "fmt"

func main() {
	// 1. Hello World - Điểm bắt đầu
	fmt.Println("=== Lesson 1: Hello World & Basic Syntax ===")
	fmt.Println("Hello, World!")
	fmt.Println("Xin chào từ Go!")

	// 2. Comments
	// Đây là comment 1 dòng
	/* Đây là
	   comment nhiều dòng */

	// 3. Print functions
	fmt.Print("Không xuống dòng")
	fmt.Println(" - Có xuống dòng")
	fmt.Printf("Format: %s có %d tuổi\n", "John", 30)

	// 4. Variables - Khai báo biến
	var name string = "John" // Explicit type
	var age = 30             // Type inferred
	city := "Ho Chi Minh"    // Short declaration (most common)

	fmt.Printf("Name: %s, Age: %d, City: %s\n", name, age, city)

	// 5. Multiple variables
	var (
		x int    = 10
		y string = "hello"
		z bool   = true
	)
	fmt.Printf("x=%d, y=%s, z=%v\n", x, y, z)

	// 6. Zero values - Go không có undefined/null
	var i int
	var s string
	var b bool
	var p *int

	fmt.Printf("Zero values: int=%d, string=%q, bool=%v, pointer=%v\n", i, s, b, p)

	// 7. Constants
	const Pi = 3.14159
	const Greeting = "Hello"
	// Pi = 3.14  // ❌ Error: cannot assign to constant

	fmt.Printf("Pi: %f, Greeting: %s\n", Pi, Greeting)

	// 8. Type conversion - Phải explicit
	var numInt int = 42
	var numFloat float64 = float64(numInt) // Phải convert thủ công
	var numStr string = fmt.Sprintf("%d", numInt)

	fmt.Printf("int: %d, float: %f, string: %s\n", numInt, numFloat, numStr)

	// 9. fmt verbs - Các cách format output
	fmt.Println("\n--- Format Verbs ---")
	fmt.Printf("%%v  = %v\n", name)                         // Default format
	fmt.Printf("%%+v = %+v\n", struct{ Name string }{name}) // With field names
	fmt.Printf("%%#v = %#v\n", name)                        // Go syntax
	fmt.Printf("%%T  = %T\n", name)                         // Type
	fmt.Printf("%%s  = %s\n", name)                         // String
	fmt.Printf("%%d  = %d\n", age)                          // Integer
	fmt.Printf("%%f  = %f\n", 3.14)                         // Float
	fmt.Printf("%%t  = %t\n", true)                         // Boolean
	fmt.Printf("%%p  = %p\n", &name)                        // Pointer address

	fmt.Println("\n✅ Lesson 1 hoàn thành!")
}
