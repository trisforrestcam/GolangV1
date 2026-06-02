// ============================================
// LESSON 2: Data Types & Operators
// ============================================
// Chạy: go run ./less2
// Mục tiêu: Hiểu các kiểu dữ liệu trong Go

package main

import "fmt"

func main() {
	fmt.Println("=== Lesson 2: Data Types & Operators ===")

	// ===== INTEGER TYPES =====
	fmt.Println("\n--- Integer Types ---")

	var i8 int8 = 127            // -128 to 127
	var i16 int16 = 32767        // -32768 to 32767
	var i32 int32 = 2147483647   // -2^31 to 2^31-1
	var i64 int64 = 9223372036854775807
	var u8 uint8 = 255           // 0 to 255
	var u16 uint16 = 65535       // 0 to 65535

	fmt.Printf("int8: %d, int16: %d, int32: %d, int64: %d\n", i8, i16, i32, i64)
	fmt.Printf("uint8: %d, uint16: %d\n", u8, u16)

	// int - platform dependent (32 or 64 bit)
	var i int = 42
	fmt.Printf("int: %d\n", i)

	// ===== FLOAT TYPES =====
	fmt.Println("\n--- Float Types ---")

	var f32 float32 = 3.14
	var f64 float64 = 3.141592653589793

	fmt.Printf("float32: %f\n", f32)
	fmt.Printf("float64: %.15f\n", f64)

	// ===== STRING TYPES =====
	fmt.Println("\n--- String Types ---")

	s1 := "Hello, World!"       // Double quotes
	s2 := `Line 1
Line 2
Line 3`                        // Backticks = raw string (no escape)

	fmt.Println(s1)
	fmt.Println(s2)

	// String operations
	fmt.Printf("Length: %d\n", len(s1))
	fmt.Printf("First char: %c\n", s1[0])
	fmt.Printf("Slice: %s\n", s1[0:5])

	// ===== BOOLEAN TYPES =====
	fmt.Println("\n--- Boolean Types ---")

	t := true
	f := false

	fmt.Printf("true: %v, false: %v\n", t, f)
	fmt.Printf("AND: %v, OR: %v, NOT: %v\n", t && f, t || f, !t)

	// ===== ARITHMETIC OPERATORS =====
	fmt.Println("\n--- Arithmetic Operators ---")

	a, b := 10, 3
	fmt.Printf("%d + %d = %d\n", a, b, a+b)
	fmt.Printf("%d - %d = %d\n", a, b, a-b)
	fmt.Printf("%d * %d = %d\n", a, b, a*b)
	fmt.Printf("%d / %d = %d\n", a, b, a/b)   // Integer division
	fmt.Printf("%d %% %d = %d\n", a, b, a%b)  // Modulo

	// ===== COMPARISON OPERATORS =====
	fmt.Println("\n--- Comparison Operators ---")

	x, y := 5, 10
	fmt.Printf("%d == %d: %v\n", x, y, x == y)
	fmt.Printf("%d != %d: %v\n", x, y, x != y)
	fmt.Printf("%d < %d: %v\n", x, y, x < y)
	fmt.Printf("%d > %d: %v\n", x, y, x > y)
	fmt.Printf("%d <= %d: %v\n", x, y, x <= y)
	fmt.Printf("%d >= %d: %v\n", x, y, x >= y)

	// ===== LOGICAL OPERATORS =====
	fmt.Println("\n--- Logical Operators ---")

	p, q := true, false
	fmt.Printf("true && false = %v\n", p && q)
	fmt.Printf("true || false = %v\n", p || q)
	fmt.Printf("!true = %v\n", !p)

	// ===== BITWISE OPERATORS =====
	fmt.Println("\n--- Bitwise Operators ---")

	m, n := 0b1010, 0b1100  // Binary: 10, 12
	fmt.Printf("m = %b (%d), n = %b (%d)\n", m, m, n, n)
	fmt.Printf("AND: %b (%d)\n", m&n, m&n)
	fmt.Printf("OR:  %b (%d)\n", m|n, m|n)
	fmt.Printf("XOR: %b (%d)\n", m^n, m^n)
	fmt.Printf("NOT: %b (%d)\n", ^m, ^m)
	fmt.Printf("LEFT SHIFT:  %b (%d)\n", m<<1, m<<1)
	fmt.Printf("RIGHT SHIFT: %b (%d)\n", m>>1, m>>1)

	// ===== TYPE ALIASES =====
	fmt.Println("\n--- Type Aliases ---")

	type Celsius float64
	type Fahrenheit float64

	var temp Celsius = 100
	var fTemp Fahrenheit = Fahrenheit(temp*9/5 + 32)

	fmt.Printf("%.1f°C = %.1f°F\n", temp, fTemp)

	fmt.Println("\n✅ Lesson 2 hoàn thành!")
}
