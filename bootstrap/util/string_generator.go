package util

import "math/rand"

func GenerateRandomString(min, max int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	length := rand.Intn(max-min+1) + min
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
