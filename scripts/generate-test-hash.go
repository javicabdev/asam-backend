package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error generating hash: %v\n", err)
		return
	}
	
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Hash: %s\n", string(hashedPassword))
	
	// Verify the hash works
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		fmt.Printf("Hash verification failed: %v\n", err)
	} else {
		fmt.Println("Hash verification successful!")
	}
}
