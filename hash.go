package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash, _ := bcrypt.GenerateFromPassword([]byte("TopLearn@admin"), bcrypt.DefaultCost)
	fmt.Print(string(hash))
}
