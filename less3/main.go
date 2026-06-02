// ============================================
// LESSON 3: Functions - Nền tảng của Go
// ============================================
// Chạy: go run ./less3
// Mục tiêu: Hiểu functions trong Go (multiple returns, closures, variadic)

package main

import (
	"errors"
	"fmt"
	"strings"
)

func main() {
	fmt.Println("=== Lesson 3: Functions ===")

	// 1. Basic function
	fmt.Println("\n--- 1. Basic Function ---")
	greet("John")

	// 2. Function với return value
	fmt.Println("\n--- 2. Return Value ---")
	result := add(5, 3)
	fmt.Printf("5 + 3 = %d\n", result)

	// 3. Multiple return values (RẤT PHỔ BIẾN trong Go)
	fmt.Println("\n--- 3. Multiple Return Values ---")
	quotient, remainder := divide(17, 5)
	fmt.Printf("17 / 5 = %d remainder %d\n", quotient, remainder)

	// 4. Error return pattern (QUAN TRỌNG NHẤT)
	fmt.Println("\n--- 4. Error Return Pattern ---")
	result2, err := safeDivide(10, 2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("10 / 2 = %f\n", result2)
	}

	result3, err := safeDivide(10, 0)
	if err != nil {
		fmt.Printf("10 / 0 Error: %v\n", err)
	} else {
		fmt.Printf("10 / 0 = %f\n", result3)
	}

	// 5. Named return values
	fmt.Println("\n--- 5. Named Return Values ---")
	width, height := getDimensions()
	fmt.Printf("Dimensions: %d x %d\n", width, height)

	// 6. Variadic functions (giống ...args trong JS)
	fmt.Println("\n--- 6. Variadic Functions ---")
	fmt.Println(sum(1, 2, 3, 4, 5))

	nums := []int{10, 20, 30}
	fmt.Println(sum(nums...))  // Unpack slice

	// 7. Anonymous functions (closures)
	fmt.Println("\n--- 7. Anonymous Functions ---")
	
	// Inline function
	func(name string) {
		fmt.Printf("Hello from closure, %s!\n", name)
	}("John")

	// Assign to variable
	square := func(x int) int {
		return x * x
	}
	fmt.Printf("Square of 5: %d\n", square(5))

	// 8. Closures - Capturing variables
	fmt.Println("\n--- 8. Closures ---")
	counter := makeCounter()
	fmt.Println(counter()) // 1
	fmt.Println(counter()) // 2
	fmt.Println(counter()) // 3

	// 9. Function as parameter (Higher-order function)
	fmt.Println("\n--- 9. Function as Parameter ---")
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	
	evens := filter(numbers, func(n int) bool {
		return n%2 == 0
	})
	fmt.Printf("Even numbers: %v\n", evens)

	doubled := mapInts(numbers, func(n int) int {
		return n * 2
	})
	fmt.Printf("Doubled: %v\n", doubled)

	// 10. Function returning function
	fmt.Println("\n--- 10. Function Returning Function ---")
	adder := makeAdder(10)
	fmt.Printf("Add 10 to 5: %d\n", adder(5))
	fmt.Printf("Add 10 to 20: %d\n", adder(20))

	// 11. Defer - Chạy khi function kết thúc
	fmt.Println("\n--- 11. Defer ---")
	processFile("data.txt")

	// 12. Panic và Recover
	fmt.Println("\n--- 12. Panic & Recover ---")
	safeDivide2(10, 0)
	safeDivide2(10, 2)
	fmt.Println("Program continues after panic recovery")

	fmt.Println("\n✅ Lesson 3 hoàn thành!")
}

// Basic function
func greet(name string) {
	fmt.Printf("Hello, %s!\n", name)
}

// Function với return
func add(a, b int) int {
	return a + b
}

// Multiple return values
func divide(a, b int) (int, int) {
	return a / b, a % b
}

// Error return pattern - ĐIỂM MẠNH CỦA GO
// NodeJS: throw/catch exceptions
// Go: return error as value
func safeDivide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// Named return values
func getDimensions() (width, height int) {
	width = 1920
	height = 1080
	return  // Named returns automatically returned
}

// Variadic function
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Closure - captures variable
func makeCounter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

// Higher-order function
func filter(nums []int, predicate func(int) bool) []int {
	var result []int
	for _, n := range nums {
		if predicate(n) {
			result = append(result, n)
		}
	}
	return result
}

func mapInts(nums []int, transform func(int) int) []int {
	result := make([]int, len(nums))
	for i, n := range nums {
		result[i] = transform(n)
	}
	return result
}

// Function returning function
func makeAdder(x int) func(int) int {
	return func(y int) int {
		return x + y
	}
}

// Defer - runs when function exits
func processFile(filename string) {
	fmt.Printf("Opening %s...\n", filename)
	defer fmt.Printf("Closing %s (defer)\n", filename)  // Runs last
	
	fmt.Printf("Processing %s...\n", filename)
	fmt.Printf("Writing to %s...\n", filename)
}

// Panic & Recover
func safeDivide2(a, b int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic: %v\n", r)
		}
	}()
	
	if b == 0 {
		panic("division by zero!")
	}
	fmt.Printf("%d / %d = %d\n", a, b, a/b)
}

// Helper function
func formatName(first, last string) string {
	return strings.TrimSpace(first + " " + last)
}
