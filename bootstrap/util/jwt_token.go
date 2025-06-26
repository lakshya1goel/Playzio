package util

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

var secretKey = []byte("secretKey")

func GenerateToken(userID uint, exp int64) (string, error) {
	claims := jwt.MapClaims{}
	claims["type"] = "authenticated"
	claims["user_id"] = userID
	claims["exp"] = exp

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func GenerateGuestToken(name string, guestID string, exp int64) (string, error) {
	claims := jwt.MapClaims{}
	claims["type"] = "guest"
	claims["guest_id"] = guestID
	claims["name"] = name
	claims["exp"] = exp

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Println(token.Method.Alg())
			return nil, fmt.Errorf("invalid signing method")
		}

		return secretKey, nil
	})

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
