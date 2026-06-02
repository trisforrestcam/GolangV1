package main

import "fmt"

type User struct {
	Name string
	Age  int
}

func main() {
	user := User{Name: "Tris", Age: 28}
	fmt.Println(user)
}
