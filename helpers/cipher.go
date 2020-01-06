package helpers

import (
	"crypto/md5"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"time"
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
	hash := md5.New()
	hash.Write([]byte(data))
	hashedData := hex.EncodeToString(hash.Sum(nil))
	rand.Seed(time.Now().Unix())
	emojiSlice := []string{"🌵", "🔥", "🍑", "📀", "😀", "💰", "🍒", "🚒", "🌽", "🌍", "🍺", "😟", "💪", "🎤", "🎵", "🤓", "🍄", "🎩", "🎯", "🙃", "🌛", "🎨", "🎧", "😆", "🎾", "✋", "⭐"}
	n := rand.Int() % len(emojiSlice)
	return emojiSlice[n] + hashedData[len(hashedData)-3:]
}
