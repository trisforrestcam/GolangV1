// ============================================
// LESSON 11: Error Handling - Tại sao Go khác biệt
// ============================================
// Chạy: go run ./less11
// Mục tiêu: Hiểu Error Handling patterns trong Go

package main

import (
	"errors"
	"fmt"
	"strconv"
)

func main() {
	fmt.Println("=== Lesson 11: Error Handling ===")
	fmt.Println()
	fmt.Println("📌 NodeJS: try/catch (exceptions)")
	fmt.Println("📌 Go: error as return value (explicit handling)")
	fmt.Println()

	// ===== 1. Basic Error Handling =====
	fmt.Println("--- 1. Basic Error Handling ---")
	
	result, err := divide(10, 2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("10 / 2 = %.2f\n", result)
	}

	result, err = divide(10, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// ===== 2. Multiple Error Checks =====
	fmt.Println("\n--- 2. Multiple Error Checks ---")
	
	value, err := parseAndDouble("21")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Double of 21 = %d\n", value)
	}

	value, err = parseAndDouble("abc")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// ===== 3. Custom Error Types =====
	fmt.Println("\n--- 3. Custom Error Types ---")
	
	err = withdraw(100, 500)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		
		// Check specific error type
		var insufficientErr *InsufficientFundsError
		if errors.As(err, &insufficientErr) {
			fmt.Printf("  Balance: $%.2f, Attempted: $%.2f\n", 
				insufficientErr.Balance, insufficientErr.Amount)
		}
	}

	err = withdraw(100, 50)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Withdrawal successful!")
	}

	// ===== 4. Sentinel Errors =====
	fmt.Println("\n--- 4. Sentinel Errors ---")
	
	_, err = findUser(0)
	if errors.Is(err, ErrUserNotFound) {
		fmt.Println("User not found (sentinel error)")
	}

	_, err = findUser(1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("User found!")
	}

	// ===== 5. Error Wrapping =====
	fmt.Println("\n--- 5. Error Wrapping ---")
	
	err = processOrder(0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		
		// Unwrap to find root cause
		unwrapped := errors.Unwrap(err)
		if unwrapped != nil {
			fmt.Printf("Root cause: %v\n", unwrapped)
		}
		
		// Check wrapped error
		if errors.Is(err, ErrUserNotFound) {
			fmt.Println("Contains ErrUserNotFound")
		}
	}

	// ===== 6. Error Handling Pattern (Production) =====
	fmt.Println("\n--- 6. Production Error Pattern ---")
	
	err = createUser("", "invalid-email", -5)
	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
	}

	err = createUser("John", "john@example.com", 25)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("User created successfully!")
	}

	// ===== 7. Error Handling với Errors Slice =====
	fmt.Println("\n--- 7. Multiple Validation Errors ---")
	
	errs := validateUser("", "", -5)
	if len(errs) > 0 {
		fmt.Println("Validation errors:")
		for _, e := range errs {
			fmt.Printf("  - %v\n", e)
		}
	}

	// ===== 8. Panic & Recover =====
	fmt.Println("\n--- 8. Panic & Recover ---")
	
	safeOperation()
	fmt.Println("Program continues after panic recovery")

	// ===== 9. Error Handling Best Practices =====
	fmt.Println("\n--- 9. Best Practices ---")
	fmt.Println(`
1. Always check errors immediately:
   result, err := doSomething()
   if err != nil {
       return err  // or handle it
   }

2. Don't ignore errors:
   // ❌ BAD
   doSomething()  // ignoring error
   
   // ✅ GOOD
   if err := doSomething(); err != nil {
       return err
   }

3. Wrap errors with context:
   // ❌ BAD
   return err
   
   // ✅ GOOD
   return fmt.Errorf("failed to process order: %w", err)

4. Use sentinel errors for expected conditions:
   if errors.Is(err, ErrNotFound) {
       // Handle not found
   }

5. Use custom errors for structured data:
   var appErr *AppError
   if errors.As(err, &appErr) {
       // Access appErr.Code, appErr.Message
   }
`)

	fmt.Println("\n✅ Lesson 11 hoàn thành!")
}

// ===== BASIC ERROR =====

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func parseAndDouble(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parseAndDouble: %w", err)
	}
	return n * 2, nil
}

// ===== CUSTOM ERROR TYPES =====

type InsufficientFundsError struct {
	Balance  float64
	Amount   float64
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("insufficient funds: balance $%.2f, attempted $%.2f", 
		e.Balance, e.Amount)
}

func withdraw(balance, amount float64) error {
	if amount > balance {
		return &InsufficientFundsError{
			Balance: balance,
			Amount:  amount,
		}
	}
	return nil
}

// ===== SENTINEL ERRORS =====

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidEmail  = errors.New("invalid email")
	ErrInvalidAge    = errors.New("invalid age")
)

func findUser(id int) (*User, error) {
	if id == 0 {
		return nil, ErrUserNotFound
	}
	return &User{ID: id, Name: "John"}, nil
}

// ===== ERROR WRAPPING =====

func processOrder(userID int) error {
	user, err := findUser(userID)
	if err != nil {
		return fmt.Errorf("processOrder: %w", err)
	}
	_ = user
	return nil
}

// ===== VALIDATION =====

type User struct {
	ID    int
	Name  string
	Email string
	Age   int
}

func createUser(name, email string, age int) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if email == "" || !contains(email, "@") {
		return fmt.Errorf("valid email is required")
	}
	if age < 0 || age > 150 {
		return fmt.Errorf("age must be between 0 and 150")
	}
	return nil
}

func validateUser(name, email string, age int) []error {
	var errs []error
	
	if name == "" {
		errs = append(errs, fmt.Errorf("name is required"))
	}
	if email == "" || !contains(email, "@") {
		errs = append(errs, fmt.Errorf("valid email is required"))
	}
	if age < 0 || age > 150 {
		errs = append(errs, fmt.Errorf("age must be between 0 and 150"))
	}
	
	return errs
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ===== PANIC & RECOVER =====

func safeOperation() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic: %v\n", r)
		}
	}()
	
	panic("something went terribly wrong!")
}
