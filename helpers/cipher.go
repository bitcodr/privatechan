package helpers

import (
	"golang.org/x/crypto/bcrypt"
	"hash/fnv"
	"strconv"
)

//HashPassword helper
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

//CheckPasswordHash handler
func CheckPasswordHash(password, hash string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return false
	}
	return true
}

func Hash(data string) string {
	hash := fnv.New32a()
	hash.Write([]byte(data))
	return strconv.FormatInt(int64(hash.Sum32()), 10)
}
