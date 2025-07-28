package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		fmt.Println("Please set JWT_SECRET env variable")
		os.Exit(1)
	}
	claims := jwt.MapClaims{
		"sub": "user-id",
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	fmt.Println(tokenStr)
}
