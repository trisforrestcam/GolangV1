// ============================================
// LESSON 9: JSON Handling
// ============================================
// Chạy: go run ./less9
// Mục tiêu: Xử lý JSON trong Go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("=== Lesson 9: JSON Handling ===")
	fmt.Println()

	// ===== 1. Struct to JSON =====
	fmt.Println("--- 1. Struct to JSON ---")
	
	type User struct {
		ID        int       `json:"id"`
		Name      string    `json:"name"`
		Email     string    `json:"email,omitempty"`  // Omitempty: bỏ qua nếu empty
		Age       int       `json:"age"`
		IsActive  bool      `json:"is_active"`
		Tags      []string  `json:"tags,omitempty"`
		CreatedAt time.Time `json:"created_at"`
		Password  string    `json:"-"`  // Bỏ qua hoàn toàn
	}

	user := User{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		IsActive:  true,
		Tags:      []string{"admin", "user"},
		CreatedAt: time.Now(),
		Password:  "secret123",  // Sẽ không xuất hiện trong JSON
	}

	// Marshal: Struct → JSON
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("JSON: %s\n", string(jsonBytes))

	// Pretty print
	prettyJSON, _ := json.MarshalIndent(user, "", "  ")
	fmt.Printf("\nPretty JSON:\n%s\n", string(prettyJSON))

	// ===== 2. JSON to Struct =====
	fmt.Println("\n--- 2. JSON to Struct ---")
	
	jsonData := `{
		"id": 2,
		"name": "Jane Doe",
		"email": "jane@example.com",
		"age": 25,
		"is_active": true,
		"tags": ["user", "moderator"],
		"created_at": "2024-01-15T10:30:00Z"
	}`

	var user2 User
	err = json.Unmarshal([]byte(jsonData), &user2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("User: %+v\n", user2)

	// ===== 3. Map to JSON =====
	fmt.Println("\n--- 3. Map to JSON ---")
	
	data := map[string]interface{}{
		"name":    "John",
		"age":     30,
		"active":  true,
		"scores":  []int{95, 87, 92},
		"address": map[string]string{
			"city":    "Ho Chi Minh",
			"country": "Vietnam",
		},
	}

	jsonMap, _ := json.Marshal(data)
	fmt.Printf("Map JSON: %s\n", string(jsonMap))

	// ===== 4. JSON to Map =====
	fmt.Println("\n--- 4. JSON to Map ---")
	
	var result map[string]interface{}
	err = json.Unmarshal(jsonMap, &result)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	// Access values
	name := result["name"].(string)
	age := result["age"].(float64)  // JSON numbers are float64
	fmt.Printf("Name: %s, Age: %d\n", name, int(age))

	// ===== 5. Slice of JSON =====
	fmt.Println("\n--- 5. Slice of JSON ---")
	
	type Product struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}

	jsonArray := `[
		{"id": 1, "name": "Laptop", "price": 999.99},
		{"id": 2, "name": "Phone", "price": 699.99},
		{"id": 3, "name": "Tablet", "price": 499.99}
	]`

	var products []Product
	json.Unmarshal([]byte(jsonArray), &products)
	
	for _, p := range products {
		fmt.Printf("  %s: $%.2f\n", p.Name, p.Price)
	}

	// ===== 6. JSON với Interface =====
	fmt.Println("\n--- 6. JSON với Interface ---")
	
	type Response struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data,omitempty"`
		Error   string      `json:"error,omitempty"`
	}

	// Success response
	successResp := Response{
		Success: true,
		Data:    user,
	}
	successJSON, _ := json.Marshal(successResp)
	fmt.Printf("Success: %s\n", string(successJSON))

	// Error response
	errorResp := Response{
		Success: false,
		Error:   "User not found",
	}
	errorJSON, _ := json.Marshal(errorResp)
	fmt.Printf("Error: %s\n", string(errorJSON))

	// ===== 7. Custom JSON Marshaling =====
	fmt.Println("\n--- 7. Custom JSON Marshaling ---")
	
	type CustomTime struct {
		time.Time
	}

	ct := CustomTime{time.Now()}
	
	// Custom marshal function
	customMarshal := func(t CustomTime) ([]byte, error) {
		return []byte(`"` + t.Format("02/01/2006 15:04:05") + `"`), nil
	}

	type Event struct {
		Name string      `json:"name"`
		Date CustomTime  `json:"date"`
	}

	event := Event{
		Name: "Conference",
		Date: ct,
	}
	
	// Manual JSON building for custom format
	eventJSON := fmt.Sprintf(`{"name":"%s","date":"%s"}`, 
		event.Name, event.Date.Format("02/01/2006 15:04:05"))
	fmt.Printf("Custom format: %s\n", eventJSON)
	
	// Show what customMarshal produces
	dateJSON, _ := customMarshal(ct)
	fmt.Printf("Custom date: %s\n", string(dateJSON))

	// ===== 8. JSON Decoder (streaming) =====
	fmt.Println("\n--- 8. JSON Decoder ---")
	
	jsonStream := `{"name": "John"}{"name": "Jane"}{"name": "Bob"}`
	
	decoder := json.NewDecoder(bytes.NewReader([]byte(jsonStream)))
	
	for decoder.More() {
		var u map[string]string
		if err := decoder.Decode(&u); err != nil {
			break
		}
		fmt.Printf("  Decoded: %s\n", u["name"])
	}

	// ===== 9. JSON Encoder (streaming) =====
	fmt.Println("\n--- 9. JSON Encoder ---")
	
	// Encode to stdout
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(user)

	fmt.Println("\n✅ Lesson 9 hoàn thành!")
}
